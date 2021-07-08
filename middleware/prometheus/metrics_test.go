package prometheus

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/require"
	"github.com/taylorchu/work"
)

func TestHandleFuncMetrics(t *testing.T) {
	job := work.NewJob()
	opt := &work.DequeueOptions{
		Namespace: "{ns1}",
		QueueID:   "q1",
	}
	h := HandleFuncMetrics(func(context.Context, *work.Job, *work.DequeueOptions) error {
		return nil
	})

	err := h(context.Background(), job, opt)
	require.NoError(t, err)

	h = HandleFuncMetrics(func(context.Context, *work.Job, *work.DequeueOptions) error {
		return errors.New("no reason")
	})
	err = h(context.Background(), job, opt)
	require.Error(t, err)

	r := httptest.NewRecorder()
	promhttp.Handler().ServeHTTP(r, httptest.NewRequest("GET", "/metrics", nil))

	for _, m := range []string{
		`work_job_executed_total{`,
		`work_job_execution_time_seconds_bucket{`,
		`work_job_busy{`,
	} {
		require.Contains(t, r.Body.String(), m)
	}
}

func TestEnqueueFuncMetrics(t *testing.T) {
	job := work.NewJob()
	opt := &work.EnqueueOptions{
		Namespace: "{ns1}",
		QueueID:   "q1",
	}
	h := EnqueueFuncMetrics(func(context.Context, *work.Job, *work.EnqueueOptions) error {
		return nil
	})

	err := h(context.Background(), job, opt)
	require.NoError(t, err)

	h = EnqueueFuncMetrics(func(context.Context, *work.Job, *work.EnqueueOptions) error {
		return errors.New("no reason")
	})
	err = h(context.Background(), job, opt)
	require.Error(t, err)

	r := httptest.NewRecorder()
	promhttp.Handler().ServeHTTP(r, httptest.NewRequest("GET", "/metrics", nil))

	for _, m := range []string{
		`work_job_enqueued_total{`,
	} {
		require.Contains(t, r.Body.String(), m)
	}
}
