package logger

import (
	"github.com/vlxdisluv/shortener/config"
	"go.uber.org/zap"
)

// Log will be available throughout all code as a singleton.
// No code except Initialize should modify this variable.
// By default, a no-op logger is set, which logs nothing.
var Log *zap.Logger = zap.NewNop()

func Initialize(cfg *config.Config) error {
	lvl, err := zap.ParseAtomicLevel(cfg.LogLevel)
	if err != nil {
		return err
	}

	var zapCfg zap.Config
	if cfg.Environment == "production" {
		zapCfg = zap.NewProductionConfig()
	} else {
		zapCfg = zap.NewDevelopmentConfig()
	}

	// Set desired levels and outputs: info/debug to stdout, errors to stderr
	zapCfg.Level = lvl
	zapCfg.OutputPaths = []string{"stdout"}
	zapCfg.ErrorOutputPaths = []string{"stderr"}

	zl, err := zapCfg.Build()
	if err != nil {
		return err
	}

	Log = zl
	return nil
}
