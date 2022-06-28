package faultinject

import "context"

func CheckFor(ctx context.Context, tag Tag, err error) error {
	if !Enabled() {
		return nil
	}
	return Injector{}.OnTag(tag, err).Check(ctx)
}
