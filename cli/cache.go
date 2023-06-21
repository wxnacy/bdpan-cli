package cli

import (
	"github.com/wxnacy/bdpan"
	"github.com/wxnacy/bdpan-cli/terminal"
)

var (
	// key: fileDirPath
	cacheFileSelectMap = make(map[string]*FileSelectCache, 0)
	CacheSyncModelMap  = make(map[string]*bdpan.SyncModel, 0)
)

func init() {
	RefreshSyncModelCache()
}

func NewFileSelectCache(dir string, s *terminal.Select, files []*bdpan.FileInfoDto) *FileSelectCache {
	return &FileSelectCache{
		Dir:         dir,
		Files:       files,
		SelectIndex: s.SelectIndex,
	}
}

func GetFileSelectCache(dir string) *FileSelectCache {
	return cacheFileSelectMap[dir]
}

type FileSelectCache struct {
	Dir         string
	Files       []*bdpan.FileInfoDto
	SelectIndex int
}

func (f *FileSelectCache) Save() {
	cacheFileSelectMap[f.Dir] = f
}

func RefreshSyncModelCache() {
	CacheSyncModelMap = bdpan.GetSyncModels()
}

func GetSyncModelsByFilePath(path string) []*bdpan.SyncModel {
	return bdpan.GetSyncModelsByRemote(path)
}

func IsSync(file *bdpan.FileInfoDto) bool {
	if len(GetSyncModelsByFilePath(file.Path)) > 0 {
		return true
	}
	return false
}
