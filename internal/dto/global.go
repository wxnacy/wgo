package dto

func NewGlobalReq() *GlobalReq {
	return &GlobalReq{}
}

type GlobalReq struct {
	IsVerbose bool
}
