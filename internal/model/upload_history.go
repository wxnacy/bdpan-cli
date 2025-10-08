package model

type UploadHistory struct {
	FSID           uint64 `json:"fs_id" gorm:"primaryKey;column:fs_id"`
	Path           string `json:"path"`
	Size           uint64 `json:"size"`
	ServerFilename string `json:"server_filename"`
	Category       int32  `json:"category"`
	MD5            string `json:"md5"`
	LocalMD5       string `json:"local_md5"`
	CTime          uint64 `json:"ctime"`
	MTime          uint64 `json:"mtime"`
	ORMModel
}

func (UploadHistory) TableName() string {
	return "upload_history"
}

func FindUploadHistoryByLocalMD5(md5 string) *UploadHistory {
	var m UploadHistory
	GetDB().Where(
		"local_md5 = ?",
		md5,
	).First(&m)
	return &m
}
