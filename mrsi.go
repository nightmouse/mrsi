package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/nightmouse/mrsi/urlrandomizer"
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

func main() {
	seed := flag.Int64("s", 0, "random seed number")
	workerCount := flag.Int("n", 4, "number of worker threads each requesting the url")
	requestCount := flag.Int("r", 1024, "total number of requests")

	intValFlag := flag.NewFlagSet("intval", flag.ExitOnError)
	intKey := intValFlag.String("key", "", "a token to replace in the url")
	intMin := intValFlag.Int64("min", 0, "minimum number in a random range")
	intMax := intValFlag.Int64("max", 0, "maximum number in a random range")

	strValFlag := flag.NewFlagSet("strchoice", flag.ExitOnError)
	strKey := strValFlag.String("key", "", "a token to replace in the url")
	strVal := strValFlag.String("values", "", "a comma delimited set of strings")

	intVals := make([]*urlrandomizer.IntVal, 0)
	strVals := make([]*urlrandomizer.StringVal, 0)

	flag.Parse()
	args := flag.Args()
	lastIndex := 0
	for i, v := range args {
		lastIndex = i
		switch v {
		case "intval":
			if err := intValFlag.Parse(args[i+1:]); err == nil {
				tmp, err := urlrandomizer.NewIntVal(*intKey, *intMin, *intMax)
				if err != nil {
					fmt.Println("unable to parse flags for intval: ", err)
					os.Exit(1)
				}
				intVals = append(intVals, tmp)

			} else {
				fmt.Println("unable to parse flags for intval: ", err)
				os.Exit(1)
			}
		case "strval":
			if err := strValFlag.Parse(args[i+1:]); err == nil {
				sv, err := urlrandomizer.NewStringVal(*strKey, *strVal)
				if err != nil {
					fmt.Println("unable to parse flags for intval: ", err)
					os.Exit(1)
				}
				strVals = append(strVals, sv)
			} else {
				fmt.Println("unable to parse flags for strval: ", err)
				os.Exit(1)
			}
		}
	}

	quitChan := make(chan bool)
	resultChan := make(chan result)

	TrapSigInt(quitChan)

	randomizer := urlrandomizer.NewURLRandomizer(*seed, args[lastIndex:], intVals, strVals)
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
	os.Exit(0)
}
