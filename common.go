package main

import (
    "github.com/wxnacy/wgo/logger"
    "log"
    "fmt"
    "os"
    "time"
    "strings"
)

var wlog *logger.Logger         // 日志
var tempdir string              // 临时目录
var tempCmpltFile string

func tempDir() string {
    if tempdir == "" {
        tmp := os.TempDir()
        if !strings.HasSuffix(tmp, "/") {
            tmp += "/"

        }
        tempdir = fmt.Sprintf("%s%s-%d/", tmp, "wgo", time.Now().Unix())
        tempdir = fmt.Sprintf("%s%s-%d/", tmp, "wgo", 0)
    }
    return tempdir
}

func destroyTempDir() {
    err := os.RemoveAll(tempDir())
    handlerErr(err)
}

func tempCompleteFile() string {
    if tempCmpltFile == "" {
        tempCmpltFile = fmt.Sprintf("%swgo_complete.go", tempDir())
    }
    return tempCmpltFile
}
func Logger() *logger.Logger {
    if wlog == nil {
        wlog = logger.NewLogger()
    }
    return wlog
}

func initLogger() {
    initTempDir()

    h, err := logger.NewRotatingFileHandler(tempDir() + "wgo.log")
    handlerErr(err)
    Logger().AddHandler(h)
    Logger().SetLevel(logger.LevelError)
}

func handlerErr(err error) {
    if err != nil {
        Logger().Error(err)
        log.Fatal(err)
    }
}

