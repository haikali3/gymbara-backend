package utils

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func InitializeLogger() {
	var err error

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
		if err := os.Setenv("APP_ENV", env); err != nil {
			panic("Failed to set environment variable: " + err.Error())
		}
	}

	var encoder zapcore.Encoder
	if env == "development" {
		encoder = zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			TimeKey:       "time",
			LevelKey:      "level",
			NameKey:       "logger",
			CallerKey:     "caller",
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",
			EncodeTime:    zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
			EncodeLevel:   colorLevelEncoder(), // Use custom color level encoder
			EncodeCaller:  zapcore.ShortCallerEncoder,
		})
	} else {
		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "time"
		cfg.EncoderConfig.MessageKey = "msg"
		cfg.EncoderConfig.LevelKey = "level"
		cfg.EncoderConfig.CallerKey = "caller"
		cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
		Logger, err = cfg.Build()
		if err != nil {
			panic("Failed to initialize logger: " + err.Error())
		}
		return
	}

	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel)
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

func SyncLogger() error {
	err := Logger.Sync()
	// Ignore "inappropriate ioctl for device" errors.
	if err != nil && strings.Contains(err.Error(), "inappropriate ioctl for device") {
		return nil
	}
	return err
}

// Custom level encoder with colors
func colorLevelEncoder() zapcore.LevelEncoder {
	return func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		var color string
		switch level {
		case zapcore.DebugLevel:
			color = "\033[36m" // Cyan
		case zapcore.InfoLevel:
			color = "\033[32m" // Green
		case zapcore.WarnLevel:
			color = "\033[33m" // Yellow
		case zapcore.ErrorLevel:
			color = "\033[31m" // Red
		case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
			color = "\033[35m" // Magenta
		default:
			color = "\033[0m" // Default (no color)
		}
		enc.AppendString(color + level.CapitalString() + "\033[0m") // Reset color after level
	}
}
