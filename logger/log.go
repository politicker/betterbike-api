package logger

import (
	"fmt"
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
	return fmt.Printf("[%s] %s service=%s\n", level, message, writer.service)
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
