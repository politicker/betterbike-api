package logger

import (
	"fmt"
	"time"
)

type LogWriter struct {
	service string
}

func New(service string) LogWriter {
	return LogWriter{
		service: service,
	}
}

// TODO: Dynamically build canonical log line from params
func (writer *LogWriter) write(message, level string, params ...interface{}) (int, error) {
	timeStamp := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")
	return fmt.Printf("[%s] [%s] %s service=%s\n", timeStamp, level, message, writer.service)
}

func (writer *LogWriter) Debug(message string, params ...interface{}) {
	writer.write(message, "DEBUG", params...)
}

func (writer *LogWriter) Info(message string, params ...interface{}) {
	writer.write(message, "INFO", params...)
}

func (writer *LogWriter) Error(message string, params ...interface{}) {
	writer.write(message, "Error", params...)
}
