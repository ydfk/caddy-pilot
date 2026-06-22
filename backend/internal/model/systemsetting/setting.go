package systemsetting

import "time"

type Setting struct {
	Key       string    `gorm:"primaryKey;size:128"`
	Value     string    `gorm:"type:text;not null"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
