package client

import (
	"errors"
    "os"
	"os/signal"
	"io/ioutil"
	"time"
	"fmt"
    "net/url"
	"net/http"
)

// a struct to handle json martialing
// todo: add fields for results file, method
type RunConf struct {
    Requests uint32 `json: "requests"`
    Workers uint32 `json: "workers"`
    URLRandomizer
}

type Result struct { 
    Duration  time.Duration
    Bytes int64
    Err error
}

func (c *RunConf) worker(jobs chan *url.URL, results chan Result) {
	for {
		url, ok := <-jobs
        if !ok {
            return
        }

        start := time.Now()
        resp, err := http.Get(url.String())

        if err != nil {
            results <- Result{Err: err}
            continue
        }

        defer resp.Body.Close()
        if resp.StatusCode < 200 || resp.StatusCode >= 300 {
            err = errors.New("Server returned " + resp.Status)
            results <- Result{Err: err}
            continue
        }

        bytes, _ := ioutil.ReadAll(resp.Body)
        end := time.Now()
        results <- Result{Duration: end.Sub(start), Bytes: int64(len(bytes))}
        fmt.Println("finished: ", url, " ", len(bytes))
    }
}

func trapSigInt(quitChan chan bool) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		quitChan <- true
		return
	}()
}

func (c *RunConf) Exec() {
	quitChan := make(chan bool)
	resultChan := make(chan Result)

	trapSigInt(quitChan)

	jobChan := c.GetChannel(c.Requests, quitChan)

	// start workers
	for i := uint32(0); i != c.Workers; i++ {
		go c.worker(jobChan, resultChan)
	}

	// wait for results
	totalTime := 0.0
	totalBytes := int64(0)
	totalErrors := 0
	var totalResults uint32
	for totalResults = uint32(0); totalResults < c.Requests; totalResults++ {
		r := <-resultChan
		totalResults++
		if r.Err != nil {
			totalErrors++
			fmt.Println(r.Err)
		} else {
			totalTime += r.Duration.Seconds()
			totalBytes += r.Bytes
		}
	}
	fmt.Printf("%d requests %d errors %d byes %f seconds\n", totalResults, totalErrors, totalBytes, totalTime)
	os.Exit(0)
}
