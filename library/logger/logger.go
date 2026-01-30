package logger

import (
	"aitrics-vital-signs/library/envs"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logger struct {
	Logger     *zap.Logger
	GormLogger *gormLogger
}

var ZapLogger *logger

func MustInitZapLogger() {
	// Define custom encoder configuration
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "",
		MessageKey:     "message",
		StacktraceKey:  "",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,    // Capitalize the log level names
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 UTC timestamp format
		EncodeDuration: zapcore.SecondsDurationEncoder, // Duration in seconds
		EncodeCaller:   zapcore.ShortCallerEncoder,     // Short caller (file and line)
	}

	loglevel := strings.ToUpper(envs.LogLevel)
	var zapLogLevel zapcore.Level
	switch loglevel {
	case "DEBUG":
		zapLogLevel = zapcore.DebugLevel
	case "INFO":
		zapLogLevel = zapcore.InfoLevel
	case "WARN":
		zapLogLevel = zapcore.WarnLevel
	case "ERROR":
		zapLogLevel = zapcore.ErrorLevel
	case "FATAL":
		zapLogLevel = zapcore.FatalLevel
	default:
		zapLogLevel = zapcore.DebugLevel
	}

	stderrLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zapLogLevel && (level == zapcore.ErrorLevel || level == zapcore.FatalLevel || level == zapcore.WarnLevel)
	})
	stdoutLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zapLogLevel && (level == zapcore.DebugLevel || level == zapcore.InfoLevel)
	})

	stderrSyncer := zapcore.Lock(os.Stderr)
	stdoutSyncer := zapcore.Lock(os.Stdout)

	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			stderrSyncer,
			stderrLevel,
		),
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			stdoutSyncer,
			stdoutLevel,
		),
	)

	named := fmt.Sprintf("%s-%s", envs.ServerName, envs.ServiceType)
	l := zap.New(core).Named(named)
	ZapLogger = &logger{Logger: l, GormLogger: &gormLogger{
		Logger:                l,
		SkipErrRecordNotFound: true,
	}}
}
