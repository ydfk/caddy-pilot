package dnsprovider

import (
	"go-fiber-starter/internal/model/base"

	"gorm.io/gorm"
)

type DNSProvider struct {
	base.BaseModel
	Name            string         `gorm:"size:128;not null" json:"name"`
	ProviderType    string         `gorm:"size:32;not null;index" json:"provider_type"`
	EncryptedConfig string         `gorm:"type:text;not null" json:"-"`
	Enabled         bool           `gorm:"not null;default:true;index" json:"enabled"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}
