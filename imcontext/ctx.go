package imcontext

import "context"

var (
	mapper = []struct {
	}{
		Operation{},
		OpUserID{},
		OpUserPlatform{},
		ConnID{},
		RemoteAddr{},
		TriggerID{},
	}
)

type Operation struct {
}

type OpUserID struct {
}

type OpUserPlatform struct {
}

type ConnID struct{}

type RemoteAddr struct{}
type TriggerID struct{}

func WithMustInfoCtx(ctx context.Context, value []struct{}) context.Context {
	nCtx := ctx
	for i, v := range value {
		nCtx = context.WithValue(nCtx, mapper[i], v)
	}
	return nCtx
}

func WithOpUserID(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, OpUserID{}, value)
}

func WithOpUserPlatform(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, OpUserPlatform{}, value)
}

func WithConnID(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, ConnID{}, value)
}

func GetOperation(ctx context.Context) string {
	if v := ctx.Value(Operation{}); v != nil {
		return v.(string)
	}
	return ""
}

func GetOpUserID(ctx context.Context) string {
	if v := ctx.Value(OpUserID{}); v != nil {
		return v.(string)
	}
	return ""
}

func GetConnID(ctx context.Context) string {
	if v := ctx.Value(ConnID{}); v != nil {
		return v.(string)
	}
	return ""
}

func GetTriggerID(ctx context.Context) string {
	if v := ctx.Value(TriggerID{}); v != nil {
		return v.(string)
	}
	return ""
}

func GetOpUserPlatform(ctx context.Context) string {
	if v := ctx.Value(OpUserPlatform{}); v != nil {
		return v.(string)
	}
	return ""
}

func GetRemoteAddr(ctx context.Context) string {
	if v := ctx.Value(RemoteAddr{}); v != nil {
		return v.(string)
	}
	return ""
}
