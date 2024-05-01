package bonsai

import (
	"errors"
	"fmt"
	"io"
)

// IoClose will catch io.Closer errors and wrap them around the
// previous errors, if any.
//
// Note: due to the Go spec, in order to actually modify the parent's
// error value, the calling function *must* use named result parameters.
//
// ref: https://go.dev/ref/spec#Defer_statements
func IoClose(c io.Closer, err error) error {
	cerr := c.Close()

	if cerr != nil {
		return errors.Join(
			fmt.Errorf("failed to close io.Closer: %w", cerr),
			err,
		)
	}
	return err
}
