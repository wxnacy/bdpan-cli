package dto

func NewGlobalReq() *GlobalReq {
	return &GlobalReq{}
}

type GlobalReq struct {
	AppId     string
	IsVerbose bool
	Path      string
	Config    string
}
