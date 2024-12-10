package logger

import (
	"testing"

	"go.uber.org/zap"
)

// var TestLogger zap.SugaredLogger

func initLogger() {
	logger, lerr := zap.NewDevelopment()
	if lerr != nil {
		panic(lerr)
	}
	Logger = *logger.Sugar()
}

func TestLoggerInitialization(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Logger initialization failed with panic: %v", r)
		}
	}()

	initLogger()

	if Logger.Desugar() == nil {
		t.Errorf("Logger should be initialized, but got nil")
	}
}
