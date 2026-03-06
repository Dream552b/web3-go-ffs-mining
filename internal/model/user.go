package model

// User 用户表
type User struct {
	BaseModel
	Username     string `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Password     string `gorm:"size:128;not null" json:"-"`
	Email        string `gorm:"uniqueIndex;size:128" json:"email"`
	Phone        string `gorm:"index;size:32" json:"phone"`
	Role         string `gorm:"size:16;default:user" json:"role"` // user | admin
	Status       int8   `gorm:"default:1" json:"status"`          // 1=正常 0=禁用
	ReferrerID   *uint  `gorm:"index" json:"referrer_id"`         // 邀请人 ID
	InviteCode   string `gorm:"uniqueIndex;size:16" json:"invite_code"`
}

func (User) TableName() string { return "users" }

const (
	UserStatusActive   int8 = 1
	UserStatusDisabled int8 = 0

	UserRoleUser  = "user"
	UserRoleAdmin = "admin"
)
