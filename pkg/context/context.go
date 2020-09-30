package context

import (
	"context"
	gocontext "context"

	"k8s.io/klog"
)

// CanclableContext a cancleable context
type CanclableContext interface {
	gocontext.Context
	Cancle(error)
	Derive() CanclableContext
	Cleanup(func() error)
}

type cancleContext struct {
	gocontext.Context
	context.CancelFunc
}

// NewCanclableContext create a cancleable context with optional nil error collector
func NewCanclableContext(ctx gocontext.Context) CanclableContext {
	cancleCtx, cancler := gocontext.WithCancel(ctx)
	return cancleContext{
		cancleCtx,
		cancler,
	}
}

func (ctx cancleContext) Cancle(reason error) {
	if reason != nil {
		klog.V(6).Infof("cancel context:%v \nreason:%v", ctx, reason)
	}

	ctx.CancelFunc()
}

func (ctx cancleContext) Derive() CanclableContext {
	return NewCanclableContext(ctx)
}

func (ctx cancleContext) Cleanup(cleanup func() error) {
	go func() {
		<-ctx.Done()
		if err := cleanup(); err != nil {
			klog.Errorf("fail cleanup context:%v \nreason:%v", ctx, err)
		}
	}()
}
