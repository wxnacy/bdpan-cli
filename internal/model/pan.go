package model

import "github.com/wxnacy/go-bdpan"

func NewPan(id int, pan *bdpan.GetPanInfoRes) *Pan {
	return &Pan{
		GetPanInfoRes: pan,
		ID:            id,
		Total:         pan.GetTotal(),
		Free:          pan.GetFree(),
		Expire:        pan.GetExpire(),
		Used:          pan.GetUsed(),
	}
}

type Pan struct {
	*bdpan.GetPanInfoRes `gorm:"-"`
	ID                   int   `gorm:"primaryKey;"`
	Total                int64 `json:"total,omitempty"`
	Free                 int64 `json:"free,omitempty"`
	Expire               bool  `json:"expire,omitempty"`
	Used                 int64 `json:"used,omitempty"`
}
