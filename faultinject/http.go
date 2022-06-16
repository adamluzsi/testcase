package faultinject

//import (
//	"context"
//	"net"
//	"net/http"
//	"syscall"
//)
//
//type RoundTripper struct{ Next http.RoundTripper }
//
//const TagRoundTripper = "faultinject.RoundTripper"
//
//func InjectNetTimeoutError(ctx context.Context) context.Context {
//	return Inject(ctx, fault{
//		Tag: TagRoundTripper,
//		Error: netTimeoutError{},
//	})
//}
//
//func InjectNetConnectionRefusedError(ctx context.Context, intn func(int) int) context.Context {
//
//	switch t := err.(type) {
//	case *net.OpError:
//		if t.Op == "dial" {
//			println("Unknown host")
//		} else if t.Op == "read" {
//			println("Connection refused")
//		}
//
//	case syscall.Errno:
//		if t == syscall.ECONNREFUSED {
//			println("Connection refused")
//		}
//	}
//
//	return Inject(ctx, fault{
//		Tag: TagRoundTripper,
//		Error: syscall.ECONNREFUSED,
//	})
//}
//
//func (rt RoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
//	if err, _ := Check(r.Context(), TagRoundTripper); err != nil {
//		return nil, err
//	}
//
//	if err == nil {
//		println("Ok")
//		return
//
//	} else if netError, ok := err.(net.Error); ok && netError.Timeout() {
//		println("Timeout")
//		return
//	}
//
//	switch t := err.(type) {
//	case *net.OpError:
//		if t.Op == "dial" {
//			println("Unknown host")
//		} else if t.Op == "read" {
//			println("Connection refused")
//		}
//
//	case syscall.Errno:
//		if t == syscall.ECONNREFUSED {
//			println("Connection refused")
//		}
//	}
//
//	return rt.Next.RoundTrip(r)
//}
//
//type netTimeoutError struct{}
//
//func (e netTimeoutError) Error() string   { return "i/o timeout" }
//func (e netTimeoutError) Timeout() bool   { return true }
//func (e netTimeoutError) Temporary() bool { return true }
