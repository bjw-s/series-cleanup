package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var zapLog *zap.Logger
var atomicLevel zap.AtomicLevel

func init() {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	encoderConfig.StacktraceKey = ""
	atomicLevel = zap.NewAtomicLevel()

	zapLog = zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.Lock(os.Stdout),
		atomicLevel,
	))
	defer zapLog.Sync()
}

func SetLevel(l string) {
	loglevel, err := zap.ParseAtomicLevel(l)
	if err != nil {
		return
	}
	atomicLevel.SetLevel(loglevel.Level())
}

func Info(message string, fields ...zap.Field) {
	zapLog.Info(message, fields...)
}

func Debug(message string, fields ...zap.Field) {
	zapLog.Debug(message, fields...)
}

func Error(message string, fields ...zap.Field) {
	zapLog.Error(message, fields...)
}

func Fatal(message string, fields ...zap.Field) {
	zapLog.Fatal(message, fields...)
}
