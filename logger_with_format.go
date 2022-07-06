package swag

import (
	"fmt"
	"io"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	// Default log format will output [INFO]: 2006-01-02T15:04:05Z07:00 - Log message
	defaultLogFormat       = "%time% %lvl%%msg%"
	defaultTimestampFormat = "2006/01/02 15:04:05"
)

func NewLogger(ws ...io.Writer) *logrus.Logger {
	var logger = logrus.New()
	logger.SetFormatter(&logFormatter{})
	if len(ws) > 0 {
		logger.SetOutput(ws[0])
	}
	return logger
}

type logFormatter struct {
	// Timestamp format
	TimestampFormat string
	// Available standard keys: time, msg, lvl
	// Also can include custom fields but limited to strings.
	// All of fields need to be wrapped inside %% i.e %time% %msg%
	LogFormat string
}

// Format building log message.
func (f *logFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var output = defaultLogFormat

	output = strings.Replace(output, "%time%", entry.Time.Format(defaultTimestampFormat), 1)

	output = strings.Replace(output, "%msg%", entry.Message, 1)

	var levelStr = ""
	if entry.Level > logrus.DebugLevel {
		levelStr = "[" + strings.ToUpper(entry.Level.String()) + "]"
	}
	output = strings.Replace(output, "%lvl%", levelStr, 1)

	for k, val := range entry.Data {
		output += fmt.Sprintf("\t%s=%v", k, val)
	}

	return []byte(output + "\n"), nil
}
