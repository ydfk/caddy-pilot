package nginximport

import "testing"

func TestParseMergesRedirectAndTLSProxy(t *testing.T) {
	result, err := Parse(`
upstream app_backend {
    server 127.0.0.1:3000;
    server 127.0.0.1:3001;
}
server {
    listen 80;
    server_name example.com www.example.com;
    return 301 https://$host$request_uri;
}
server {
    listen 443 ssl http2;
    server_name example.com www.example.com;
    gzip on;
    access_log /var/log/nginx/example.log;
    location / { proxy_pass http://app_backend; }
}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Sites) != 1 {
		t.Fatalf("站点数量不正确: %+v", result)
	}
	site := result.Sites[0]
	if !site.EnableHTTPS || !site.ForceHTTPS || !site.EnableGzip || !site.EnableLog {
		t.Fatalf("站点开关不正确: %+v", site)
	}
	if len(site.Upstreams) != 2 || site.Upstreams[1] != "127.0.0.1:3001" {
		t.Fatalf("upstream 组未展开: %+v", site.Upstreams)
	}
}

func TestParseKeepsNonstandardHTTPSPort(t *testing.T) {
	result, err := Parse(`server {
    listen 8443 ssl;
    server_name admin.example.com;
    location / { proxy_pass https://10.0.0.8:9443/api; }
}`)
	if err != nil {
		t.Fatal(err)
	}
	site := result.Sites[0]
	if site.UpstreamType != "https" || len(result.Warnings) != 1 {
		t.Fatalf("非标准端口解析不正确: %+v", result)
	}
}

func TestParseRejectsConfigWithoutProxySite(t *testing.T) {
	if _, err := Parse(`server { listen 80; server_name example.com; root /var/www; }`); err == nil {
		t.Fatal("应拒绝没有 proxy_pass 的配置")
	}
}

func TestParseReadsServerInsideHTTPBlock(t *testing.T) {
	result, err := Parse(`events {} http { server {
		listen 80;
		server_name nested.example.com;
		location / { proxy_pass http://127.0.0.1:8080; }
	} }`)
	if err != nil || len(result.Sites) != 1 || result.Sites[0].Domains[0] != "nested.example.com" {
		t.Fatalf("未解析 http 块中的站点: %+v, %v", result, err)
	}
}
