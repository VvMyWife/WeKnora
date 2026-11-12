package logger

import (
	"context"
	gormlog "gorm.io/gorm/logger"
	"time"
)

// PrivateGORMLogger preserves normal SQL logging and suppresses anonymous evidence.
type PrivateGORMLogger struct{ gormlog.Interface }

func (l PrivateGORMLogger) LogMode(level gormlog.LogLevel) gormlog.Interface {
	return PrivateGORMLogger{l.Interface.LogMode(level)}
}
func (l PrivateGORMLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if PrivateContent(ctx) {
		return
	}
	l.Interface.Trace(ctx, begin, fc, err)
}
func (l PrivateGORMLogger) Info(ctx context.Context, s string, args ...interface{}) {
	if !PrivateContent(ctx) {
		l.Interface.Info(ctx, s, args...)
	}
}
func (l PrivateGORMLogger) Warn(ctx context.Context, s string, args ...interface{}) {
	if !PrivateContent(ctx) {
		l.Interface.Warn(ctx, s, args...)
	}
}
func (l PrivateGORMLogger) Error(ctx context.Context, s string, args ...interface{}) {
	if !PrivateContent(ctx) {
		l.Interface.Error(ctx, s, args...)
	}
}
