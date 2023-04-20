package wgo

import (
	"os"
	"path/filepath"

	"github.com/wxnacy/gotool"
)

var (
	tempdir      string
	tempMainFile string
)

func init() {
	initTempDir()
}

func initTempDir() {
	if tempdir == "" {
		// TODO: 修改临时目录
		tmp := "/Users/wxnacy/Downloads" //  os.TempDir()
		// tempdir = filepath.Join(tmp, fmt.Sprintf("wgo-%d", os.Getpid()))
		tempdir = filepath.Join(tmp, "wgo-")
	}
	gotool.DirExistsOrCreate(tempdir)
	if tempMainFile == "" {
		tempMainFile = filepath.Join(tempdir, "main.go")
	}
}

func DestroyTempDir() error {
	return os.RemoveAll(GetTempDir())
}

func GetTempDir() string {
	return tempdir
}

func GetTempMainFile() string {
	return tempMainFile
}
