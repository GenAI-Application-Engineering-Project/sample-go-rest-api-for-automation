package applogger

type LoggerInterface interface {
	LogError(op string, err error, msg string)
}
