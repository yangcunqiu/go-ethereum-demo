package dao

import "gorm.io/gorm"

type NetworkInfo struct {
	gorm.Model
	ChainId   int `gorm:"type:bigint;not null;comment:链id"`
	ChainName int `gorm:"type:varchar(200);not null;comment:链名称"`
}

func (ci NetworkInfo) TableName() string {
	return "chain_info"
}
