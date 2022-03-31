package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Payload struct {
	testID string
}

func main() {
	srv := flag.String("server", "", "Server address to call")
	// numReq := flag.Int("numReq", 100, "Number of API calls to make per second")
	testID := flag.String("testID", "test-api", "Test ID that will be the key on Redis")
	flag.Parse()

	c := http.Client{Timeout: time.Duration(1) * time.Second}
	// payload, _ := json.Marshal(Payload{testID: *testID})
	// var payload = fmt.Sprintf("testID=%s", testID)
	var payload = url.Values{
		"testID": {*testID},
	}

	var counter = 0

	// var timeInterval = 1.0 / float64(*numReq)
	for i := 0; i < 100; i++ {
		resp, err := c.PostForm(*srv, payload)
		if err != nil {
			fmt.Printf("Error %s", err)
			return
		}
		if resp.StatusCode == 200 {
			counter++
		}
		defer resp.Body.Close()
		time.Sleep(time.Duration(1) * time.Millisecond)
	}
	fmt.Printf("Total 200 OK responses: %d\n", counter)

	resp, err := c.Get(*srv + "?" + payload.Encode())
	if err != nil {
		fmt.Printf("Error %s", err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	fmt.Printf("Redis API count: %s\n", body)
}
