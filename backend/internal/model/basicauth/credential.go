package basicauth

import (
	"go-fiber-starter/internal/model/base"

	"gorm.io/gorm"
)

type Credential struct {
	base.BaseModel
	Name         string         `gorm:"size:128;not null" json:"name"`
	Username     string         `gorm:"size:128;not null" json:"username"`
	PasswordHash string         `gorm:"type:text;not null" json:"-"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
