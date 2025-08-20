package agent

import "go.uber.org/zap"

// createWorkerPool creates a worker pool.
func (j *jobs) createWorkerPool(request func(name string, endpoint string, bodyBytes []byte) error, numWorkers int, logger *zap.SugaredLogger) chan error {
	logger.Debug("Creating worker pool with ", numWorkers, " workers")
	errChan := make(chan error, numWorkers)
	// Create a worker pool.
	for i := 0; i < numWorkers; i++ {
		wrkId := i
		// Create a worker.
		j.wg.Add(1)
		go func(jobsQueue <-chan batchRequest, errChan chan<- error, wrkId int) {
			logger.Debug("Worker ", wrkId, " started")
			defer j.wg.Done()
			// Process jobs from the jobs queue.
			for job := range jobsQueue {
				logger.Debug("Worker ", wrkId, " processing job ", job.name, " with endpoint ", job.endpoint)
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
