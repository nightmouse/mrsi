package client

import (
	"net"
	//"crypto/tls"
	"io/ioutil"
)

var counter chan uint64

func init() {
    counter = make(chan uint64)
    count := uint64(0)
    go func() {
        for {
            count += 1
            counter <- count
        }
    }()
}

//func DialTLS(host string, timing *Timing) func(string, string)(*net.Conn, Timing, error) {
//}

func DialHttp(host string, timing *Timing) func(*net.Conn, error) {
	return func(protcol, host string) (*net.Conn, error) {
		start := time.Now()
		conn, err := net.Dial(protocol, host)
		end := time.Now()
		timing.Category = "connect"
		timing.Duration = end.Sub(start)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}

func GETRequest(rc *RunConf, url *url.URL) (*Result) {

    result := &Result {
        RequestNo: <- counter,
        Timings: make([]Timing, 1), // TODO: increase to the number of timings
        Bytes: uint64(0),
        Err:    nil, }

	// create a transport
	transport := http.Transport{}

	if url.Proto == "http" {
	    transport.Dial = DialHttp(url.Host, &result.Timings[0])
	} else if url.Proto == "https" {
	    transport.DialTLS = DialTLS(url.Host, &result.Timings[0])
	}

	// create a request and add the headers
	req, err := http.NewRequest("GET", *url, nil)
	for k,v := range rc.Headers {
		req.Header.Add(k, v)
	}

	// create a client with the custom client with the transport above
	client = http.Client{Transport: transport}

	// write the request method
	// TODO - subtract the connect timing from the request timing
	start := time.Now()
	resp, err := client.Do(req)
	end := time.Now()

	timing := Timing{Category: "request"}
	timing.Duration = end.Sub(start)
    result.Timings = append(result.Timings, timing)
    result.Err = err

    if err != nil {
        timing = Timing{Category: "processing"}
        start = time.Now();
        _, err := ioutil.ReadAll(resp.Body)
        resp.Body.Close();
        end = time.Now();
        result.Err = err
	    timing.Duration = end.Sub(start)
        result.Bytes = resp.ContentLength;
        result.Timings = append(result.Timings, timing)
    }

    return result
}

//func POSTRequest(*net.Conn, url *url.URL, headers map[string]string) (*Result, error) {
//}

//func PUTRequest(*net.Conn, url *url.URL, headers map[string]string) (*Result, error) {
//}

//func HEADRequest(*net.Conn, url *url.URL , headers map[string]string) (*Result, error) {
//}

//func PATCHRequest(*net.conn, url *url.URL , headers map[string]string) (*Result, error) {
//}

//func DELETERequest(*net.conn, url *url.URL , headers map[string]string) (*Result, error) {
//}
