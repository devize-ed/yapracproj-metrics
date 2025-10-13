package logger

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Initialize singleton logger.
func Initialize(level string) (*zap.SugaredLogger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	cfg := zap.NewProductionConfig()
	// Add level
	cfg.Level = lvl
	// Add time
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.EpochMillisTimeEncoder
	// Add caller
	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.MessageKey = "msg"
	cfg.EncoderConfig.LevelKey = "level"
	// Disable stacktrace
	cfg.DisableStacktrace = true

	// Add sampling (cuts allocations on logs)
	cfg.Sampling = &zap.SamplingConfig{
		Initial:    100,
		Thereafter: 100,
	}

	zl, err := cfg.Build(
		zap.AddCaller(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return zl.Sugar(), nil
}

func SafeSync(sugar *zap.SugaredLogger) {
	if sugar == nil {
		return
	}
	if err := sugar.Sync(); err != nil {
		var pe *os.PathError
		if errors.As(err, &pe) && (errors.Is(pe.Err, syscall.EINVAL) || errors.Is(pe.Err, syscall.ENOTTY)) {
			return
		}
		if errors.Is(err, syscall.EINVAL) || errors.Is(err, syscall.ENOTTY) {
			return
		}
		_, _ = os.Stderr.WriteString("failed to sync logger: " + err.Error() + "\n")
	}
}
