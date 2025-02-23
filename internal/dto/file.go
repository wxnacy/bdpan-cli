package dto

func NewDownloadReq() *DownloadReq {
	return &DownloadReq{}
}

type DownloadReq struct {
	GlobalReq
	OutputDir   string
	OutputPath  string
	IsSync      bool
	IsRecursion bool
}
