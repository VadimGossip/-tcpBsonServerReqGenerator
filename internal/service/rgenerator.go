package service

import (
	"fmt"
	"github.com/VadimGossip/tcpBsonServerReqGenerator/internal/config"
	"github.com/VadimGossip/tcpBsonServerReqGenerator/internal/domain"
	"github.com/VadimGossip/tcpBsonServerReqGenerator/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"sync"
	"time"
)

type RequestGenerator struct {
	config    config.RGeneratorConfig
	mu        sync.RWMutex
	sent      int
	received  int
	durations []time.Duration
	finished  bool
}

func NewRequestGenerator(cfg config.RGeneratorConfig) *RequestGenerator {
	return &RequestGenerator{config: cfg}
}

func (rg *RequestGenerator) SetSent() {
	rg.mu.Lock()
	defer rg.mu.Unlock()
	rg.sent++
}

func (rg *RequestGenerator) SetReceived() {
	rg.mu.Lock()
	defer rg.mu.Unlock()
	rg.received++
}

func (rg *RequestGenerator) SetFinished() {
	rg.mu.Lock()
	defer rg.mu.Unlock()
	rg.finished = true
}

func (rg *RequestGenerator) SetDuration(duration time.Duration) {
	rg.mu.Lock()
	defer rg.mu.Unlock()
	rg.durations = append(rg.durations, duration)
}

func (rg *RequestGenerator) GetSent() int {
	rg.mu.RLock()
	defer rg.mu.RUnlock()
	return rg.sent
}

func (rg *RequestGenerator) GetReceived() int {
	rg.mu.RLock()
	defer rg.mu.RUnlock()
	return rg.received
}

func (rg *RequestGenerator) GetFinished() bool {
	rg.mu.RLock()
	defer rg.mu.RUnlock()
	return rg.finished
}

func (rg *RequestGenerator) GetDurations() []time.Duration {
	rg.mu.RLock()
	defer rg.mu.RUnlock()
	return rg.durations
}

func (rg *RequestGenerator) GenerateRequestsEndless(reqBytesChan chan<- domain.ByteMsg) {
	var tickerStep time.Duration
	var ttlNum int
	req := domain.RouteRequest{
		RequestType:   "route",
		RouteDuration: time.Millisecond * 100,
	}
	numTick := rg.config.RoutePerSec / 10
	periodSlice := make([]int, 0, 10)
	if numTick == 0 {
		tickerStep = 1 * time.Second
		periodSlice = append(periodSlice, rg.config.RoutePerSec)
	} else {
		tickerStep = 100 * time.Millisecond
		for {
			if ttlNum == rg.config.RoutePerSec {
				break
			}

			if numTick+ttlNum > rg.config.RoutePerSec {
				numTick = rg.config.RoutePerSec - ttlNum
			}
			periodSlice = append(periodSlice, numTick)
			ttlNum += numTick
		}
	}
	ticker := time.NewTicker(tickerStep)
	tickerIndex := 0
	done := make(chan bool)
	counter := 0
	go func() {
		for {
			select {
			case <-done:
				rg.SetFinished()
				return
			case <-ticker.C:
				for i := 0; i < periodSlice[tickerIndex]; i++ {
					req.MsgId = counter + 1
					req.SendTime = time.Now()
					reqBytes, _ := bson.Marshal(&req)
					reqBytesChan <- domain.ByteMsg{MsgBody: reqBytes}
					rg.SetSent()
					counter++
				}
				tickerIndex++
				if tickerIndex == 10 {
					tickerIndex = 0
				}
			}
		}
	}()
	time.Sleep(time.Duration(rg.config.WorkTimeSec) * time.Second)
	ticker.Stop()
	done <- true
}

func (rg *RequestGenerator) analyzeDurations(durations []time.Duration) (map[float64]int, time.Duration, time.Duration, time.Duration) {
	histogram := make(map[float64]int)
	var min, max, sum time.Duration
	for key, val := range durations {
		if key == 0 {
			min = val
			max = val
		} else {
			if val < min {
				min = val
			}
			if val > max {
				max = val
			}
		}
		sum += val
		histogram[(utils.Round(float64(val.Milliseconds()/100), 0)*100)+100]++
	}
	return histogram, sum / time.Duration(len(durations)), min, max
}

func (rg *RequestGenerator) PrintStatReport() {
	fmt.Printf("config %+v\n", rg.config)
	fmt.Printf("Route Request Sent %d\n", rg.GetSent())
	durations := rg.GetDurations()
	h, avg, min, max := rg.analyzeDurations(durations)
	fmt.Printf("Route Histogramm %+v\n", h)
	fmt.Printf("Route Request Avg Answer Duration %+v\n", avg)
	fmt.Printf("Route Request Min Answer Duration %+v\n", min)
	fmt.Printf("Route Request Max Answer Duration %+v\n", max)
}
