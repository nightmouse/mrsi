package main

import (
	"fmt"
	"github.com/nightmouse/mrsi/client"
	"os"
	"runtime"
    "github.com/codegangsta/cli"
    "encoding/json"
    "io/ioutil"
)


func Init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func runJson(c *cli.Context) {
    fmt.Println("runJson: ", c.Args())
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
    runConf.Exec()
}

func initJson(c *cli.Context) {
    if len(c.Args()) != 1 { 
        fmt.Println("error: expecting one argument")
        os.Exit(1)
    }

	fileName := c.Args()[0]

	tmpUrls:= []string{"http://localhost:8080/{1}/{2}.html"}
	tmpVals := []string{"index", "about", "contact"}
	tmpIntVals := []*client.IntVal{ &client.IntVal{"{1}",0,42} }
    tmpStrVals := []*client.StringVal{ &client.StringVal{"{2}", tmpVals}}
    profile := client.RunConf{
			uint32(100),
			uint32(8),
			client.URLRandomizer{ 0, tmpUrls, tmpIntVals, tmpStrVals }} 

	bytes, err := json.MarshalIndent(profile, "", "   ")
	if err != nil { 
		fmt.Println("error encoding json: ", err)
		os.Exit(1)
	}

	ioutil.WriteFile(fileName, bytes, 0644)

	os.Exit(0)
}

func main() {
    app := cli.NewApp()
    app.Name = "mrsi"
    app.Usage = "benchmarks http servers with configurable urls"
    app.Commands = []cli.Command{ 
        {
            Name: "run",
            Usage: "Run jobs defined in a .json file",
            Action: runJson,
        },
        {
            Name: "init",
            Usage: "Intialize a .json file with a test profile",
            Action: initJson,
        },
    }

    
    app.Run(os.Args)

/*
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

	intVals := make([]*client.IntVal, 0)
	strVals := make([]*client.StringVal, 0)

	flag.Parse()
	args := flag.Args()
	lastIndex := 0
	for i, v := range args {
		lastIndex = i
		switch v {
		case "intval":
			if err := intValFlag.Parse(args[i+1:]); err == nil {
				tmp, err := client.NewIntVal(*intKey, *intMin, *intMax)
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
				sv, err := client.NewStringVal(*strKey, *strVal)
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
*/
}
