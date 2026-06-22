package systeminfo

import (
	"context"

	buildversion "go-fiber-starter/pkg/version"
)

type InfoResponse struct {
	Version string `json:"version"`
}

type InfoOutput struct{ Body InfoResponse }

func Info(_ context.Context, _ *struct{}) (*InfoOutput, error) {
	return &InfoOutput{Body: InfoResponse{Version: buildversion.Current}}, nil
}
