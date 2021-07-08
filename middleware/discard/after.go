package discard

import (
	"context"
	"time"

	"github.com/taylorchu/work"
)

// After discards a job if it is already stale.
func After(d time.Duration) work.HandleMiddleware {
	return func(f work.ContextHandleFunc) work.ContextHandleFunc {
		return func(ctx context.Context, job *work.Job, opt *work.DequeueOptions) error {
			if time.Since(job.CreatedAt) > d {
				return work.ErrUnrecoverable
			}
			err := f(ctx, job, opt)
			if time.Since(job.CreatedAt) > d {
				return work.ErrUnrecoverable
			}
			return err
		}
	}
}
