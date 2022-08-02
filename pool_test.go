package workerpool

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPoolStartAndWait(t *testing.T) {
	type args[T any] struct {
		jobs []JobFn[T]
	}

	tests := []struct {
		name       string
		poolInitFn func() *Pool[int]
		args       args[int]
		want       []int
	}{
		{
			name: "new pool, add jobs, wait and assert result",
			poolInitFn: func() *Pool[int] {
				return New[int](SetWorkers(2))
			},
			args: args[int]{
				jobs: []JobFn[int]{
					func() (int, error) { return 1, nil },
					func() (int, error) { return 2, nil },
					func() (int, error) { return 3, nil },
				},
			},
			want: []int{1, 2, 3},
		},
		{
			name: "new pool with small jobs buffer, add jobs, wait and assert result",
			poolInitFn: func() *Pool[int] {
				return New[int](SetWorkers(2), SetJobsBuffer(1))
			},

			args: args[int]{
				jobs: []JobFn[int]{
					func() (int, error) { return 1, nil },
					func() (int, error) { return 2, nil },
					func() (int, error) { return 3, nil },
					func() (int, error) { return 4, nil },
					func() (int, error) { return 5, nil },
					func() (int, error) { return 6, nil },
				},
			},
			want: []int{1, 2, 3, 4, 5, 6},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := tt.poolInitFn()
			ctx := context.Background()
			pool.Start(ctx)

			got := []int{}
			go func() {
				for _, job := range tt.args.jobs {
					assert.NoError(t, pool.Add(ctx, job))
				}
				pool.Wait(ctx)
			}()

			resultCh, err := pool.Results()
			assert.NoError(t, err)

			for result := range resultCh {
				assert.NoError(t, result.Err)
				got = append(got, result.Result)
			}

			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
