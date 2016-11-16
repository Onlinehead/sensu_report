package main

import (
	"log"
	"time"
	"github.com/onlinehead/sensu_report"
)

func main() {
	/*
		Send a single message to sensu
	*/
	// Sensu param
	sensu := SensuReport.ClientSettings{
		URI:  "127.0.0.1:3000",
		Noop: false,
	}
	// Prepare message for send
	sensu_message := SensuReport.Message{
		Status: 0,
		Text:   "test",
		Name:   "test_alert",
	}
	// Send single message
	sensu.SendMessage(sensu_message)

	/*
		Run SensuReporter in goroutine and send cumulative message every 10 seconds
	*/

	// Reporter settings
	sensu_reporter := SensuReport.Reporter{
		ClientSettings: sensu,
		CumulativeTime: 10,
		MaxStatus: 2,
	}
	// Create channel
	msgs := make(chan SensuReport.Message)

	log.Print("Launch reports routine")
	go sensu_reporter.Start(msgs)

	// Add 20 messages
	for i := 0; i < 20; i++ {
		message := SensuReport.Message{
			Status: 0,
			Name:   "test_alert_cumulative",
		}
		msgs <- message
	}
	for i := 0; i < 20; i++ {
		message := SensuReport.Message{
			Status: 1,
			Name:   "test_alert_cumulative",
		}
		msgs <- message
	}
	// Just wait
	time.Sleep(30 * time.Minute)
}
