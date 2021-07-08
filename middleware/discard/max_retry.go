package discard

import (
	"context"

	"github.com/taylorchu/work"
)

// MaxRetry discards a job if its retry count is over limit.
func MaxRetry(n int64) work.HandleMiddleware {
	return func(f work.ContextHandleFunc) work.ContextHandleFunc {
		return func(ctx context.Context, job *work.Job, opt *work.DequeueOptions) error {
			err := f(ctx, job, opt)
			if job.Retries >= n {
				return work.ErrUnrecoverable
			}
			return err
		}
	}
}
