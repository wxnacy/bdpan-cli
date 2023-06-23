package cli

import "github.com/wxnacy/bdpan"

func GetRootFile() *bdpan.FileInfoDto {
	return &bdpan.FileInfoDto{
		Path:     "/",
		FileType: 1,
	}
}

func FilterFileFiles(files []*bdpan.FileInfoDto) (filterFiles []*bdpan.FileInfoDto) {
	for _, f := range files {
		if !f.IsDir() {
			filterFiles = append(filterFiles, f)
		}
	}
	return
}
