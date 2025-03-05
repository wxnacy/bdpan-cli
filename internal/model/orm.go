package model

import (
	"time"

	"github.com/wxnacy/bdpan-cli/internal/common"
)

type ORMModel struct {
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
	IsDelete   int8      `json:"is_delete" form:"is_delete" gorm:"index"`
}

func (o ORMModel) IsNil() bool {
	return common.IsNilTime(o.CreateTime)
}

func (o *ORMModel) Init() {
	o.CreateTime = time.Now()
	o.UpdateTime = time.Now()
}
