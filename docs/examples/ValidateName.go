package examples

import "errors"

var ErrTooLong = errors.New(`validate error: too long`)

func ValidateName(name string) error {
	if len(name) > 128 {
		return ErrTooLong
	}

	return nil
}
