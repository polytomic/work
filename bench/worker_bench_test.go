package work

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis"
	work1 "github.com/gocraft/work"
	redigo "github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/require"
	"github.com/taylorchu/work"
)

func BenchmarkWorkerRunJob(b *testing.B) {
	client := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:6379",
		PoolSize:     10,
		MinIdleConns: 10,
	})
	defer client.Close()

	pool := &redigo.Pool{
		MaxActive: 10,
		MaxIdle:   10,
		Dial:      func() (redigo.Conn, error) { return redigo.Dial("tcp", "127.0.0.1:6379") },
	}
	defer pool.Close()

	for k := 1; k <= 100000; k *= 10 {
		b.Run(fmt.Sprintf("work_v1_%d", k), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				b.StopTimer()
				require.NoError(b, client.FlushAll().Err())

				wp := work1.NewWorkerPoolWithOptions(
					struct{}{}, 1, "ns1", pool,
					work1.WorkerPoolOptions{
						SleepBackoffs: []int64{1000},
					},
				)

				var wg sync.WaitGroup
				wp.Job("test", func(job *work1.Job) error {
					wg.Done()
					return nil
				})

				enqueuer := work1.NewEnqueuer("ns1", pool)
				for i := 0; i < k; i++ {
					_, err := enqueuer.Enqueue("test", nil)
					require.NoError(b, err)

					wg.Add(1)
				}

				b.StartTimer()
				wp.Start()
				wg.Wait()
				wp.Stop()
			}
		})
		b.Run(fmt.Sprintf("work_v2_%d", k), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				b.StopTimer()
				require.NoError(b, client.FlushAll().Err())

				queue := work.NewRedisQueue(client)
				w := work.NewWorker(&work.WorkerOptions{
					Namespace: "ns1",
					Queue:     queue,
				})
				var wg sync.WaitGroup
				err := w.Register("test",
					func(*work.Job, *work.DequeueOptions) error {
						wg.Done()
						return nil
					},
					&work.JobOptions{
						MaxExecutionTime: time.Minute,
						IdleWait:         time.Second,
						NumGoroutines:    1,
					},
				)
				require.NoError(b, err)

				for i := 0; i < k; i++ {
					job := work.NewJob()

					err := queue.Enqueue(job, &work.EnqueueOptions{
						Namespace: "ns1",
						QueueID:   "test",
					})
					require.NoError(b, err)

					wg.Add(1)
				}

				b.StartTimer()
				w.Start()
				wg.Wait()
				w.Stop()
			}
		})
	}
}
