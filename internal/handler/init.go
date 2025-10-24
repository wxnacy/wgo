package handler

import (
	"os"
	"path/filepath"

	"github.com/wxnacy/go-tools"
	log "github.com/wxnacy/wgo/internal/logger"
)

func Init() {
	log.Init()
	logger.Infoln("Init Begin")
	// tools.DirExistsOrCreate(GetMainDir())
	// tools.DirExistsOrCreate(GetTempDir())
	// InitBuiltinFuncCode(filepath.Join(GetMainDir(), "builtin_func.go"))
	// InitBuiltinFuncCode(filepath.Join(GetTempDir(), "builtin_func.go"))
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

func GetWorkspace() string {
	return GetRequest().Workspace
}

func GetMainDir() string {
	return GetRequest().MainDir
}

func GetMainFile() string {
	return GetRequest().MainFile
}

func GetTempDir() string {
	return GetRequest().TempDir
}
