package systeminfo

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	buildversion "go-fiber-starter/pkg/version"
)

type InfoResponse struct {
	Version   string `json:"version"`
	HTTPPort  int    `json:"http_port"`
	HTTPSPort int    `json:"https_port"`
}

type InfoOutput struct{ Body InfoResponse }

func Info(_ context.Context, _ *struct{}) (*InfoOutput, error) {
	httpPort, err := configuredPort("CADDYPILOT_HTTP_PORT", 80)
	if err != nil {
		return nil, err
	}
	httpsPort, err := configuredPort("CADDYPILOT_HTTPS_PORT", 443)
	if err != nil {
		return nil, err
	}
	return &InfoOutput{Body: InfoResponse{
		Version: buildversion.Current, HTTPPort: httpPort, HTTPSPort: httpsPort,
	}}, nil
}

func configuredPort(name string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("%s 必须是 1 到 65535 之间的端口", name)
	}
	return port, nil
}
