package logger

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormLog "gorm.io/gorm/logger"
)

type gormLogger struct {
	*zap.Logger
	SlowThreshold         time.Duration
	SkipErrRecordNotFound bool
}

func (l *gormLogger) LogMode(gormLog.LogLevel) gormLog.Interface {
	return l
}

func (l *gormLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	l.Logger.Sugar().Infof(msg, args...)
}

func (l *gormLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.Logger.Sugar().Warnf(msg, args...)
}

func (l *gormLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	l.Logger.Sugar().Errorf(msg, args...)
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, _ := fc()
	if err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) && l.SkipErrRecordNotFound) {
		l.Logger.Sugar().Errorf("%s [%s]", sql, elapsed)
		return
	}

	if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		l.Logger.Sugar().Warnf("%s [%s]", sql, elapsed)
		return
	}

	l.Logger.Sugar().Debugf("%s [%s]", sql, elapsed)
}
