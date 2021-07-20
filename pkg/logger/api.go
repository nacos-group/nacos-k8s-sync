package logger

// Info is info level
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

// Warn is warning level
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

// Error is error level
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

// Debug is debug level
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

// Infof is format info level
func Infof(fmt string, args ...interface{}) {
	GetLogger().Infof(fmt, args...)
}

// Warnf is format warning level
func Warnf(fmt string, args ...interface{}) {
	GetLogger().Warnf(fmt, args...)
}

// Errorf is format error level
func Errorf(fmt string, args ...interface{}) {
	GetLogger().Errorf(fmt, args...)
}

// Debugf is format debug level
func Debugf(fmt string, args ...interface{}) {
	GetLogger().Debugf(fmt, args...)
}

func Sync() {
	_ = GetLogger().Sync()
}
