package certificate

import (
	"go-fiber-starter/internal/model/base"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CertificateProfile struct {
	base.BaseModel
	Name            string         `gorm:"size:128;not null" json:"name"`
	CertificateType string         `gorm:"size:16;not null;index" json:"certificate_type"`
	Subjects        string         `gorm:"type:text;not null" json:"subjects"`
	ChallengeType   string         `gorm:"size:16;not null" json:"challenge_type"`
	DNSProviderID   *uuid.UUID     `gorm:"type:char(36);index" json:"dns_provider_id"`
	Enabled         bool           `gorm:"not null;default:true;index" json:"enabled"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}
