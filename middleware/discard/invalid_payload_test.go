package discard

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/taylorchu/work"
)

func TestInvalidPayload(t *testing.T) {
	job := work.NewJob()
	opt := &work.DequeueOptions{
		Namespace: "{ns1}",
		QueueID:   "q1",
	}
	h := InvalidPayload(func(context.Context, *work.Job, *work.DequeueOptions) error {
		var s string
		return job.UnmarshalPayload(&s)
	})

	err := h(context.Background(), job, opt)
	require.Error(t, err)
	require.Equal(t, work.ErrUnrecoverable, err)
}
