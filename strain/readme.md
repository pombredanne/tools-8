# Strain

## (._+ )☆＼(-.-メ)

Dead simple tool to whack an HTTP server with requests in order help test load balancing/capacity/etc.

## Usage
`go get github.com/Kavec/strain`

```
USAGE:
   strain [global options] http://test-server-url.com:8080/path/to/GET

VERSION:
   0.1.0

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --workers,    -w      "10"  Number of workers
   --repeat,     -n      "-1"  How often each worker repeats before terminating
   --rate,       -r   "200ms"  How often each worker pounds the server
   --quiet,      -q            Don't give periodic reports
   
   --report_rate         "1s"  How often to produce periodic reports
   
   --help,       -h            show help
   --version,    -v            print the version
```
