package logger

import (
	"context"
	"io"
	"time"

	"github.com/rs/zerolog"
)

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, err error, fields ...Field)
	Fatal(msg string, fields ...Field)
	WithContext(ctx context.Context) Logger
}

type ZerologLogger struct {
	logger zerolog.Logger
}

type Field struct {
	Key   string
	Value interface{}
}

func NewZerologLogger(output io.Writer) *ZerologLogger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	consoleWriter := zerolog.ConsoleWriter{
		Out:        output,
		TimeFormat: time.RFC3339,
	}
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()
	return &ZerologLogger{logger: logger}
}

func (l *ZerologLogger) Debug(msg string, fields ...Field) {
	event := l.logger.Debug()
	addFields(event, fields)
	event.Msg(msg)
}

func (l *ZerologLogger) Info(msg string, fields ...Field) {
	event := l.logger.Info()
	addFields(event, fields)
	event.Msg(msg)
}

func (l *ZerologLogger) Warn(msg string, fields ...Field) {
	event := l.logger.Warn()
	addFields(event, fields)
	event.Msg(msg)
}

func (l *ZerologLogger) Error(msg string, err error, fields ...Field) {
	event := l.logger.Error()
	if err != nil {
		event = event.Err(err)
	}
	addFields(event, fields)
	event.Msg(msg)
}

func (l *ZerologLogger) Fatal(msg string, fields ...Field) {
	event := l.logger.Fatal()
	addFields(event, fields)
	event.Msg(msg)
}

func (l *ZerologLogger) WithContext(ctx context.Context) Logger {
	l2 := *l

	if requestID := ctx.Value("request_id"); requestID != nil {
		if reqID, ok := requestID.(string); ok {
			l2.logger = l2.logger.With().Str("request_id", reqID).Logger()
		}
	}
	if userID := ctx.Value("user_id"); userID != nil {
		if uid, ok := userID.(string); ok {
			l2.logger = l2.logger.With().Str("user_id", uid).Logger()
		}
	}
	return &l2
}

func addFields(event *zerolog.Event, fields []Field) {
	for _, f := range fields {
		event = event.Interface(f.Key, f.Value)
	}
}
