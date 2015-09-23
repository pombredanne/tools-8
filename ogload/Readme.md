## ogload

Originally based on the live-reload mechanism in https://github.com/spf13/hugo, ogload is for when you don't need features or sparkles.

Like decent security. Or templating. Orrr..

Look, all og does is load. He's not a smart program, but he is hard working and wants to help develop static html/css/js.



Ogload is heavily discouraged for anything that can even be mistaken for production use. Even if you've slammed expired 2006 4loko cans until you're 4 standard deviations past the Balmer peak, the author urges you to live a more responsible lifestyle by not using ogload in production.

```
ogload serves and hot-reloads static files in the current working directory

Usage:
  ogload [flags]
  ogload [command]

Available Commands:
  version     Print the current version
  help        Help about any command

Flags:
      --server_root="/":  Root webserver directory
      --static_files=".": Directory to serve static files from
  -a, --addr="127.0.0.1": Address to listen on
  -p, --port=8080:        Port to listen on

      --cert_file="":     /path/to/tls.cert
      --key_file="":      /path/to/tls.key

  -h, --help[=false]:     help for ogload
      


Use "ogload [command] --help" for more information about a command.
```
