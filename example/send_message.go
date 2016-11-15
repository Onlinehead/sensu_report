package main

import (
	"github.com/onlinehead/sensu_report"
	"log"
	"time"
)

func main() {
	/*
		Send a single message to sensu
	*/
	// Sensu param
	sensu_settings := SensuReport.SensuSettings{
		URI:  "127.0.0.1:3000",
		Noop: false,
	}
	// Prepare message for send
	sensu_message := SensuReport.SensuMessage{
		Status: 0,
		Text:   "test",
		Name:   "test_alert",
	}
	// Send single message
	sensu_settings.SensuSendMessage(sensu_message)

	/*
		Run SensuReporter in goroutine and send cumulative message every 10 seconds
	*/

	// Reporter settings
	var sensu_report_settings SensuReport.SensuReportSettings
	sensu_report_settings.SensuSettings = sensu_settings
	sensu_report_settings.CumulativeTime = 10
	sensu_report_settings.MaxStatus = 2

	// Create channel
	testchan := make(chan SensuReport.SensuMessage)

	log.Print("Launch reports routine")
	go sensu_report_settings.SensuReporter(testchan)

	// Add 20 messages
	for i := 0; i < 20; i++ {
		message := SensuReport.SensuMessage{
			Status: 0,
			Name:   "test_alert_cumulative",
		}
		testchan <- message
	}
	// Just wait
	time.Sleep(30 * time.Minute)
}
