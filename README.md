# mrsi


https://en.wikipedia.org/wiki/Artillery#MRSI

```
./mrsi --help
Usage of ./mrsi:
  -n=4: number of worker threads each requesting the url
  -r=1024: total number of requests
  intval
    -key="": a token to replace in the url
    -min=0: minimum number in a random range
    -max=0: maximum number in a random range
  strval
    -key="": a token to replace in the url
    -values="": a token to replace in the url
  url1, url2...

./mrsi -n 8 -r 100 intval -key="{1}" -min=0 -max=10  strval -key="{2}" -values="one,two,three" "http://localhost/{1}/{2}"  "http://localhost/items/id={1}"

```
