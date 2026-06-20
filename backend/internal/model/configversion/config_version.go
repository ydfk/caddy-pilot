package configversion

import "time"

type ConfigVersion struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Version        uint       `gorm:"uniqueIndex;not null" json:"version"`
	Reason         string     `gorm:"size:255" json:"reason"`
	BusinessConfig string     `gorm:"type:text;not null" json:"business_config"`
	CaddyJSON      string     `gorm:"type:text;not null" json:"caddy_json"`
	Status         string     `gorm:"size:32;not null;index" json:"status"`
	ErrorMessage   string     `gorm:"type:text" json:"error_message"`
	PublishedAt    *time.Time `json:"published_at"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
}
