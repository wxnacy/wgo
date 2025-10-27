package dto

const (
	ENV_PRODUCTION = "production"
	ENV_DEV        = "dev"
)

func NewGlobalReq() *GlobalReq {
	return &GlobalReq{
		Env: ENV_PRODUCTION,
	}
}

type GlobalReq struct {
	IsVerbose bool
	Env       string
}

// 是否为开发环境
func (r GlobalReq) IsDev() bool {
	return r.Env == ENV_DEV
}

// 是否为正式环境
func (r GlobalReq) IsProduction() bool {
	return r.Env == ENV_PRODUCTION
}
