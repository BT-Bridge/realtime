package shared

import (
	"context"
	"fmt"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

type Logger struct {
	*otelzap.Logger
	Fields []zap.Field
}

func NewLogger(customFields ...zap.Field) *Logger {
	config := zap.NewProductionConfig()
	config.DisableStacktrace = true
	z, err := config.Build()
	if err != nil {
		panic(err)
	}
	return &Logger{
		Logger: otelzap.New(z),
		Fields: customFields,
	}
}

func mergeFields(source1, source2 []zap.Field) []zap.Field {
	fields := make([]zap.Field, 0, len(source1)+len(source2))
	fields = append(fields, source1...)
	fields = append(fields, source2...)
	return fields
}

func (l Logger) Info(ctx context.Context, msg string) {
	l.Logger.Ctx(ctx).Info(msg, l.Fields...)
}

func (l Logger) Infof(ctx context.Context, msg string, args ...any) {
	l.Logger.Ctx(ctx).Info(fmt.Sprintf(msg, args...), l.Fields...)
}

func (l Logger) InfoFields(ctx context.Context, msg string, fields ...zap.Field) {
	l.Logger.Ctx(ctx).Info(msg, mergeFields(l.Fields, fields)...)
}

func (l Logger) Trace(ctx context.Context, msg string) {
	l.Logger.Ctx(ctx).Debug(msg, l.Fields...)
}

func (l Logger) Tracef(ctx context.Context, msg string, args ...any) {
	l.Logger.Ctx(ctx).Debug(fmt.Sprintf(msg, args...), l.Fields...)
}

func (l Logger) Debug(ctx context.Context, msg string) {
	l.Logger.Ctx(ctx).Debug(msg, l.Fields...)
}

func (l Logger) Debugf(ctx context.Context, msg string, args ...any) {
	l.Logger.Ctx(ctx).Debug(fmt.Sprintf(msg, args...), l.Fields...)
}

func (l Logger) Warn(ctx context.Context, msg string) {
	l.Logger.Ctx(ctx).Warn(msg, l.Fields...)
}

func (l Logger) Warnf(ctx context.Context, msg string, args ...any) {
	l.Logger.Ctx(ctx).Warn(fmt.Sprintf(msg, args...), l.Fields...)
}

func (l Logger) WarnFields(ctx context.Context, msg string, fields ...zap.Field) {
	l.Logger.Ctx(ctx).Warn(msg, mergeFields(l.Fields, fields)...)
}

func (l Logger) Error(ctx context.Context, err error, msg string) {
	if msg == "" {
		l.Logger.Ctx(ctx).Error(err.Error(), l.Fields...)
		return
	}
	l.Logger.Ctx(ctx).Error(msg, mergeFields(l.Fields, []zap.Field{zap.Error(err)})...)
}

func (l Logger) Errorf(ctx context.Context, err error, msg string, args ...any) {
	l.Logger.Ctx(ctx).Error(fmt.Sprintf(msg, args...), mergeFields(l.Fields, []zap.Field{zap.Error(err)})...)
}

func (l Logger) ErrorFields(ctx context.Context, err error, msg string, fields ...zap.Field) {
	if msg == "" {
		l.Logger.Ctx(ctx).Error(err.Error(), fields...)
		return
	}
	fields = append(fields, zap.Error(err))
	l.Logger.Ctx(ctx).Error(msg, fields...)
}

func (l Logger) Panic(ctx context.Context, msg string) {
	l.Logger.Ctx(ctx).Panic(msg)
}

func (l Logger) Panicf(ctx context.Context, msg string, args ...any) {
	l.Logger.Ctx(ctx).Panic(fmt.Sprintf(msg, args...), l.Fields...)
}

func (l Logger) Fatal(ctx context.Context, msg string) {
	l.Logger.Ctx(ctx).Fatal(msg, l.Fields...)
}

func (l Logger) Fatalf(ctx context.Context, msg string, args ...any) {
	l.Logger.Ctx(ctx).Fatal(fmt.Sprintf(msg, args...), l.Fields...)
}

func (l Logger) NoCtxInfof(msg string, args ...any) {
	l.Logger.Ctx(context.Background()).Info(fmt.Sprintf(msg, args...), l.Fields...)
}

func (l Logger) NoCtxInfoFields(msg string, fields ...zap.Field) {
	l.Logger.Ctx(context.Background()).Info(msg, mergeFields(l.Fields, fields)...)
}

func (l Logger) NoCtxTrace(msg string) {
	l.Logger.Ctx(context.Background()).Debug(msg, l.Fields...)
}

func (l Logger) NoCtxTracef(msg string, args ...any) {
	l.Logger.Ctx(context.Background()).Debug(fmt.Sprintf(msg, args...), l.Fields...)
}

func (l Logger) NoCtxDebug(msg string) {
	l.Logger.Ctx(context.Background()).Debug(msg, l.Fields...)
}

func (l Logger) NoCtxDebugf(msg string, args ...any) {
	l.Logger.Ctx(context.Background()).Debug(fmt.Sprintf(msg, args...), l.Fields...)
}

func (l Logger) NoCtxWarn(msg string) {
	l.Logger.Ctx(context.Background()).Warn(msg, l.Fields...)
}

func (l Logger) NoCtxWarnf(msg string, args ...any) {
	l.Logger.Ctx(context.Background()).Warn(fmt.Sprintf(msg, args...), l.Fields...)
}

func (l Logger) NoCtxWarnFields(msg string, fields ...zap.Field) {
	l.Logger.Ctx(context.Background()).Warn(msg, mergeFields(l.Fields, fields)...)
}

func (l Logger) NoCtxError(err error, msg string) {
	if msg == "" {
		l.Logger.Ctx(context.Background()).Error(err.Error(), l.Fields...)
		return
	}
	l.Logger.Ctx(context.Background()).Error(msg, mergeFields(l.Fields, []zap.Field{zap.Error(err)})...)
}

func (l Logger) NoCtxErrorf(err error, msg string, args ...any) {
	l.Logger.Ctx(context.Background()).Error(fmt.Sprintf(msg, args...), mergeFields(l.Fields, []zap.Field{zap.Error(err)})...)
}

func (l Logger) NoCtxErrorFields(err error, msg string, fields ...zap.Field) {
	if msg == "" {
		l.Logger.Ctx(context.Background()).Error(err.Error(), fields...)
		return
	}
	fields = append(fields, zap.Error(err))
	l.Logger.Ctx(context.Background()).Error(msg, fields...)
}

func (l Logger) NoCtxPanic(msg string) {
	l.Logger.Ctx(context.Background()).Panic(msg)
}

func (l Logger) NoCtxPanicf(msg string, args ...any) {
	l.Logger.Ctx(context.Background()).Panic(fmt.Sprintf(msg, args...), l.Fields...)
}

func (l Logger) NoCtxFatal(msg string) {
	l.Logger.Ctx(context.Background()).Fatal(msg, l.Fields...)
}

func (l Logger) NoCtxFatalf(msg string, args ...any) {
	l.Logger.Ctx(context.Background()).Fatal(fmt.Sprintf(msg, args...), l.Fields...)
}
