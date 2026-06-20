package base

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	Id        uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"` // 自动写入创建时间
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"` // 自动写入更新时间
}

func (base *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	// 创建记录前生成主键，兼容 SQLite、PostgreSQL 和 MySQL
	base.Id = uuid.New()
	return
}
