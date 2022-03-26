package domain

import "github.com/fundwit/go-commons/types"

type AuthChannel int

const (
	AuthChannelInternal = AuthChannel(0)
)

type User struct {
	ID          types.ID `json:"id" gorm:"primary_key;type:BIGINT UNSIGNED NOT NULL"`
	Username    string   `json:"username" gorm:"type:VARCHAR(255) NOT NULL"`
	Email       string   `json:"email" gorm:"type:VARCHAR(255) NOT NULL"`
	Salt        string   `json:"salt" gorm:"type:VARCHAR(255) NOT NULL"`
	Avatar      string   `json:"avatar" gorm:"type:VARCHAR(255)"`
	Theme       string   `json:"theme" gorm:"type:VARCHAR(255)"`
	ThemeEditor string   `json:"theme_editor" gorm:"type:VARCHAR(255)"`
	RealName    string   `json:"real_name" gorm:"type:VARCHAR(255)"`
	PhoneNo     string   `json:"phone_no" gorm:"type:VARCHAR(255)"`
	IsLocked    bool     `json:"islock" gorm:"type:TINYINT NOT NULL DEFAULT '0'"`

	CreateTime types.Timestamp `json:"create_time" gorm:"type:DATETIME NOT NULL"`
	ModifyTime types.Timestamp `json:"modify_time" gorm:"type:DATETIME NOT NULL"`
}

type UserIdentity struct {
	ID types.ID `json:"id" gorm:"primary_key;type:BIGINT UNSIGNED NOT NULL"`

	User        types.ID    `json:"user" gorm:"type:BIGINT UNSIGNED NOT NULL"`
	AuthChannel AuthChannel `json:"auth_channel" gorm:"type:INT UNSIGNED NOT NULL"`
	ChannelKey  string      `json:"channel_key" gorm:"type:VARCHAR(255) NOT NULL"`
}

func (r *User) TableName() string {
	return "user"
}

func (r *UserIdentity) TableName() string {
	return "user_identity"
}
