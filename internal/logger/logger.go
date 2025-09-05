package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"time"
)

type LogLevel string

const (
	INFO  LogLevel = "INFO"
	DEBUG LogLevel = "DEBUG"
	ERROR LogLevel = "ERROR"
)

type ErrorInfo struct {
	Msg   string `json:"msg"`
	Stack string `json:"stack"`
}

type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Service   string                 `json:"service"`
	Action    string                 `json:"action"`
	Message   string                 `json:"message"`
	Hostname  string                 `json:"hostname"`
	RequestID string                 `json:"request_id"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Error     *ErrorInfo             `json:"error,omitempty"`
}

func getHostname() string {
	host, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return host
}

func Log(level LogLevel, service, action, message, requestID string, details map[string]interface{}, errObj error) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level,
		Service:   service,
		Action:    action,
		Message:   message,
		Hostname:  getHostname(),
		RequestID: requestID,
		Details:   details,
	}

	if level == ERROR && errObj != nil {
		entry.Error = &ErrorInfo{
			Msg:   errObj.Error(),
			Stack: string(debug.Stack()),
		}
	}

	data, _ := json.Marshal(entry)
	fmt.Println(string(data))
}
