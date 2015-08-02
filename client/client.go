package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"
)

type RunConf struct {
	Requests uint32            `json:"requests"`
	Workers  uint32            `json:"workers"`
	Method   string            `json:"method,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
	*URLRandomizer
	Body json.RawMessage `json:"body,omitempty,string"`
	file string          `json:"file,omitempty"` // this is used for json only
}

type Result struct {
	Duration time.Duration
	Bytes    int64
	Err      error
}

func NewRunConf(requests, workers uint32, method string, headers map[string]string, ur *URLRandomizer, requestBody []byte) (*RunConf, error) {
	r := &RunConf{
		requests,
		workers,
		method,
		headers,
		ur,
		requestBody,
		""}

	if err := r.Check(); err != nil {
		return nil, err
	}
	return r, nil
}

func (c *RunConf) Check() error {
	var err error
    
	// make sure the method, the request body and file are set appropriately
	c.Method = strings.ToUpper(c.Method)
	switch c.Method {
	case "GET", "DELETE", "HEAD":
		if c.Body != nil {
			err = errors.New("Can't specify a request body for " + c.Method)
		}
	case "PUT", "POST", "PATCH":
		if len(c.file) == 0 && (c.Body == nil || len(c.Body) == 0) {
			err = errors.New(c.Method + " requres a request body")
		}

		if len(c.file) != 0 {
			c.Body, err = ioutil.ReadFile(c.file)
		}

	case "TRACE", "OPTIONS", "CONNECT":
		err = errors.New(c.Method + " is not supported")
	default:
		err = errors.New("invalid method: " + c.Method)
	}
	return err
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

	urlChan := c.GetChannel(c.Requests, quitChan)

	// start workers
	for i := uint32(0); i != c.Workers; i++ {
		go c.worker(urlChan, resultChan)
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
