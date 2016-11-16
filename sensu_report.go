package SensuReport

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

type Message struct {
	Text   string `json:"output" binding:"required"`
	Name   string `json:"name" binding:"required"`
	Status int    `json:"status" binding:"required"`
}

type Reporter struct {
	ClientSettings
	CumulativeTime int
	MaxStatus      int
}

type ClientSettings struct {
	URI  string // As example, default "127.0.0.1:3000"
	Noop bool   // Test run, not really send messages
}

//Get channel with sensu messages and wait time cumulative messages
func (sensu Reporter) Start(reports <-chan Message) chan int {
	if sensu.CumulativeTime < 10 {
		log.Println("CummulativeTime must be >= 10, set default 10")
		sensu.CumulativeTime = 10
	}
	var mu sync.Mutex
	// Create map with messages
	var reports_cache []Message
	// Collect messages to cache
	go func() {
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
		// Group messages by name
		rep := make(map[string]map[int]int)
		// Create working copy and clear cache with locking
		mu.Lock()
		reserved_messages := reports_cache
		reports_cache = nil
		mu.Unlock()
		for s := range reserved_messages {
			if val, ok := rep[reserved_messages[s].Name]; ok {
				// Update number of reports with that code
				if _, code_ok := val[reserved_messages[s].Status]; code_ok {
					val[reserved_messages[s].Status] += 1
				} else {
					val[reserved_messages[s].Status] = 1
				}
			} else {
				res := make(map[int]int)
				res[reserved_messages[s].Status] = 1
				rep[reserved_messages[s].Name] = res
			}
		}
		// Send report to Sensu, one per name
		for i := range rep {
			text := "We have:\n"
			var max_status int
			for c := range rep[i] {
				text += fmt.Sprintf("%s results with status %s;\n", strconv.Itoa(rep[i][c]), strconv.Itoa(c))
				if c > max_status {
					max_status = c
				}
			}
			smess := Message{
				Status: max_status,
				Name:   i,
				Text:   text,
			}
			log.Printf("Message to Sensu with name '%s' sent", i)
			sensu.SendMessage(smess)
		}
		time.Sleep(time.Duration(sensu.CumulativeTime) * time.Second)
	}
}

// Standalone sender
func (sensu ClientSettings) SendMessage(message Message) {
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
