package logger

import (
    "log"
    "os"
)

type IHandler interface {
    SetLevel(Level)
    getLevel() Level
    logger() *log.Logger

}

type RotatingFileHandler struct {
    log *log.Logger
    level Level
    logPath string
}

func NewRotatingFileHandler(logPath string) (IHandler, error) {
    h := &RotatingFileHandler{
        level: LevelDebug,
        logPath: logPath,
    }

    file, err := os.OpenFile(
        logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644,
    )
    if err != nil {
        return nil, err
    }

    h.log = log.New(file, "", log.LstdFlags)

    var i IHandler
    i = h
    return i, nil
}

func (this *RotatingFileHandler) SetLevel(level Level) {
    this.level = level
}

func (this *RotatingFileHandler) getLevel() Level {
    return this.level
}

func (this *RotatingFileHandler) logger() *log.Logger {
    return this.log
}

type StreamHandler struct {
    log *log.Logger
    level Level
}
func NewStreamHandler() (IHandler, error) {
    h := &StreamHandler{
        level: LevelDebug,
    }

    h.log = log.New(os.Stdout, "", log.LstdFlags)

    var i IHandler
    i = h
    return i, nil
}

func (this *StreamHandler) SetLevel(level Level) {
    this.level = level
}

func (this *StreamHandler) getLevel() Level {
    return this.level
}

func (this *StreamHandler) logger() *log.Logger {
    return this.log
}
