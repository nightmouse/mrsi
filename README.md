# mrsi

mrsi sends parallel requests to a http server and measures the response times while the sever is under load.  

## Example Usage

Run a test on the command line using 8 workers to send 100 requests to "http://localhost/items/id={1}"
where _{1}_ is defined as a random number between 1 and 1024(inclusive).

```
./mrsi test -w 8 -r 100  -u "http://localhost/items/id={1}" intvar --key "{1}" --min 1 --max 1024
```
