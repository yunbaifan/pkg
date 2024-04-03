package plugins

import (
	"context"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var (
	_                  gorm.Plugin = &TracingPlugin{}
	spanName                       = "gorm:zero"
	gormSpanKey                    = "gorm:zero:span"
	callBackBeforeName             = "gorm:zero:trace:before"
	callBackAfterName              = "gorm:zero:trace:after"

	spanEventName        = "gorm-zero-event"
	spanAttrTable        = attribute.Key("gorm.table")
	spanAttrSql          = attribute.Key("gorm.sql")
	spanAttrRowsAffected = attribute.Key("gorm.rowsAffected")
)

type gormRegister interface {
	Register(name string, fn func(*gorm.DB)) error
}
type gormHookFunc func(tx *gorm.DB)

type TracingPlugin struct{}

func (gp *TracingPlugin) Name() string {
	return "gorm-zero-tracing-plugin"
}

func (gp *TracingPlugin) Initialize(db *gorm.DB) error {
	cb := db.Callback()
	// 注册钩子函数
	hooks := []struct {
		callback gormRegister
		hook     gormHookFunc
		name     string
	}{
		{callback: cb.Create().Before("gorm:before_create"), hook: gp.before(), name: callBackBeforeName},
		{callback: cb.Query().Before("gorm:before_query"), hook: gp.before(), name: callBackAfterName},
		{callback: cb.Update().Before("gorm:before_update"), hook: gp.before(), name: callBackBeforeName},
		{callback: cb.Delete().Before("gorm:before_delete"), hook: gp.before(), name: callBackBeforeName},
		{callback: cb.Row().Before("gorm:before_row"), hook: gp.before(), name: callBackBeforeName},
		{callback: cb.Raw().Before("gorm:before_raw"), hook: gp.before(), name: callBackBeforeName},

		{callback: cb.Create().After("gorm:after_create"), hook: gp.after(), name: callBackAfterName},
		{callback: cb.Query().After("gorm:after_query"), hook: gp.after(), name: callBackAfterName},
		{callback: cb.Update().After("gorm:after_update"), hook: gp.after(), name: callBackAfterName},
		{callback: cb.Delete().After("gorm:after_delete"), hook: gp.after(), name: callBackAfterName},
		{callback: cb.Row().After("gorm:after_row"), hook: gp.after(), name: callBackAfterName},
		{callback: cb.Raw().After("gorm:after_raw"), hook: gp.after(), name: callBackAfterName},
	}
	var firstErr error
	for _, h := range hooks {
		if err := h.callback.Register("otel:"+h.name, h.hook); err != nil && firstErr == nil {
			err = fmt.Errorf("callback register %s failed: %w", h.name, err)
		}
	}
	return firstErr
}

func (gp *TracingPlugin) before() gormHookFunc {
	return func(tx *gorm.DB) {
		_, span := gp.startSpan(tx.Statement.Context)
		// 利用db实例去传递span
		tx.InstanceSet(gormSpanKey, span)
	}
}

func (gp *TracingPlugin) startSpan(ctx context.Context) (context.Context, oteltrace.Span) {
	tracer := trace.TracerFromContext(ctx)
	start, span := tracer.Start(ctx,
		spanName,
		oteltrace.WithSpanKind(oteltrace.SpanKindClient),
	)
	return start, span
}

func (gp *TracingPlugin) after() gormHookFunc {
	return func(tx *gorm.DB) {
		_span, isExist := tx.InstanceGet(gormSpanKey)
		if !isExist {
			return
		}
		// 断言
		span, ok := _span.(oteltrace.Span)
		if !ok {
			return
		}
		defer func() {
			gp.endSpan(span, tx.Error)
		}()
		span.AddEvent(spanEventName, oteltrace.WithAttributes(
			spanAttrTable.String(tx.Statement.Table),
			spanAttrRowsAffected.Int64(tx.RowsAffected),
			spanAttrSql.String(tx.Dialector.Explain(tx.Statement.SQL.String(), tx.Statement.Vars...)),
		))
	}
}

func (gp *TracingPlugin) endSpan(span oteltrace.Span, err error) {
	defer span.End()
	if err == nil && errors.Is(err, gorm.ErrRecordNotFound) {
		span.SetStatus(codes.Ok, "")
		return
	}
	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err)
}
