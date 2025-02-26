package model

import (
	"fmt"
	"reflect"

	"github.com/wxnacy/bdpan"
	"github.com/wxnacy/go-tools"
	"gorm.io/gorm"
)

func NewRootFile() *File {
	file := File{
		ID:        1,
		FSID:      1,
		Path:      "/",
		FileType:  1,
		IsRefresh: 1,
	}
	return &file
}

func NewFileFromDto(fileInfoDto *bdpan.FileInfoDto) *File {
	file := File{
		ID:             fileInfoDto.FSID,
		FSID:           fileInfoDto.FSID,
		Path:           fileInfoDto.Path,
		Size:           fileInfoDto.Size,
		FileType:       fileInfoDto.FileType,
		Filename:       fileInfoDto.Filename,
		ServerFilename: fileInfoDto.ServerFilename,
		Category:       fileInfoDto.Category,
		Dlink:          fileInfoDto.Dlink,
		MD5:            fileInfoDto.MD5,
		ServerCTime:    fileInfoDto.ServerCTime,
		ServerMTime:    fileInfoDto.ServerMTime,
		LocalCTime:     fileInfoDto.LocalCTime,
		LocalMTime:     fileInfoDto.LocalMTime,
		// Thumbs:         fileInfoDto.Thumbs,
	}
	// if fileInfoDto.IsDir() {
	// file.Size = -1
	// }
	return &file
}

// 生成 sqlite3 建表语句,开头加上删除语句
// sqlite3 字段要符合 gorm 命名规则
type File struct {
	ID             uint64 `gorm:"primaryKey;"`
	FSID           uint64 `json:"fs_id" gorm:"column:fs_id"`
	Path           string `json:"path"`
	Size           int    `json:"size"`
	FileType       int    `json:"isdir" gorm:"column:is_dir"`
	Filename       string `json:"filename"`
	ServerFilename string `json:"server_filename"`
	Category       int    `json:"category"`
	Dlink          string `json:"dlink"`
	MD5            string `json:"md5"`
	ServerCTime    int64  `json:"server_ctime"`
	ServerMTime    int64  `json:"server_mtime"`
	LocalCTime     int64  `json:"local_ctime"`
	LocalMTime     int64  `json:"local_mtime"`
	IsRefresh      int    `json:"is_refresh"`
	// Level          int    `json:"level"`
}

func (File) TableName() string {
	return "file"
}

func (f *File) GetSize() string {
	return tools.FormatSize(int64(f.Size))
}

func (f *File) Save() *gorm.DB {
	return GetDB().Save(f)
}

func (f *File) Resave() *gorm.DB {
	GetDB().Where("id = ?", f.ID).Delete(f)
	return f.Save()
}

func FindNeedRefreshFiles(path string) []*File {
	var files []*File
	GetDB().Where("is_refresh = 0 and is_dir = 1 and path like ?", fmt.Sprintf("%s%%", path)).Find(&files)
	return files
}

func FindFirstByID(id uint64) *File {
	var file *File
	GetDB().Where("id = ?", id).First(&file)
	return file
}

func FindFirstByPath(path string) *File {
	var file *File
	GetDB().Where("path = ?", path).First(&file)
	return file
}

func FindFilesPrefixPath(path string, isDir bool) []*File {
	var files []*File
	d := 0
	if isDir {
		d = 1
	}
	GetDB().Where(
		"path like ? and is_dir = ?",
		fmt.Sprintf("%s%%", path),
		d,
	).Find(&files)
	return files
}

func copyFields(dst, src interface{}) {
	dstVal := reflect.ValueOf(dst).Elem()
	srcVal := reflect.ValueOf(src)

	for i := 0; i < dstVal.NumField(); i++ {
		srcField := srcVal.Type().Field(i)
		dstField := dstVal.FieldByName(srcField.Name)
		if dstField.IsValid() && dstField.CanSet() {
			dstField.Set(srcVal.Field(i))
		}
	}
}
