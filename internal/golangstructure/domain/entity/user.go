package entity

type User struct {
	ID   int    `gorm:"column:id;primaryKey;autoIncrement"`
	Name string `gorm:"column:name;not null"`
	Age  int    `gorm:"column:age;not null"`
}

func (User) TableName() string {
	return "tbl_users"
}
