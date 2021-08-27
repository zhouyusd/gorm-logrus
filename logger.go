package gorm_logrus

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"time"
)

type (
	Option  func(opt *options)
	options struct {
		log *logrus.Logger
		cfg logger.Config
	}
)

func WithLogger(log *logrus.Logger) Option {
	return func(opt *options) {
		opt.log = log
	}
}

func WithConfig(cfg logger.Config) Option {
	return func(opt *options) {
		opt.cfg = cfg
	}
}

type Logger struct {
	log *logrus.Logger
	cfg logger.Config
}

func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.cfg.LogLevel = level
	return &newLogger
}

// Info print info
func (l *Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.log.WithContext(ctx).Infof(msg, data...)
}

// Warn print warn messages
func (l *Logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.log.WithContext(ctx).Warnf(msg, data...)
}

// Error print error messages
func (l *Logger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.log.WithContext(ctx).Errorf(msg, data...)
}

// Trace print sql message
func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	switch {
	case err != nil && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.cfg.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			l.log.WithContext(ctx).WithFields(logrus.Fields{
				"file":          utils.FileWithLineNum(),
				logrus.ErrorKey: err,
			}).Errorf("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.log.WithContext(ctx).WithFields(logrus.Fields{
				"file":          utils.FileWithLineNum(),
				logrus.ErrorKey: err,
			}).Errorf("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.cfg.SlowThreshold && l.cfg.SlowThreshold != 0:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.cfg.SlowThreshold)
		if rows == -1 {
			l.log.WithContext(ctx).WithFields(logrus.Fields{
				"file":    utils.FileWithLineNum(),
				"slowLog": slowLog,
			}).Warnf("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.log.WithContext(ctx).WithFields(logrus.Fields{
				"file":    utils.FileWithLineNum(),
				"slowLog": slowLog,
			}).Warnf("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	default:
		sql, rows := fc()
		if rows == -1 {
			l.log.WithContext(ctx).WithFields(logrus.Fields{
				"file": utils.FileWithLineNum(),
			}).Debugf("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.log.WithContext(ctx).WithFields(logrus.Fields{
				"file": utils.FileWithLineNum(),
			}).Debugf("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

func New(opts ...Option) logger.Interface {
	var opt options
	for _, o := range opts {
		o(&opt)
	}
	if opt.log == nil {
		opt.log = logrus.StandardLogger()
	}
	return &Logger{
		log: opt.log,
		cfg: opt.cfg,
	}
}
