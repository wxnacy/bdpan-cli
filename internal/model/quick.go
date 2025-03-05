package model

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
)

type Quick struct {
	ID       uint64 `gorm:"primaryKey;"`
	Filename string `json:"filename"`
	Path     string `json:"path"`
	Key      string `json:"key"`
	ORMModel
}

func (Quick) TableName() string { return "quick" }
func (q Quick) Title() string {
	var text = fmt.Sprintf("%s\tKey: g%s", q.Filename, q.Key)
	return text
}
func (q Quick) FilterValue() string { return q.Filename }
func (q Quick) Description() string {
	// var desc = fmt.Sprintf("%s%s", q.Key, q.Path)
	return q.Path
}

func ToList(items []*Quick) []list.Item {
	_items := make([]list.Item, 0)
	for _, v := range items {
		_items = append(_items, v)
	}
	return _items
}
