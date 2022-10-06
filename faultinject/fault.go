package faultinject

import "context"

func Check(ctx context.Context, fault any) error {
	if ctx == nil {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err, ok := ctx.Value(fault).(error); ok {
		return err
	}
	return nil
}

// Finish is function that can be called from a deferred context,
// and will inject fault when a function finished its execution.
// The error pointer should point to the function's named return error variable.
// If the function encountered an actual error, fault injection is skipped.
func Finish(returnErr *error, ctx context.Context, faults ...any) {
	if ctx == nil {
		return
	}
	if *returnErr != nil {
		return
	}
	if err := ctx.Err(); err != nil {
		*returnErr = err
		return
	}
	for _, fault := range faults {
		if err := Check(ctx, fault); err != nil {
			*returnErr = err
			return
		}
	}
}
