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
	parseFlags()
	mStorage := &agent.AgentStorage{}
	client := resty.New()

	for {
		mStorage.CollectMetrics()
		time.Sleep(popollInterval)
		timer += popollInterval
		if timer <= reportInterval {
			fmt.Println("timer == reportInterval, reporting metrics...")
			val := reflect.ValueOf(mStorage).Elem()
			typ := reflect.TypeOf(mStorage).Elem()

			for i := 0; i < val.NumField(); i++ {
				metric := typ.Field(i).Name
				value := val.Field(i)
				// fmt.Printf("%s = %v\n", metric, value)
				agent.SendMetric(client, metric, fmt.Sprint(value), host)
			}
			// fmt.Println("==========================================")
			timer = 0
		}
	}
}
