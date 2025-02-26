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

func NewListReq() *ListReq {
	return &ListReq{}
}

type ListReq struct {
	GlobalReq
	Page  int
	Limit int32
}

func NewRefreshReq() *RefreshReq {
	return &RefreshReq{}
}

type RefreshReq struct {
	GlobalReq
	IsSync bool
}
