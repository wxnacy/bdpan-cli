package model

import "gorm.io/gorm"

func Save(value interface{}) *gorm.DB {
	return GetDB().Save(value)
}

func FindFirstByID[T any](id int) *T {
	var v T
	GetDB().Where("id = ?", id).First(&v)
	return &v
}

func FindItems[T any]() []*T {
	var v []*T
	GetDB().Where("is_delete = 0").Order("create_time").Find(&v)
	return v
}
