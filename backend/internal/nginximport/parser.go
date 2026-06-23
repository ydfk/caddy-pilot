package nginximport

import (
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

type Site struct {
	Domains      []string
	Upstreams    []string
	UpstreamType string
	EnableHTTPS  bool
	ForceHTTPS   bool
	EnableGzip   bool
	EnableLog    bool
}

type Result struct {
	Sites    []Site
	Warnings []string
}

type directive struct {
	name     string
	args     []string
	children []directive
}

type siteGroup struct {
	site        Site
	hasProxy    bool
	hasRedirect bool
}

func Parse(payload string) (Result, error) {
	tokens, err := tokenize(payload)
	if err != nil {
		return Result{}, err
	}
	directives, next, err := parseDirectives(tokens, 0, false)
	if err != nil {
		return Result{}, err
	}
	if next != len(tokens) {
		return Result{}, fmt.Errorf("Nginx 配置存在未解析内容")
	}
	scope := importScope(directives)
	upstreams := collectUpstreamGroups(scope)
	groups := make(map[string]*siteGroup)
	order := make([]string, 0)
	warnings := make([]string, 0)
	for _, item := range scope {
		if item.name != "server" {
			continue
		}
		domains := serverNames(item)
		if len(domains) == 0 {
			warnings = append(warnings, "已跳过缺少 server_name 的 server 块")
			continue
		}
		key := strings.Join(domains, "\x00")
		group, exists := groups[key]
		if !exists {
			group = &siteGroup{site: Site{Domains: domains, UpstreamType: "http"}}
			groups[key] = group
			order = append(order, key)
		}
		applyServerBlock(group, item, upstreams, &warnings)
	}
	result := Result{Sites: make([]Site, 0, len(order)), Warnings: warnings}
	for _, key := range order {
		group := groups[key]
		if !group.hasProxy {
			result.Warnings = append(result.Warnings, fmt.Sprintf("已跳过 %s：没有可导入的 proxy_pass", strings.Join(group.site.Domains, ", ")))
			continue
		}
		group.site.ForceHTTPS = group.hasRedirect && group.site.EnableHTTPS
		result.Sites = append(result.Sites, group.site)
	}
	if len(result.Sites) == 0 {
		return result, fmt.Errorf("Nginx 配置中没有可导入的反向代理站点")
	}
	return result, nil
}

func importScope(directives []directive) []directive {
	scope := make([]directive, 0, len(directives))
	for _, item := range directives {
		if item.name == "http" {
			scope = append(scope, item.children...)
			continue
		}
		scope = append(scope, item)
	}
	return scope
}

func applyServerBlock(group *siteGroup, server directive, upstreamGroups map[string][]string, warnings *[]string) {
	if _, tlsEnabled := serverTLS(server); tlsEnabled {
		group.site.EnableHTTPS = true
	}
	if hasHTTPSRedirect(server) {
		group.hasRedirect = true
	}
	if directiveEnabled(server.children, "gzip") {
		group.site.EnableGzip = true
	}
	if hasAccessLog(server) {
		group.site.EnableLog = true
	}
	passes := findDirectives(server.children, "proxy_pass")
	if len(passes) == 0 || group.hasProxy {
		return
	}
	upstreams, upstreamType, warning, err := resolveProxyPass(passes[0], upstreamGroups)
	if err != nil {
		*warnings = append(*warnings, fmt.Sprintf("%s：%v", strings.Join(group.site.Domains, ", "), err))
		return
	}
	if warning != "" {
		*warnings = append(*warnings, fmt.Sprintf("%s：%s", strings.Join(group.site.Domains, ", "), warning))
	}
	if len(passes) > 1 {
		*warnings = append(*warnings, fmt.Sprintf("%s：检测到多个 proxy_pass，仅导入第一个", strings.Join(group.site.Domains, ", ")))
	}
	group.site.Upstreams = upstreams
	group.site.UpstreamType = upstreamType
	group.hasProxy = true
}

func collectUpstreamGroups(directives []directive) map[string][]string {
	result := make(map[string][]string)
	for _, item := range directives {
		if item.name != "upstream" || len(item.args) == 0 {
			continue
		}
		servers := findDirectives(item.children, "server")
		values := make([]string, 0, len(servers))
		for _, server := range servers {
			if len(server.args) > 0 {
				values = append(values, server.args[0])
			}
		}
		if len(values) > 0 {
			result[item.args[0]] = values
		}
	}
	return result
}

func resolveProxyPass(item directive, groups map[string][]string) ([]string, string, string, error) {
	if len(item.args) == 0 {
		return nil, "", "", fmt.Errorf("proxy_pass 缺少地址")
	}
	raw := item.args[0]
	if strings.Contains(raw, "$") {
		return nil, "", "", fmt.Errorf("暂不支持包含变量的 proxy_pass")
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" {
		return nil, "", "", fmt.Errorf("无法解析 proxy_pass %q", raw)
	}
	upstreamType := "http"
	switch parsed.Scheme {
	case "http":
	case "https":
		upstreamType = "https"
	case "grpc":
		upstreamType = "h2c"
	default:
		return nil, "", "", fmt.Errorf("暂不支持 %s 类型的 proxy_pass", parsed.Scheme)
	}
	if values := groups[parsed.Host]; len(values) > 0 {
		return values, upstreamType, pathWarning(parsed), nil
	}
	if parsed.Host == "" {
		return nil, "", "", fmt.Errorf("proxy_pass 缺少主机或 upstream 组")
	}
	return []string{parsed.Host}, upstreamType, pathWarning(parsed), nil
}

func pathWarning(parsed *url.URL) string {
	if parsed.Path != "" && parsed.Path != "/" {
		return "proxy_pass 的路径前缀无法映射到当前站点模型，已仅导入上游地址"
	}
	return ""
}

func serverNames(server directive) []string {
	values := make([]string, 0)
	for _, item := range server.children {
		if item.name != "server_name" {
			continue
		}
		for _, value := range item.args {
			if value != "_" && !strings.HasPrefix(value, "$") {
				values = append(values, strings.ToLower(value))
			}
		}
	}
	sort.Strings(values)
	return compact(values)
}

func serverTLS(server directive) (int, bool) {
	for _, item := range server.children {
		if item.name != "listen" {
			continue
		}
		tlsEnabled := false
		port := 0
		for _, value := range item.args {
			if value == "ssl" {
				tlsEnabled = true
			}
			if parsedPort := listenPort(value); parsedPort > 0 {
				port = parsedPort
			}
		}
		if port == 443 {
			tlsEnabled = true
		}
		if tlsEnabled {
			if port == 0 {
				port = 443
			}
			return port, true
		}
	}
	return 443, false
}

func listenPort(value string) int {
	value = strings.TrimSuffix(value, "default_server")
	if port, err := strconv.Atoi(value); err == nil {
		return port
	}
	if _, port, err := net.SplitHostPort(value); err == nil {
		parsed, _ := strconv.Atoi(port)
		return parsed
	}
	return 0
}

func hasHTTPSRedirect(server directive) bool {
	for _, item := range findDirectives(server.children, "return") {
		if len(item.args) >= 2 && (item.args[0] == "301" || item.args[0] == "302" || item.args[0] == "307" || item.args[0] == "308") && strings.HasPrefix(item.args[1], "https://") {
			return true
		}
	}
	return false
}

func directiveEnabled(items []directive, name string) bool {
	for _, item := range findDirectives(items, name) {
		if len(item.args) > 0 && item.args[0] == "on" {
			return true
		}
	}
	return false
}

func hasAccessLog(server directive) bool {
	for _, item := range findDirectives(server.children, "access_log") {
		if len(item.args) == 0 || item.args[0] != "off" {
			return true
		}
	}
	return false
}

func findDirectives(items []directive, name string) []directive {
	result := make([]directive, 0)
	for _, item := range items {
		if item.name == name {
			result = append(result, item)
		}
		result = append(result, findDirectives(item.children, name)...)
	}
	return result
}

func parseDirectives(tokens []string, start int, stopOnBrace bool) ([]directive, int, error) {
	result := make([]directive, 0)
	index := start
	for index < len(tokens) {
		if tokens[index] == "}" {
			if !stopOnBrace {
				return nil, index, fmt.Errorf("Nginx 配置存在多余的右花括号")
			}
			return result, index + 1, nil
		}
		name := strings.ToLower(tokens[index])
		index++
		args := make([]string, 0)
		for index < len(tokens) && tokens[index] != ";" && tokens[index] != "{" && tokens[index] != "}" {
			args = append(args, tokens[index])
			index++
		}
		if index >= len(tokens) {
			return nil, index, fmt.Errorf("指令 %s 缺少分号或配置块", name)
		}
		switch tokens[index] {
		case ";":
			result = append(result, directive{name: name, args: args})
			index++
		case "{":
			children, next, err := parseDirectives(tokens, index+1, true)
			if err != nil {
				return nil, next, err
			}
			result = append(result, directive{name: name, args: args, children: children})
			index = next
		default:
			return nil, index, fmt.Errorf("指令 %s 缺少分号", name)
		}
	}
	if stopOnBrace {
		return nil, index, fmt.Errorf("Nginx 配置块缺少右花括号")
	}
	return result, index, nil
}

func tokenize(payload string) ([]string, error) {
	tokens := make([]string, 0)
	var current strings.Builder
	quote := rune(0)
	comment := false
	flush := func() {
		if current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}
	for _, character := range payload {
		if comment {
			if character == '\n' {
				comment = false
			}
			continue
		}
		if quote != 0 {
			if character == quote {
				quote = 0
			} else {
				current.WriteRune(character)
			}
			continue
		}
		switch character {
		case '#':
			flush()
			comment = true
		case '\'', '"':
			quote = character
		case '{', '}', ';':
			flush()
			tokens = append(tokens, string(character))
		case ' ', '\t', '\r', '\n':
			flush()
		default:
			current.WriteRune(character)
		}
	}
	if quote != 0 {
		return nil, fmt.Errorf("Nginx 配置存在未闭合的引号")
	}
	flush()
	return tokens, nil
}

func compact(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}
