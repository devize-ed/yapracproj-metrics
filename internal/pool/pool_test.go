package pool_test

import (
	"testing"

	"github.com/devize-ed/yapracproj-metrics.git/internal/pool"
	"github.com/stretchr/testify/assert"
)

type testJob struct {
	ID   int
	Data []byte
}

func (j *testJob) Reset() {
	j.ID = 0
	j.Data = j.Data[:0]
}

func TestNewPool(t *testing.T) {
	testCases := []struct {
		name    string
		jobPool *pool.Pool[*testJob]
	}{
		{
			name: "test_job",
			jobPool: pool.NewPool(func() *testJob {
				return &testJob{Data: make([]byte, 0, 2048)}
			}),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc.jobPool, "pool should not be nil")
		})
	}
}

func TestGet(t *testing.T) {
	pool := pool.NewPool(func() *testJob {
		return &testJob{Data: make([]byte, 0, 2048)}
	})
	j := pool.Get()
	assert.NotNil(t, j, "job should not be nil")
	assert.Equal(t, 0, j.ID, "job ID should be 0")
	assert.Equal(t, []byte{}, j.Data, "job Data should be empty")
}

func TestPut(t *testing.T) {
	pool := pool.NewPool(func() *testJob {
		return &testJob{Data: make([]byte, 0, 2048)}
	})
	j := pool.Get()
	defer pool.Put(j)
	assert.NotNil(t, j, "job should not be nil")
	assert.Equal(t, 0, j.ID, "job ID should be 0")
	assert.Equal(t, []byte{}, j.Data, "job Data should be empty")
}
