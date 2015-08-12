package client

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
)

type IntVar struct {
	Key string `json:"key"`
	Min int64  `json:"min"`
	Max int64  `json:"max"`
}

func NewIntVar(key string, min, max int64) (*IntVar, error) {
	if len(key) == 0 {
		return nil, errors.New("key must have a length")
	} else if min >= max {
		return nil, errors.New("intval must have a min value less than max")
	}
	return &IntVar{key, min, max}, nil
}

type StringVar struct {
	Key  string   `json:"key"`
	Vals []string `json:"values"`
}

func NewStringVar(key string, vals []string) (*StringVar, error) {
	if len(key) == 0 {
		return nil, errors.New("key must have a length")
	} else if len(vals) == 0 {
		return nil, errors.New("values must have a length")
	}

	return &StringVar{key, vals}, nil
}

type URLRandomizer struct {
	Seed       int64        `json:"seed,omitempty"`
	Urls       []string     `json:"urls"`
	IntVars    []*IntVar    `json:"intvars"`
	StringVars []*StringVar `json:"stringvars"`
}

func NewURLRandomizer(seed int64, urls []string, intRanges []*IntVar, stringVals []*StringVar) *URLRandomizer {
	rand.Seed(seed)
	u := &URLRandomizer{
		Seed:       seed,
		Urls:       urls,
		IntVars:    intRanges,
		StringVars: stringVals}
	return u
}

func (u *URLRandomizer) subInts(url string) string {
	for _, r := range u.IntVars {
		delta := int(math.Abs(float64(r.Max-r.Min))) + 1
		random := rand.Int()%delta - int(math.Abs(float64(r.Min)))
		url = strings.Replace(url, r.Key, strconv.Itoa(random), -1)
	}
	return url
}

func (u *URLRandomizer) subStrs(url string) string {
	for _, s := range u.StringVars {
		url = strings.Replace(url, s.Key, s.Vals[rand.Intn(len(s.Vals))], -1)
	}
	return url
}

func (u *URLRandomizer) GetChannel(numRequests uint32, quit chan bool) chan *url.URL {
	ch := make(chan *url.URL)
	go func() {
		defer close(ch)
		count := 0
		size := len(u.Urls)
		for i := uint32(0); i != numRequests; i++ {
			select {
			case <-quit:
				break
			default:
				rawUrl := u.subStrs(u.subInts(u.Urls[count%size]))
				cookedUrl, err := url.Parse(rawUrl)
				if err != nil {
					fmt.Println("Aborting program for bad url: ", err)
					break
				}
				ch <- cookedUrl
				count++
			}
		}
	}()
	return ch
}
