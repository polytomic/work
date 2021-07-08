package discard

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/taylorchu/work"
)

func TestMaxRetry(t *testing.T) {
	job := work.NewJob()
	opt := &work.DequeueOptions{
		Namespace: "{ns1}",
		QueueID:   "q1",
	}
	d := MaxRetry(1)
	h := d(func(context.Context, *work.Job, *work.DequeueOptions) error {
		return errors.New("no reason")
	})

	err := h(context.Background(), job, opt)
	require.Error(t, err)
	require.NotEqual(t, work.ErrUnrecoverable, err)

	job.Retries = 1
	err = h(context.Background(), job, opt)
	require.Equal(t, work.ErrUnrecoverable, err)
}
