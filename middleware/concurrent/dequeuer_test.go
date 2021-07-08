package concurrent

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"github.com/taylorchu/work"
	"github.com/taylorchu/work/redistest"
)

func TestDequeuer(t *testing.T) {
	ctx := context.Background()
	client := redistest.NewClient()
	defer client.Close()
	require.NoError(t, redistest.Reset(client))

	opt := &work.DequeueOptions{
		Namespace:    "{ns1}",
		QueueID:      "q1",
		At:           time.Now(),
		InvisibleSec: 60,
	}
	var called int
	h := func(context.Context, *work.DequeueOptions) (*work.Job, error) {
		called++
		return work.NewJob(), nil
	}

	for i := 0; i < 3; i++ {
		deq := Dequeuer(&DequeuerOptions{
			Client:   client,
			Max:      2,
			workerID: fmt.Sprintf("w%d", i),
		})
		_, err := deq(h)(ctx, opt)

		require.NoError(t, err)
	}
	require.Equal(t, 3, called)

	// worker 0, 1 get the lock
	// worker 2 is locked
	for i := 0; i < 3; i++ {
		deq := Dequeuer(&DequeuerOptions{
			Client:        client,
			Max:           2,
			workerID:      fmt.Sprintf("w%d", i),
			disableUnlock: true,
		})
		_, err := deq(h)(ctx, opt)

		if i <= 1 {
			require.NoError(t, err)
		} else {
			require.Equal(t, work.ErrEmptyQueue, err)
		}
	}
	require.Equal(t, 5, called)

	z, err := client.ZRangeByScoreWithScores(
		context.Background(),
		"{ns1}:lock:q1",
		&redis.ZRangeBy{
			Min: "-inf",
			Max: "+inf",
		}).Result()
	require.NoError(t, err)
	require.Len(t, z, 2)
	require.Equal(t, "w0", z[0].Member)
	require.Equal(t, "w1", z[1].Member)
	require.EqualValues(t, opt.At.Unix()+60, z[0].Score)
	require.EqualValues(t, opt.At.Unix()+60, z[1].Score)

	require.NoError(t, client.ZRem(context.Background(), "{ns1}:lock:q1", "w1").Err())
	optLater := *opt
	optLater.At = opt.At.Add(10 * time.Second)
	// worker 0 is locked already
	for i := 0; i < 3; i++ {
		deq := Dequeuer(&DequeuerOptions{
			Client:        client,
			Max:           2,
			workerID:      "w0",
			disableUnlock: true,
		})
		_, err := deq(h)(ctx, &optLater)
		require.Equal(t, work.ErrEmptyQueue, err)
	}
	require.Equal(t, 5, called)

	z, err = client.ZRangeByScoreWithScores(
		context.Background(),
		"{ns1}:lock:q1",
		&redis.ZRangeBy{
			Min: "-inf",
			Max: "+inf",
		}).Result()
	require.NoError(t, err)
	require.Len(t, z, 1)
	require.Equal(t, "w0", z[0].Member)
	require.EqualValues(t, opt.At.Unix()+60, z[0].Score)

	// lock key expired
	// worker 3, 4 get lock
	optExpired := *opt
	optExpired.At = opt.At.Add(60 * time.Second)
	for i := 3; i < 6; i++ {
		deq := Dequeuer(&DequeuerOptions{
			Client:        client,
			Max:           2,
			workerID:      fmt.Sprintf("w%d", i),
			disableUnlock: true,
		})
		_, err := deq(h)(ctx, &optExpired)
		if i < 5 {
			require.NoError(t, err)
		} else {
			require.Equal(t, work.ErrEmptyQueue, err)
		}
	}
	require.Equal(t, 7, called)

	z, err = client.ZRangeByScoreWithScores(
		context.Background(),
		"{ns1}:lock:q1",
		&redis.ZRangeBy{
			Min: "-inf",
			Max: "+inf",
		}).Result()
	require.NoError(t, err)
	require.Len(t, z, 2)
	require.Equal(t, "w3", z[0].Member)
	require.Equal(t, "w4", z[1].Member)
	require.EqualValues(t, optExpired.At.Unix()+60, z[0].Score)
	require.EqualValues(t, optExpired.At.Unix()+60, z[1].Score)
}

func BenchmarkConcurrency(b *testing.B) {
	ctx := context.Background()
	b.StopTimer()

	client := redistest.NewClient()
	defer client.Close()
	require.NoError(b, redistest.Reset(client))

	opt := &work.DequeueOptions{
		Namespace:    "{ns1}",
		QueueID:      "q1",
		At:           time.Now(),
		InvisibleSec: 60,
	}
	deq := Dequeuer(&DequeuerOptions{
		Client: client,
		Max:    1,
	})
	var called int
	h := deq(func(context.Context, *work.DequeueOptions) (*work.Job, error) {
		called++
		return work.NewJob(), nil
	})

	b.StartTimer()
	for n := 0; n < b.N; n++ {
		h(ctx, opt)
	}
	b.StopTimer()
	require.Equal(b, b.N, called)
}
