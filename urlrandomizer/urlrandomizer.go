package urlrandomizer

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
)

type IntVal struct {
	Key string
	Min int64
	Max int64
}

func NewIntVal(key string, min, max int64) (*IntVal, error) {
	if min >= max {
		return nil, errors.New("intval must have a min value less than max")
	}
	return &IntVal{key, min, max}, nil
}

type StringVal struct {
	Key  string
	Vals []string
}

func NewStringVal(key, vals string) (*StringVal, error) {
	v := strings.Split(vals, ",")
	if len(v) == 0 {
		return nil, errors.New("no string values specified")
	}
	return &StringVal{key, v}, nil
}

type URLRandomizer struct {
	Seed       int64
	Urls       []string
	IntVals    []*IntVal
	StringVals []*StringVal
}

func NewURLRandomizer(seed int64, urls []string, intRanges []*IntVal, stringVals []*StringVal) *URLRandomizer {
	rand.Seed(seed)
	u := &URLRandomizer{
		Seed:       seed,
		Urls:       urls,
		IntVals:    intRanges,
		StringVals: stringVals}
	return u
}

func (u *URLRandomizer) subInts(url string) string {
	for _, r := range u.IntVals {
		delta := int(math.Abs(float64(r.Max-r.Min))) + 1
		random := rand.Int()%delta - int(math.Abs(float64(r.Min)))
		url = strings.Replace(url, r.Key, strconv.Itoa(random), -1)
	}
	return url
}

func (u *URLRandomizer) subStrs(url string) string {
	for _, s := range u.StringVals {
		url = strings.Replace(url, s.Key, s.Vals[rand.Intn(len(s.Vals))], -1)
	}
	return url
}

func (u *URLRandomizer) GetChannel(numRequests int, quit chan bool) chan *url.URL {
	ch := make(chan *url.URL)
	go func() {
		defer close(ch)
		count := 0
		size := len(u.Urls)
		for i := 0; i != numRequests; i++ {
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
