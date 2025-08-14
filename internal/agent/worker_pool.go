package agent

import "github.com/devize-ed/yapracproj-metrics.git/internal/logger"

// createWorkerPool creates a worker pool.
func (a *Agent) createWorkerPool(numWorkers int) chan error {
	logger.Log.Debug("Creating worker pool with ", numWorkers, " workers")
	errChan := make(chan error, numWorkers)
	// Create a worker pool.
	for i := 0; i < numWorkers; i++ {
		// Create a worker.
		a.wg.Add(1)
		go func(jobsQueue <-chan batchRequest, errChan chan<- error) {
			logger.Log.Debug("Worker ", i, " started")
			defer a.wg.Done()
			// Process jobs from the jobs queue.
			for job := range jobsQueue {
				logger.Log.Debug("Worker ", i, " processing job ", job.name, " with endpoint ", job.endpoint)
				// Send a request to the server.
				err := a.Request(job.name, job.endpoint, job.bodyBytes)
				if err != nil {
					// Send an error to the error channel.
					errChan <- err
				}
			}
		}(a.jobsQueue, errChan)
	}
	return errChan
}
