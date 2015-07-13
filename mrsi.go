// https://en.wikipedia.org/wiki/Artillery#MRSI

package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"time"
)

type result struct {
	Duration time.Duration
	Bytes    int64
	Err      error
}

func worker(jobs chan *url.URL, results chan result) {
	for {
		select {
		case url, ok := <-jobs:
			if !ok {
				return
			}

			start := time.Now()
			resp, err := http.Get(url.String())

			if err != nil {
				results <- result{Err: err}
				continue
			}

			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				err = errors.New("Server returned " + resp.Status)
				results <- result{Err: err}
				continue
			}

			bytes, _ := ioutil.ReadAll(resp.Body)
			end := time.Now()
			results <- result{Duration: end.Sub(start), Bytes: int64(len(bytes))}
			fmt.Println("finished: ", url, " ", len(bytes))
		}
	}
}

func Init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func TrapSigInt(quitChan chan bool) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		quitChan <- true
		return
	}()
}

var workerCount = flag.Int("n", 4, "number of worker threads each requesting the url")
var requestCount = flag.Int("r", 1024, "total number of requests")

func main() {
	flag.Parse()

	quitChan := make(chan bool)
	jobChan := make(chan *url.URL)
	resultChan := make(chan result)

	TrapSigInt(quitChan)

	// parse the urls
	urls := make([]*url.URL, flag.NArg())
	for i := 0; i != flag.NArg(); i++ {
		u, err := url.Parse(flag.Arg(i))
		if err != nil {
			fmt.Println(err)
			return
		}
		urls[i] = u
	}

	// start workers
	for i := 0; i != (*workerCount); i++ {
		go worker(jobChan, resultChan)
	}

	// distribute the jobs
	go func() {
		for i := 0; i != (*requestCount); i++ {
			select {
			case <-quitChan:
				break
			default:
				jobChan <- urls[i%len(urls)]
			}
		}
		close(jobChan)
	}()

	// wait for results
	totalTime := 0.0
	totalBytes := int64(0)
	totalErrors := 0
	var totalResults int
	for totalResults = 0; totalResults < (*requestCount); totalResults++ {
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
}