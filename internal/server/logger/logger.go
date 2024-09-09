package logger

import "go.uber.org/zap"

var Logger zap.SugaredLogger

func init() {
	logger, lerr := zap.NewDevelopment()
	if lerr != nil {
		panic(lerr)
	}
	Logger = *logger.Sugar()
}
