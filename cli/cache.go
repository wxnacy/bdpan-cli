package cli

import (
	"github.com/wxnacy/bdpan"
	"github.com/wxnacy/bdpan-cli/terminal"
)

var (
	// key: fileDirPath
	CacheSelectMap = make(map[string]*terminal.Select, 0)
	// key: fileDirPath
	CacheFilesMap     = make(map[string][]*bdpan.FileInfoDto, 0)
	CacheSyncModelMap = make(map[string]*bdpan.SyncModel, 0)
)

func init() {
	RefreshSyncModelCache()
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
