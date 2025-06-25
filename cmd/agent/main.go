package main

import (
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/devize-ed/yapracproj-metrics.git/internal/agent"
)

const (
	host           = "localhost:8080"
	popollInterval = 2
	reportInterval = 10
)

var (
	timer = 0
)

func main() {

	mStorage := &agent.AgentStorage{}
	client := &http.Client{}

	for {
		mStorage.CollectMetrics()
		time.Sleep(popollInterval * time.Second)
		timer += popollInterval
		if timer == reportInterval {
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
