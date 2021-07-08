package discard

import (
	"context"
	"errors"

	"github.com/taylorchu/work"
)

// InvalidPayload discards a job if it has decode error.
func InvalidPayload(f work.ContextHandleFunc) work.ContextHandleFunc {
	return func(ctx context.Context, job *work.Job, opt *work.DequeueOptions) error {
		err := f(ctx, job, opt)
		if err != nil {
			var perr *work.InvalidJobPayloadError
			if errors.As(err, &perr) {
				return work.ErrUnrecoverable
			}
			return err
		}
		return nil
	}
}

var _ work.HandleMiddleware = InvalidPayload
