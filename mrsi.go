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
	"github.com/nightmouse/mrsi/urlrandomizer"
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


var seed = flag.Int("s", 0, "random seed number")
var workerCount = flag.Int("n", 4, "number of worker threads each requesting the url")
var requestCount = flag.Int("r", 1024, "total number of requests")

// mrsi -s 348547 -n 8 -r 100000000 --intchoice --key "{{i1}}" --min 0 --max 1000  --stringchoice --key "{{s1}}" --choices "items,users" "http://localhost:8080/{{s1}}/{{i1}}/"

func main() {
	flag.Parse()

	quitChan := make(chan bool)
	resultChan := make(chan result)

	TrapSigInt(quitChan)

    randomizer := urlrandomizer.NewUrlRandomizer(seed, flag.Args(), intChoices, stringChoices)
	jobChan := randomizer.GetChannel(*requestCount, quitChan)

	// start workers
	for i := 0; i != (*workerCount); i++ {
		go worker(jobChan, resultChan)
	}

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
