package SensuReport

import (
	"sync"
	"fmt"
	"strconv"
	"log"
	"time"
	"net"
	"encoding/json"
)

type SensuMessage struct {
	Text string `json:"output" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Status  int    `json:"status" binding:"required"`
}

type SensuReportSettings struct {
	SensuSettings
	CumulativeTime int
	MaxStatus int
}

type SensuSettings struct {
	URI string // As example, default "127.0.0.1:3000"
	Noop bool // Test run, not really send messages
}


//Get channel with sensu messages and wait time cumulative messages
func (sensu SensuReportSettings) SensuReporter(reports <-chan SensuMessage) chan int {
	if sensu.CumulativeTime < 10 {
		log.Println("CummulativeTime must be >= 10, set default 10")
		sensu.CumulativeTime = 10
	}
	var mu sync.Mutex
	// Create map with messages
	var reports_cache []SensuMessage
	// Collect messages to cache
	go func () {
		for report := range reports {
				if report.Status > sensu.MaxStatus {
					report.Status = sensu.MaxStatus
				}
				mu.Lock()
				reports_cache = append(reports_cache, report)
				mu.Unlock()
			}
	}()
	// Process messages cache each 30 seconds
	for {
		type MessageGroup struct {
			MaxStatus int
			// Numbers of messages with each code in format [code]number_of_messages
			Results  map[int]int
		}
		// Group messages by name
		rep := make(map[string]MessageGroup)
		// Create working copy and clear cache with locking
		mu.Lock()
		reserved_messages := reports_cache
		reports_cache = nil
		mu.Unlock()
		for s := range reserved_messages {
			if val, ok := rep[reserved_messages[s].Name]; ok {
				// Update MaxStatus if needed
				if val.MaxStatus < reserved_messages[s].Status {
						val.MaxStatus = reserved_messages[s].Status
				}
				// Update number of reports with that code
				if _, code_ok := val.Results[reserved_messages[s].Status]; code_ok {
					val.Results[reserved_messages[s].Status] += 1
				} else {
					val.Results[reserved_messages[s].Status] = 1
				}
			} else {
				res := make(map[int]int)
				res[reserved_messages[s].Status] = 1
				m := MessageGroup{
					MaxStatus: reserved_messages[s].Status,
					Results: res,
				}
				rep[reserved_messages[s].Name] = m
			}
		}
		// Send report to Sensu, one per name
		for i := range rep {
			var text string
			for c := range rep[i].Results {
				text += fmt.Sprintf("We have %s results with status %s", strconv.Itoa(rep[i].Results[c]), strconv.Itoa(c))
			}
			smess := SensuMessage {
				Status: rep[i].MaxStatus,
				Name: i,
				Text: text,
			}
			log.Printf("Message to Sensu with name '%s' sent", i)
			sensu.SensuSendMessage(smess)
		}
		time.Sleep(time.Duration(sensu.CumulativeTime)*time.Second)
	}
}
// Standalone sender
func (sensu SensuSettings) SensuSendMessage(message SensuMessage) {
	if sensu.Noop == false {
		conn, err := net.Dial("udp", sensu.URI)
		if err != nil {
			log.Println("Cannot open connection to Sensu with err:", err)
			return
		}

		mes, err := json.Marshal(message)
		if err != nil {
			log.Println("Cannot prepare message for Sensu with err:", err)
			return
		}
		conn.Write(mes)
	}
}