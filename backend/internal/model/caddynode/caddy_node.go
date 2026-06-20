package caddynode

import "go-fiber-starter/internal/model/base"

type CaddyNode struct {
	base.BaseModel
	Name     string `gorm:"size:64;not null;uniqueIndex" json:"name"`
	AdminAPI string `gorm:"size:255;not null" json:"admin_api"`
	Enabled  bool   `gorm:"not null;default:true" json:"enabled"`
}
