package sidekiq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/taylorchu/work"
)

func (q *sidekiqQueue) ExternalEnqueue(ctx context.Context, job *work.Job, opt *work.EnqueueOptions) error {
	return q.ExternalBulkEnqueue(ctx, []*work.Job{job}, opt)
}

func (q *sidekiqQueue) ExternalBulkEnqueue(ctx context.Context, jobs []*work.Job, opt *work.EnqueueOptions) error {
	now := time.Now()
	readyJobs := make([]*work.Job, 0, len(jobs))
	scheduledJobs := make([]*work.Job, 0, len(jobs))
	for _, job := range jobs {
		if job.EnqueuedAt.After(now) {
			scheduledJobs = append(scheduledJobs, job)
		} else {
			readyJobs = append(readyJobs, job)
		}
	}

	err := q.externalBulkEnqueue(ctx, readyJobs, opt)
	if err != nil {
		return err
	}
	err = q.externalBulkEnqueueIn(ctx, scheduledJobs, opt)
	if err != nil {
		return err
	}
	return nil
}

func (q *sidekiqQueue) externalBulkEnqueue(ctx context.Context, jobs []*work.Job, opt *work.EnqueueOptions) error {
	if len(jobs) == 0 {
		return nil
	}
	sqQueue, sqClass, err := ParseQueueID(opt.QueueID)
	if err != nil {
		return err
	}
	args := make([]interface{}, 2+len(jobs))
	args[0] = opt.Namespace
	args[1] = sqQueue
	for i, job := range jobs {
		sqJob, err := newSidekiqJob(job, sqQueue, sqClass)
		if err != nil {
			return err
		}
		err = sqJob.Validate()
		if err != nil {
			return err
		}
		jobm, err := json.Marshal(sqJob)
		if err != nil {
			return err
		}
		args[2+i] = jobm
	}
	return q.enqueueScript.Run(ctx, q.client, nil, args...).Err()
}

func (q *sidekiqQueue) externalBulkEnqueueIn(ctx context.Context, jobs []*work.Job, opt *work.EnqueueOptions) error {
	if len(jobs) == 0 {
		return nil
	}
	sqQueue, sqClass, err := ParseQueueID(opt.QueueID)
	if err != nil {
		return err
	}
	args := make([]interface{}, 1+2*len(jobs))
	args[0] = opt.Namespace
	for i, job := range jobs {
		sqJob, err := newSidekiqJob(job, sqQueue, sqClass)
		if err != nil {
			return err
		}
		err = sqJob.Validate()
		if err != nil {
			return err
		}
		jobm, err := json.Marshal(sqJob)
		if err != nil {
			return err
		}
		args[1+2*i] = job.EnqueuedAt.Unix()
		args[1+2*i+1] = jobm
	}
	return q.enqueueInScript.Run(ctx, q.client, nil, args...).Err()
}

var (
	_ work.ExternalEnqueuer     = (*sidekiqQueue)(nil)
	_ work.ExternalBulkEnqueuer = (*sidekiqQueue)(nil)
)
