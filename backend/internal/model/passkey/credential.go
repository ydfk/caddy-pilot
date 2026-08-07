package passkey

import (
	"time"

	"go-fiber-starter/internal/model/base"

	"github.com/google/uuid"
)

type Credential struct {
	base.BaseModel
	UserID              uuid.UUID  `gorm:"type:char(36);not null;index" json:"user_id"`
	Name                string     `gorm:"size:128;not null" json:"name"`
	CredentialIDHash    string     `gorm:"size:64;not null;uniqueIndex" json:"-"`
	EncryptedCredential string     `gorm:"type:text;not null" json:"-"`
	LastUsedAt          *time.Time `json:"last_used_at,omitempty"`
}

func (Credential) TableName() string {
	return "passkey_credentials"
}
