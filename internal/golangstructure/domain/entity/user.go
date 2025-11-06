package entity

import "time"

type User struct {
	Base
	Name      string     `gorm:"column:name;not null"`
	Age       int        `gorm:"column:age;not null"`
	DeletedAt *time.Time `gorm:"column:deleted_at;default:null"`
}

func (User) TableName() string {
	return "tbl_users"
}
