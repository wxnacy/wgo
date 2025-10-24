package log

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"github.com/wxnacy/code-prompt/pkg/log"
	"github.com/wxnacy/go-tools"
)

func Init() error {
	SetLogFile()
	return nil
}

func SetLogLevel(level logrus.Level) {
	log.SetLogLevel(level)
}

func SetLogFile() {
	logPath := os.Getenv("HOME") + "/.local/share/wgo/log/wgo.log"
	tools.DirExistsOrCreate(filepath.Dir(logPath))
	// 设置按日期分割日志，最多十个文件
	logf, err := rotatelogs.New(
		logPath+".%Y%m%d",
		rotatelogs.WithLinkName(logPath),
		rotatelogs.WithRotationTime(24*time.Hour),
		rotatelogs.WithRotationCount(10),
	)
	if err != nil {
		GetLogger().Errorf("failed to create rotatelogs: %s", err)
		return
	}
	GetLogger().SetOutput(logf)
}

func GetLogger() *logrus.Logger {
	// return log.GetLogger()
	return log.GetLogger()
}

func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

func Infoln(args ...interface{}) {
	GetLogger().Infoln(args...)
}

func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

func Printf(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
	Infof(format, args...)
}
