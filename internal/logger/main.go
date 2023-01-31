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

// SetLevel sets the zap log level
func SetLevel(l string) {
	loglevel, err := zap.ParseAtomicLevel(l)
	if err != nil {
		return
	}
	atomicLevel.SetLevel(loglevel.Level())
}

// Info logs a message at level Info on the zap logger
func Info(message string, fields ...zap.Field) {
	zapLog.Info(message, fields...)
}

// Debug logs a message at level Debug on the zap logger
func Debug(message string, fields ...zap.Field) {
	zapLog.Debug(message, fields...)
}

// Error logs a message at level Error on the zap logger
func Error(message string, fields ...zap.Field) {
	zapLog.Error(message, fields...)
}

// Fatal logs a message at level Fatal on the zap logger
func Fatal(message string, fields ...zap.Field) {
	zapLog.Fatal(message, fields...)
}
