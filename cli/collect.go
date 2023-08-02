package cli

import (
	"time"

	"github.com/wxnacy/bdpan"
	"github.com/wxnacy/go-tools"
)

var (
	collectPath = bdpan.JoinStoage("collect")
)

func NewCollect(file *bdpan.FileInfoDto) CollectInfo {
	return CollectInfo{
		File:       file,
		CreateTime: time.Now(),
	}
}

type CollectInfo struct {
	File       *bdpan.FileInfoDto `json:"file"`
	CreateTime time.Time          `json:"create_time"`
}

func GetCollectFiles() []*bdpan.FileInfoDto {
	items := make([]*bdpan.FileInfoDto, 0)
	m := make(map[string]CollectInfo, 0)
	err := tools.FileReadForInterface(collectPath, &m)
	Log.Info(err)
	if err != nil {
		return items
	}
	for _, item := range m {
		Log.Infof("collect %v", item)
		items = append(items, item.File)
	}
	return items
}

func IsCollect(file *bdpan.FileInfoDto) bool {
	m, err := tools.FileReadToMap(collectPath)
	if err != nil {
		return false
	}
	_, ok := m[file.Path]
	return ok
}

func SaveCollect(file *bdpan.FileInfoDto) error {
	m := make(map[string]CollectInfo, 0)
	err := tools.FileReadForInterface(collectPath, &m)
	if err != nil {
		m = make(map[string]CollectInfo, 0)
	}
	m[file.Path] = NewCollect(file)
	return tools.FileWriteWithInterface(collectPath, m)
}

func CancelCollect(file *bdpan.FileInfoDto) error {
	m := make(map[string]CollectInfo, 0)
	err := tools.FileReadForInterface(collectPath, &m)
	if err != nil {
		return nil
	}
	delete(m, file.Path)
	return tools.FileWriteWithInterface(collectPath, m)
}
