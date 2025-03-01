package model

import (
	"fmt"
	"reflect"

	"github.com/wxnacy/go-bdpan"
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

func NewFiles(files []*bdpan.FileInfo) []*File {
	res := make([]*File, 0)
	for _, f := range files {
		res = append(res, NewFile(f))
	}
	return res
}

func NewFile(f *bdpan.FileInfo) *File {
	file := File{
		FileInfo:       f,
		ID:             f.FSID,
		FSID:           f.FSID,
		Path:           f.Path,
		Size:           f.Size,
		FileType:       f.FileType,
		Filename:       f.Filename,
		ServerFilename: f.ServerFilename,
		Category:       f.Category,
		Dlink:          f.Dlink,
		MD5:            f.MD5,
		ServerCTime:    f.ServerCTime,
		ServerMTime:    f.ServerMTime,
		LocalCTime:     f.LocalCTime,
		LocalMTime:     f.LocalMTime,
		// Thumbs:         fileInfoDto.Thumbs,
	}
	return &file
}

// 生成 sqlite3 建表语句,开头加上删除语句
// sqlite3 字段要符合 gorm 命名规则
type File struct {
	*bdpan.FileInfo `gorm:"-"`
	ID              uint64 `gorm:"primaryKey;"`
	FSID            uint64 `json:"fs_id" gorm:"column:fs_id"`
	Path            string `json:"path"`
	Size            int    `json:"size"`
	FileType        int    `json:"isdir" gorm:"column:is_dir"`
	Filename        string `json:"filename"`
	ServerFilename  string `json:"server_filename"`
	Category        int    `json:"category"`
	Dlink           string `json:"dlink"`
	MD5             string `json:"md5"`
	ServerCTime     int64  `json:"server_ctime"`
	ServerMTime     int64  `json:"server_mtime"`
	LocalCTime      int64  `json:"local_ctime"`
	LocalMTime      int64  `json:"local_mtime"`

	// custom
	IsRefresh int `json:"is_refresh"`
	Level     int `json:"level"`
}

func (File) TableName() string {
	return "file"
}

func (f *File) Fill() *File {
	f.FileInfo = &bdpan.FileInfo{
		FSID:           f.FSID,
		Path:           f.Path,
		Size:           f.Size,
		FileType:       f.FileType,
		Filename:       f.Filename,
		ServerFilename: f.ServerFilename,
		Category:       f.Category,
		Dlink:          f.Dlink,
		MD5:            f.MD5,
		ServerCTime:    f.ServerCTime,
		ServerMTime:    f.ServerMTime,
		LocalCTime:     f.LocalCTime,
		LocalMTime:     f.LocalMTime,
	}
	return f
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
	for _, v := range files {
		v.Fill()
	}
	return files
}

func FindFirstByID(id uint64) *File {
	var file *File
	GetDB().Where("id = ?", id).First(&file)
	return file.Fill()
}

func FindFirstByPath(path string) *File {
	var file *File
	GetDB().Where("path = ?", path).First(&file)
	return file.Fill()
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
	for _, v := range files {
		v.Fill()
	}
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
