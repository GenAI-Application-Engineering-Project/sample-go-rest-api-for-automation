package applogger

type LoggerInterface interface {
	LogError(err error, msg string)
}
