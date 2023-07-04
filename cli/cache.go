package cli

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wxnacy/bdpan"
	"github.com/wxnacy/bdpan-cli/terminal"
	"github.com/wxnacy/go-tools"
)

var (
	// key: fileDirPath
	cacheFileSelectMap  = make(map[string]*FileSelectCache, 0)
	CacheSyncModelMap   = make(map[string]*bdpan.SyncModel, 0)
	cacheFilesDir       = bdpan.JoinCache("files")
	cacheSelectIndexMap = make(map[string]int, 0)
)

func init() {
	RefreshSyncModelCache()
	tools.DirExistsOrCreate(cacheFilesDir)
}

func SetCacheSelectIndex(a SystemAction, i int) {
	cacheSelectIndexMap[strconv.Itoa(int(a))] = i
}

func GetCacheSelectIndex(a SystemAction) (i int, f bool) {
	i, f = cacheSelectIndexMap[strconv.Itoa(int(a))]
	return
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

func GetAllCacheFiles() []*bdpan.FileInfoDto {
	var files = make([]*bdpan.FileInfoDto, 0)
	for _, s := range cacheFileSelectMap {
		files = append(s.Files, files...)
	}
	return files
}

func GetAllLocalFiles() (files []*bdpan.FileInfoDto, err error) {
	err = tools.NewFileFilter(cacheFilesDir, func(paths []string) error {
		for _, path := range paths {
			var subFiles []*bdpan.FileInfoDto
			err = tools.FileReadForInterface(path, &subFiles)
			if err != nil {
				return err
			}
			files = append(files, subFiles...)
		}
		return nil
	}).Run()
	return
}

type FileSelectCache struct {
	Dir         string
	Files       []*bdpan.FileInfoDto
	SelectIndex int
}

func (f *FileSelectCache) Save() {
	cacheFileSelectMap[f.Dir] = f
	name := strings.ReplaceAll(f.Dir, "/", "")
	path := filepath.Join(cacheFilesDir, name)
	tools.FileWriteWithInterface(path, f.Files)
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
