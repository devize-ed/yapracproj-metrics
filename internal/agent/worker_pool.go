package agent

// createWorkerPool creates a worker pool.
func (a *Agent) createWorkerPool(numWorkers int) chan error {
	errChan := make(chan error, numWorkers)
	// Create a worker pool.
	for i := 0; i < numWorkers; i++ {
		// Create a worker.
		a.wg.Add(1)
		go func(jobsQueue <-chan batchRequest, errChan chan<- error) {
			defer a.wg.Done()
			// Process jobs from the jobs queue.
			for job := range jobsQueue {
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
