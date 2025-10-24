package log

import (
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	log     *logrus.Logger
	onceLog sync.Once
)

func GetLogger() *logrus.Logger {
	if log == nil {
		onceLog.Do(func() {
			log = initLogger()
		})
	}
	return log
}

func initLogger() *logrus.Logger {
	// 初始化 logrus.Logger 的具体实现
	logger := logrus.New()

	logger.SetFormatter(&LogFormatter{
		TextFormatter: &logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: "01-02 15:04:05",
			// CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			// filename := filepath.Base(frame.File)
			// return "", fmt.Sprintf("[%s:%d]", filename, frame.Line)
			// },
		},
	})
	// 可以根据需要设置更多的配置项
	return logger
}

type LogFormatter struct {
	*logrus.TextFormatter
}

func (f *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return f.TextFormatter.Format(entry)
}

func SetLogLevel(level logrus.Level) {
	GetLogger().SetLevel(level)
}

// 设置日志输出文件
func SetOutputFile(path string) error {
	logFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModeAppend|os.ModePerm)
	if err != nil {
		return err
	}
	GetLogger().SetOutput(logFile)
	return nil
}

// 将日志输出到写入流中
func SetOutputWriter(w io.Writer) error {
	GetLogger().SetOutput(w)
	return nil
}

func IsLoggerDebug() bool {
	return GetLogger().GetLevel() == logrus.DebugLevel
}

func LogInfoString(w io.Writer, s string) {
	GetLogger().SetOutput(w)
	GetLogger().Info(s)
}
