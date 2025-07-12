package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// singleton logger instance
var Log *zap.SugaredLogger = zap.NewNop().Sugar()

// Initialize singleton logger
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = lvl
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006/01/02 15:04:05")
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.MessageKey = "msg"
	cfg.EncoderConfig.LevelKey = "level"

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl.Sugar()
	return nil
}
