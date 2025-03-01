package model

import "github.com/wxnacy/go-bdpan"

func NewUser(user *bdpan.GetUserInfoRes) *User {
	return &User{
		GetUserInfoRes: user,
		ID:             user.GetUk(),
		Uk:             user.GetUk(),
		AvatarUrl:      user.GetAvatarUrl(),
		BaiduName:      user.GetBaiduName(),
		NetdiskName:    user.GetNetdiskName(),
		VipType:        user.GetVipType(),
		VipName:        user.GetVipName(),
	}
}

type User struct {
	*bdpan.GetUserInfoRes `gorm:"-"`
	ID                    int    `gorm:"primaryKey;"`
	Uk                    int    `json:"uk,omitempty"`
	AvatarUrl             string `json:"avatar_url,omitempty"`
	BaiduName             string `json:"baidu_name,omitempty"`
	NetdiskName           string `json:"netdisk_name,omitempty"`
	VipType               int32  `json:"vip_type,omitempty"`
	VipName               string `json:"vip_name,omitempty"`
}
