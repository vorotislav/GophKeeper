package logger

import (
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log levels.
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// Log formats.
const (
	LogFormatText = "text"
	LogFormatJSON = "json"
)

var (
	// ErrUnsupportedLogLevel returns then unspecified level was passed.
	ErrUnsupportedLogLevel = errors.New("unsupported log level")

	// ErrUnsupportedLogFormat returns then unspecified format was passed.
	ErrUnsupportedLogFormat = errors.New("unsupported log format")
)

// New creates new "zap" logger with little customization.
func New(level, format string, output string, verbose bool) (*zap.Logger, error) {
	var resolvedLevel zapcore.Level

	switch strings.ToLower(level) {
	case LogLevelDebug:
		resolvedLevel = zapcore.DebugLevel
	case LogLevelInfo:
		resolvedLevel = zapcore.InfoLevel
	case LogLevelWarn:
		resolvedLevel = zapcore.WarnLevel
	case LogLevelError:
		resolvedLevel = zapcore.ErrorLevel
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedLogLevel, level)
	}

	var resolvedFormat string

	switch strings.ToLower(format) {
	case LogFormatJSON:
		resolvedFormat = LogFormatJSON
	case LogFormatText:
		resolvedFormat = LogFormatText
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedLogFormat, format)
	}

	var config zap.Config

	if resolvedFormat == LogFormatJSON {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "time"
		config.EncoderConfig.MessageKey = "message"
		config.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	}

	// default configuration
	config.Level = zap.NewAtomicLevelAt(resolvedLevel)
	config.OutputPaths = []string{output}
	config.ErrorOutputPaths = []string{output}
	config.Development = verbose
	config.DisableStacktrace = !verbose

	return config.Build()
}
