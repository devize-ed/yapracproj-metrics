package agent

import "github.com/devize-ed/yapracproj-metrics.git/internal/logger"

// createWorkerPool creates a worker pool.
func (j *jobs) createWorkerPool(request func(name string, endpoint string, bodyBytes []byte) error, numWorkers int) chan error {
	logger.Log.Debug("Creating worker pool with ", numWorkers, " workers")
	errChan := make(chan error, numWorkers)
	// Create a worker pool.
	for i := 0; i < numWorkers; i++ {
		wrkId := i
		// Create a worker.
		j.wg.Add(1)
		go func(jobsQueue <-chan batchRequest, errChan chan<- error, wrkId int) {
			logger.Log.Debug("Worker ", wrkId, " started")
			defer j.wg.Done()
			// Process jobs from the jobs queue.
			for job := range jobsQueue {
				logger.Log.Debug("Worker ", wrkId, " processing job ", job.name, " with endpoint ", job.endpoint)
				// Send a request to the server.
				err := request(job.name, job.endpoint, job.bodyBytes)
				if err != nil {
					// Send an error to the error channel.
					errChan <- err
				}
			}
		}(j.jobsQueue, errChan, wrkId)
	}
	return errChan
}
