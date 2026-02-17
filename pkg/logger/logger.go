package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// globalLogger is the package-level logger instance
	globalLogger *zap.Logger
)

func init() {
	// Initialize with a default production logger
	globalLogger = NewProduction()
}

// NewProduction creates a new production logger
func NewProduction() *zap.Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, _ := config.Build()
	return logger
}

// NewDevelopment creates a new development logger
func NewDevelopment() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	logger, _ := config.Build()
	return logger
}

// NewCustom creates a logger with custom settings
func NewCustom(level zapcore.Level, development bool) *zap.Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	if development {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		level,
	)

	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *zap.Logger) {
	if globalLogger != nil {
		globalLogger.Sync()
	}
	globalLogger = logger
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	return globalLogger
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	globalLogger.Info(msg, fields...)
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	globalLogger.Debug(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	globalLogger.Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	globalLogger.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	globalLogger.Fatal(msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() error {
	return globalLogger.Sync()
}

// With creates a child logger with additional fields
func With(fields ...zap.Field) *zap.Logger {
	return globalLogger.With(fields...)
}
