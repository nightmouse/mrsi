# mrsi


https://en.wikipedia.org/wiki/Artillery#MRSI

```
./mrsi --help
NAME:
   mrsi - benchmarks http servers with configurable urls

USAGE:
   mrsi [global options] command [command options] [arguments...]
   
VERSION:
   0.1.0
   
COMMANDS:
   run		Run jobs defined in a .json file
   init		Intialize a .json file with a test profile
   test		test a given set of urls specified on the command line
   help, h	Shows a list of commands or help for one command
   
GLOBAL OPTIONS:
   --help, -h		show help
   --version, -v	print the version


```

## Example Usage

Run a test on the command line using 8 workers to send 100 requests to "http://localhost/items/id={1}"
where _{1}_ is defined as a random number between 1 and 1024(includsive).

```
./mrsi test -w 8 -r 100  -u "http://localhost/items/id={1}" intval --key "{1}" --min 1 --max 1024
```
