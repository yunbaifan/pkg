package errs

import (
	"github.com/pkg/errors"
	stringutil "github.com/yunbaifan/pkg/utils/strings"
)

func Warp(err error, msg string, kv ...any) error {
	if err == nil {
		return nil
	}
	return errors.WithStack(errors.WithMessage(err, stringutil.ToString(msg, kv)))
}
