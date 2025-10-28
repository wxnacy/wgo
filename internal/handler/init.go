package handler

import (
	"os"
	"path/filepath"

	"github.com/wxnacy/go-tools"
	"github.com/wxnacy/wgo/internal/config"
	log "github.com/wxnacy/wgo/internal/logger"
)

func Init() {
	log.Init(config.Get())
	logger.Infoln("Init Begin")
	for _, dir := range []string{
		GetMainDir(),
		GetTempDir(),
	} {
		tools.DirExistsOrCreate(dir)
		WriteCode(BuiltinFuncCode, filepath.Join(dir, "builtin_func.go"))
		WriteCode(GetRequest().ToCode(), filepath.Join(dir, "request.go"))
	}
	logger.Infof("MainFile %s", GetMainFile())
	logger.Infof("TempDir %s", GetTempDir())
	logger.Infoln("Init End")
}

func Destory() {
	logger.Infoln("Destory Begin")
	os.RemoveAll(GetMainDir())
	os.RemoveAll(GetTempDir())
	logger.Infoln("Destory End")
}
