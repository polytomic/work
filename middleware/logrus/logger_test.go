package logrus

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/taylorchu/work"
)

func TestHandleFuncLogger(t *testing.T) {
	job := work.NewJob()
	opt := &work.DequeueOptions{
		Namespace: "{ns1}",
		QueueID:   "q1",
	}
	h := HandleFuncLogger(func(context.Context, *work.Job, *work.DequeueOptions) error {
		return nil
	})

	err := h(context.Background(), job, opt)
	require.NoError(t, err)

	h = HandleFuncLogger(func(context.Context, *work.Job, *work.DequeueOptions) error {
		return errors.New("no reason")
	})
	err = h(context.Background(), job, opt)
	require.Error(t, err)
}

func TestEnqueueFuncLogger(t *testing.T) {
	job := work.NewJob()
	opt := &work.EnqueueOptions{
		Namespace: "{ns1}",
		QueueID:   "q1",
	}
	h := EnqueueFuncLogger(func(context.Context, *work.Job, *work.EnqueueOptions) error {
		return nil
	})

	err := h(context.Background(), job, opt)
	require.NoError(t, err)

	h = EnqueueFuncLogger(func(context.Context, *work.Job, *work.EnqueueOptions) error {
		return errors.New("no reason")
	})
	err = h(context.Background(), job, opt)
	require.Error(t, err)
}
