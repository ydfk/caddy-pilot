package dnsprovider

import (
	"go-fiber-starter/internal/model/base"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func EnvNames(id uuid.UUID) (string, string, string) {
	prefix := "CADDYPILOT_DNS_" + strings.ToUpper(strings.ReplaceAll(id.String(), "-", "_"))
	return prefix + "_ACCESS_KEY_ID", prefix + "_ACCESS_KEY_SECRET", prefix + "_REGION_ID"
}

type DNSProvider struct {
	base.BaseModel
	Name            string         `gorm:"size:128;not null" json:"name"`
	ProviderType    string         `gorm:"size:32;not null;index" json:"provider_type"`
	EncryptedConfig string         `gorm:"type:text;not null" json:"-"`
	Enabled         bool           `gorm:"not null;default:true;index" json:"enabled"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}
