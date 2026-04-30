package entity

type UserLog struct {
	Base
	UserID int    `gorm:"column:user_id;not null"`
	Action string `gorm:"column:action;not null"`
}

func (UserLog) TableName() string {
	return "tbl_user_logs"
}
