package logger

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"
)

var defaultLogger Logger = Logger{
	output: os.Stdout,
}

type Level int

const MODULE_NAME = "logger"

var FormattedANSI = map[string]string{
	"BLUE":   "\033[34m%v\033[0m",
	"GREEN":  "\033[32m%v\033[0m",
	"YELLOW": "\033[33m%v\033[0m",
	"RED":    "\033[31m%v\033[0m",
}

var LevelNames = map[Level]string{
	1: "DEBUG",
	2: "INFO",
	3: "WARN",
	4: "ERROR",
	5: "RAW",
}

type Logger struct {
	minLevel Level
	output   io.Writer
	webhooks []Webhook
}

type Webhook interface {
	SendMessageToWebhook(logMessage LogMessage)
}

type LogMessage struct {
	Level    Level
	Time     string
	File     string
	Line     int
	Messages []any
}

func SetOutput(w io.Writer) {
	defaultLogger.output = w
}

func AddWebhooks(wh Webhook) []Webhook {
	defaultLogger.webhooks = append(defaultLogger.webhooks, wh)
	return defaultLogger.webhooks
}

func touchLogLevel(log_level string) Level {
	switch log_level {
	case "DEBUG":
		return 1
	case "INFO":
		return 2
	case "WARN":
		return 3
	case "ERROR":
		return 4
	default:
		return 1
	}
}

func Default(log_level string) {
	defaultLogger.minLevel = touchLogLevel(log_level)
}

func Debug(messages ...any) {
	LogProcess(1, messages...)
}

func Info(messages ...any) {
	LogProcess(2, messages...)
}

func Warn(messages ...any) {
	LogProcess(3, messages...)
}

func Error(messages ...any) {
	LogProcess(4, messages...)
}

func Raw(messages ...any) {
	LogProcess(5, messages...)
}

func LogProcess(level Level, messages ...any) {
	logMessage := &LogMessage{}
	logMessage.Init(level, messages...)
	logMessage.FilePreProcess(2)
	logMessage.Print()
}

func (l *LogMessage) WebhookProcess() {
	for _, webhook := range defaultLogger.webhooks {
		webhook.SendMessageToWebhook(*l)
	}
}

func (l *LogMessage) Init(level Level, messages ...any) {
	l.Level = level
	l.Messages = messages
	l.Time = time.Now().Format("2006-01-02 15:04:05")
}

func (l *LogMessage) FilePreProcess(depth int) {
	if depth < 1 {
		depth = 1
	}

	_, file, line := GetLineFromCalledFunction()

	files := strings.Split(file, "/")
	if len(files) >= depth {
		file = strings.Join(files[(len(files)-depth):], "/")
	}

	l.File = file
	l.Line = line
}

func (l *LogMessage) Print() {
	if l.Level < defaultLogger.minLevel {
		return
	}

	var colorLevel string
	var colorMessage string
	messageText := fmt.Sprintln(l.Messages...)

	switch LevelNames[l.Level] {
	case "INFO":
		colorLevel = fmt.Sprintf(FormattedANSI["BLUE"], "INF")
		colorMessage = fmt.Sprintf(FormattedANSI["BLUE"], messageText)
	case "DEBUG":
		colorLevel = fmt.Sprintf(FormattedANSI["GREEN"], "DBG")
		colorMessage = fmt.Sprintf(FormattedANSI["GREEN"], messageText)
	case "WARN":
		colorLevel = fmt.Sprintf(FormattedANSI["YELLOW"], "WRN")
		colorMessage = fmt.Sprintf(FormattedANSI["YELLOW"], messageText)
	case "ERROR":
		colorLevel = fmt.Sprintf(FormattedANSI["RED"], "ERR")
		colorMessage = fmt.Sprintf(FormattedANSI["RED"], messageText)
	case "RAW":
		defaultLogger.output.Write([]byte(messageText))
		return
	}

	fileLine := fmt.Sprintf("%s:%d", l.File, l.Line)
	formattedMessage := fmt.Sprintf("%s %s \033[1;30m%-25.25s\033[0m %s", colorLevel, l.Time, fileLine, colorMessage)
	defaultLogger.output.Write([]byte(formattedMessage))
}

func GetLineFromCalledFunction() (functionName, fileName string, line int) {
	pcs := make([]uintptr, 10)
	n := runtime.Callers(2, pcs)

	pcs = pcs[:n]
	frames := runtime.CallersFrames(pcs)

	for {
		frame, more := frames.Next()
		if !strings.Contains(frame.Function, MODULE_NAME) {
			return frame.Function, frame.File, frame.Line
		}

		if !more {
			break
		}
	}

	return "unknown", "unknown", 0
}
