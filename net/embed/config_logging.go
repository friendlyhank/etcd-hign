package embed

import "go.uber.org/zap"

// GetLogger returns the logger.
func (cfg Config) GetLogger() *zap.Logger {
	cfg.loggerMu.RLock()
	l := cfg.logger
	cfg.loggerMu.RUnlock()
	return l
}
