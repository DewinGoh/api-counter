package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"time"
)

// Payload is
type Payload struct {
	testID string
}

// ReqInfo is
type ReqInfo struct {
	srv     *string
	payload url.Values
}

// Response is
type Response struct {
	*http.Response
	err error
}

// Dispatcher
func dispatcher(reqChan chan *ReqInfo, numCalls *int, srv *string, payload url.Values) {
	defer close(reqChan)
	reqInfo := &ReqInfo{srv, payload}
	for i := 0; i < *numCalls; i++ {
		reqChan <- reqInfo
	}
}

// Worker Pool
func workerPool(reqChan chan *ReqInfo, respChan chan Response, numWorkers int) {
	c := http.Client{
		Timeout:   time.Duration(5) * time.Second,
		Transport: &http.Transport{MaxConnsPerHost: 50},
	}
	for i := 0; i < numWorkers; i++ {
		go worker(c, reqChan, respChan)
	}
}

// Worker
func worker(c http.Client, reqChan chan *ReqInfo, respChan chan Response) {
	for req := range reqChan {
		resp, err := c.PostForm(*req.srv, req.payload)
		r := Response{resp, err}
		respChan <- r
	}
}

// Consumer
func consumer(respChan chan Response, numCalls int) {
	var conns = 0
	var counter = 0
	for conns < numCalls {
		select {
		case r, ok := <-respChan:
			if ok {
				if r.err != nil {
					log.Println(r.err)
				} else {
					if r.StatusCode == 200 {
						counter++
					}
				}
				conns++
			}
		}
	}
	fmt.Printf("Total 200 OK responses: %d\n", counter)
}

func main() {
	srv := flag.String("server", "", "Server address to call")
	numWorkers := flag.Int("numWorkers", 10, "Number of worker processes")
	numCalls := flag.Int("numCalls", 1000, "Number of API calls to be made")
	testID := flag.String("testID", "test-api", "Test ID that will be the key on Redis")
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())
	reqChan := make(chan *ReqInfo)
	respChan := make(chan Response)

	var payload = url.Values{
		"testID": {*testID},
	}

	start := time.Now()
	go dispatcher(reqChan, numCalls, srv, payload)
	go workerPool(reqChan, respChan, *numWorkers)
	consumer(respChan, *numCalls)
	took := time.Since(start)
	ns := took.Nanoseconds()
	av := ns / int64(*numCalls)
	average, err := time.ParseDuration(fmt.Sprintf("%d", av) + "ns")
	if err != nil {
		log.Println(err)
	}
	fmt.Printf("Connections:\t%d\nConcurrent:\t%d\nTotal time:\t%s\nAverage time:\t%s\n", *numCalls, *numWorkers, took, average)
}
