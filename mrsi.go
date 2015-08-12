package main

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/nightmouse/mrsi/client"
	"io/ioutil"
	"os"
	"runtime"
)

var GlobalIntVars []*client.IntVar
var GlobalStringVars []*client.StringVar

func Init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	GlobalIntVars = make([]*client.IntVar, 0)
	GlobalStringVars = make([]*client.StringVar, 0)
}

func runJson(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println("error: expecting one argument")
		os.Exit(1)
	}

	fileName := c.Args()[0]
	fd, err := os.Open(fileName)
	defer fd.Close()
	if err != nil {
		fmt.Println("unable to open ", fileName, ": ", err)
		os.Exit(1)
	}

	runConf := &client.RunConf{}

	dec := json.NewDecoder(fd)
	err = dec.Decode(runConf)
	if err != nil {
		fmt.Println("error parsing json in ", fileName, ": ", err)
		os.Exit(1)
	}

	if err = runConf.Check(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	runConf.Exec()
}

func initJson(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println("error: expecting one argument")
		os.Exit(1)
	}

	fileName := c.Args()[0]

	tmpUrls := []string{"http://localhost:8080/{1}/{2}.html"}
	tmpVals := []string{"index", "about", "contact"}
	tmpIntVars := []*client.IntVar{&client.IntVar{"{1}", 0, 42}}
	tmpStrVals := []*client.StringVar{&client.StringVar{"{2}", tmpVals}}
	profile, err := client.NewRunConf(
		uint32(100),
		uint32(8),
		"GET",
		map[string]string{"": ""},
		&client.URLRandomizer{0, tmpUrls, tmpIntVars, tmpStrVals},
		nil)
	if err != nil {
		fmt.Println("this shouldn't happen")
		os.Exit(1)
	}

	bytes, err := json.MarshalIndent(profile, "", "   ")
	if err != nil {
		fmt.Println("error encoding json: ", err)
		os.Exit(1)
	}
	err = ioutil.WriteFile(fileName, bytes, 0644)
	if err != nil {
		fmt.Println("error saving json to ", fileName, ": ", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func parseIntVar(c *cli.Context) {
	key := c.String("key")
	if len(key) == 0 {
		fmt.Println("--key is required parameter")
		os.Exit(1)
	}
	min := c.Int("min")
	max := c.Int("max")
	iv, err := client.NewIntVar(key, int64(min), int64(max))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	GlobalIntVars = append(GlobalIntVars, iv)
}

func parseStrVal(c *cli.Context) {
	key := c.String("key")
	if len(key) == 0 {
		fmt.Println("--key is required parameter")
		os.Exit(1)
	}

	vals := c.StringSlice("values")
	if vals == nil || len(vals) == 0 {
		fmt.Println("no values specified")
		os.Exit(1)
	}

	sv, err := client.NewStringVar(key, vals)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	GlobalStringVars = append(GlobalStringVars, sv)
}

func runCli(c *cli.Context) {

	// seed
	seed := c.Int("seed")
	// requests
	requests := c.Int("requests")

	// workers
	workers := c.Int("workers")
	if workers < 1 {
		fmt.Println("workers flag must be a positive integer greater than zero")
		os.Exit(1)
	}

	// method
	method := c.String("method")

	// body
	body := c.String("body")
	bodyBytes := []byte(body)
	if bodyBytes == nil || len(bodyBytes) == 0 {
		fmt.Println("body flag has no value")
		os.Exit(1)
	}

	// urls
	urls := c.StringSlice("urls")
	if urls == nil || len(urls) == 0 {
		fmt.Println("no urls specified")
		os.Exit(1)
	}

	runConf, err := client.NewRunConf(
		uint32(workers),
		uint32(requests),
		method,
		map[string]string{"": ""},
		&client.URLRandomizer{
			IntVars:    GlobalIntVars,
			StringVars: GlobalStringVars,
			Seed:       int64(seed),
			Urls:       urls},
		bodyBytes,
	)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	runConf.Exec()
}

func main() {
	app := cli.NewApp()
	app.Name = "mrsi"
	app.Version = "0.1.0"
	app.Usage = "benchmarks http servers with configurable urls"
	app.Commands = []cli.Command{
		{
			Name:   "run",
			Usage:  "Run jobs defined in a .json file",
			Action: runJson,
		},
		{
			Name:   "init",
			Usage:  "Intialize a .json file with a test profile",
			Action: initJson,
		},
		{
			Name:   "test",
			Usage:  "test a given set of urls specified on the command line",
			Action: runCli,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "seed, s",
					Value: 0,
					Usage: "random number seed"},

				cli.IntFlag{
					Name:  "workers, w",
					Value: 8,
					Usage: "number of workers which may send parallel requests"},

				cli.IntFlag{
					Name:  "requests, r",
					Value: 100,
					Usage: "total number of requests to send"},

				cli.StringFlag{
					Name:  "method, m",
					Value: "GET",
					Usage: "HTTP method to use in the requests"},

				cli.StringFlag{
					Name:  "body, b",
					Value: "GET",
					Usage: "request body"},

				cli.StringSliceFlag{
					Name:  "urls, u",
					Usage: ""},
			},

			Subcommands: []cli.Command{
				cli.Command{
					Name:   "strval",
					Usage:  "defines a substitution randomly chosen from 'values' for key where 'key' may be in a url",
					Action: parseStrVal,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "key",
							Value: ""},

						cli.StringSliceFlag{
							Name: "values"},
					},
				},
				cli.Command{
					Name:   "intval",
					Usage:  "defines a substitution between min and max for key, where 'key' may be in a url",
					Action: parseIntVar,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "key",
							Value: ""},

						cli.IntFlag{
							Name:  "min",
							Value: 0},

						cli.IntFlag{
							Name:  "max",
							Value: 100},
					},
				},
			},
		},
	}

	app.Run(os.Args)

}
