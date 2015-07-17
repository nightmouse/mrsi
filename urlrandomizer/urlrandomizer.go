package urlrandomizer

import (
    "strings"
    "strconv"
    "math/rand"
    "math"
    "net/url"
    "fmt"
)

type IntChoice struct {
    Key string
    Min int64
    Max int64
}

type StringChoice struct { 
    Key string
    Choices []string
}

type URLRandomizer struct {
    Seed int64
    Urls []string
    IntChoices []IntChoice
    StringChoices []StringChoice
}

func NewURLRandomizer(seed int64, urls []string, intRanges []IntChoice, stringChoices []StringChoice) *URLRandomizer {
    rand.Seed(seed);
    u := &URLRandomizer {
        Seed: seed,
        Urls: urls,
        IntChoices: intRanges,
        StringChoices: stringChoices,}
    return u
}

func (u *URLRandomizer) subInts(url string) string { 
    for _,r := range u.IntChoices {
        delta := int(math.Abs(float64(r.Max-r.Min))) +1
        random := rand.Int() % delta - int(math.Abs(float64(r.Min)))
        url = strings.Replace(r.Key, url, strconv.Itoa(random), 0)
    }
    return url
}

func (u *URLRandomizer) subStrs(url string) string { 
    for _,s := range u.StringChoices {
        url = strings.Replace(s.Key, url, s.Choices[rand.Intn(len(s.Choices))], 0)
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
                    fmt.Println("aborting program for bad url: ", err)
                    break
                }
                ch <- cookedUrl
                count++
            }
        }
    }()
    return ch
}
