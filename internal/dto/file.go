package dto

import "github.com/mitchellh/go-homedir"

func NewDownloadReq() *DownloadReq {
	dlDir, _ := homedir.Expand("~/Downloads")
	return &DownloadReq{
		OutputDir: dlDir,
	}
}

type DownloadReq struct {
	GlobalReq
	OutputDir   string
	OutputPath  string
	IsSync      bool
	IsRecursion bool
}

func NewListReq() *ListReq {
	return &ListReq{}
}

type ListReq struct {
	GlobalReq
	Page       int
	Limit      int32
	WithoutTui bool
}

func NewRefreshReq() *RefreshReq {
	return &RefreshReq{}
}

type RefreshReq struct {
	GlobalReq
	IsSync bool
}

func NewDeleteReq() *DeleteReq {
	return &DeleteReq{}
}

type DeleteReq struct {
	GlobalReq
	FSID uint64
	Yes  bool
}
