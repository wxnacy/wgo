package wgo

import (
	"github.com/sirupsen/logrus"
)

var (
	Log *logrus.Entry
)

func init() {
	log := logrus.New()
	// log.SetReportCaller(true)
	log.SetFormatter(&LogFormatter{
		TextFormatter: &logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: "15:04:05",
			// CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {

			// filename := filepath.Base(frame.File)
			// return "", fmt.Sprintf("[%s:%d]", filename, frame.Line)

			// },
		},
	})
	Log = log.WithFields(logrus.Fields{})
}

type LogFormatter struct {
	*logrus.TextFormatter
}

func (f *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return f.TextFormatter.Format(entry)
}

func SetLogLevel(level logrus.Level) {
	Log.Logger.SetLevel(level)
}
