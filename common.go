package main

import (
    "github.com/wxnacy/wgo/logger"
    "log"
)

var wlog *logger.Logger

func Logger() *logger.Logger {
    if wlog == nil {
        wlog = logger.NewLogger()
    }
    return wlog
}

func initLogger() {
    h, err := logger.NewRotatingFileHandler("/usr/local/var/log/wgo.log")
    handlerErr(err)
    Logger().AddHandler(h)
    Logger().SetLevel(logger.LevelDebug)
}

func handlerErr(err error) {
    if err != nil {
        Logger().Error(err)
        log.Fatal(err)
    }
}

