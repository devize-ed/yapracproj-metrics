package main

import (
	"fmt"
	"reflect"
	"time"

	"github.com/devize-ed/yapracproj-metrics.git/internal/agent"
	"github.com/go-resty/resty/v2"
)

var (
	timer = 0 * time.Second
)

func main() {
	parseFlags()                      // call the function to parse cl flags
	mStorage := &agent.AgentStorage{} // init agent storage for metrics
	client := resty.New()             // init client

	// convert the interval values to time.Duration
	timePoolInterval := time.Duration(pollInterval) * time.Second
	timeReportInterval := time.Duration(reportInterval) * time.Second

	// start agent loop
	for {
		mStorage.CollectMetrics()        // collect metrics
		time.Sleep(timePoolInterval)     // wait for collecting interval
		timer += timePoolInterval        // increment the timer by
		if timer <= timeReportInterval { // if timer is less than reportInterval, report metrics
			fmt.Println("timer == reportInterval, reporting metrics...")

			// iterate over the agent storage and send metrics to the server
			val := reflect.ValueOf(mStorage).Elem()
			typ := reflect.TypeOf(mStorage).Elem()

			for i := 0; i < val.NumField(); i++ {
				metric := typ.Field(i).Name
				value := val.Field(i)
				// fmt.Printf("%s = %v\n", metric, value)
				agent.SendMetric(client, metric, fmt.Sprint(value), host)
			}
			timer = 0 // reset the timer
		}
	}
}
