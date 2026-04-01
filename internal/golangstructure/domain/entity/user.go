package entity

import "gorm.io/gorm"

type User struct {
	Base
	Name      string          `gorm:"column:name;not null"`
	Age       int             `gorm:"column:age;not null"`
	Email     string          `gorm:"column:email;not null;unique"`
	Password  string          `gorm:"column:password;not null"`
	DeletedAt *gorm.DeletedAt `gorm:"column:deleted_at;default:null"`
}

func (User) TableName() string {
	return "tbl_users"
}
