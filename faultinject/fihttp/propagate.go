package fihttp

import "context"

const Header = `Fault-Inject`

type Fault struct {
	ServiceName string `json:"service_name,omitempty"`
	Name        string `json:"name"`
}

type propagateCtxKey struct{}

func Propagate(ctx context.Context, fs ...Fault) context.Context {
	if cfs, ok := lookupFaults(ctx); ok {
		*cfs = append(*cfs, fs...)
		return ctx
	}
	return context.WithValue(ctx, propagateCtxKey{}, &fs)
}

func lookupFaults(ctx context.Context) (*[]Fault, bool) {
	value := ctx.Value(propagateCtxKey{})
	faultsPtr, ok := value.(*[]Fault)
	return faultsPtr, ok
}
