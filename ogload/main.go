package main

import (
	"os"
	"fmt"
	"time"
	"bytes"
	"strconv"
	"strings"
	"net/http"
	"path/filepath"

	"gopkg.in/fsnotify.v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	lr "github.com/jaschaephraim/lrserver"

	"github.com/Kavec/ogload/watcher"

	. "github.com/Kavec/ogload/config" // pull config structures into main.go
)

type signal struct{}

func main() {
	Ogload.Run = func(cmd *cobra.Command, args[] string) {

		//livereload.Initialize()

		fmt.Println("Running ", Version())

		launchServer()
	}

	
	Ogload.Execute()
}

// launchServer sets up the http fileserver. Exciting!
func launchServer() {
	r := mux.NewRouter().
		StrictSlash(true)

	// Grab all our config junk and prepare to launch
	ip   := viper.GetString("ListenAddr")
	port := viper.GetInt("ListenPort")
	fmt.Println(viper.GetInt("ListenPort"))

	

	// I... I guess, if you want TLS, you can totally have it
	cert, key := viper.GetString("CertFile"), viper.GetString("KeyFile")
	useTLS := len(cert) > 0 && len(key) > 0
	scheme := "https"
	if !useTLS {
		if port == 0      { port =  80 }
		scheme = "http"
	} else if port == 0 { port = 443 }

	p := fmt.Sprintf("%s:%d", ip, port)

	root, files := viper.GetString("ServerRoot"), viper.GetString("StaticFiles")

	reload, err := lr.New(lr.DefaultName, lr.DefaultPort)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	go reload.ListenAndServe()

	r.PathPrefix(root).
	Handler( 
		handlers.CombinedLoggingHandler(
			os.Stdout, injectReload(root, files, scheme),
		),
	)

	go func() {
		fmt.Printf("Launching ogload on %s\n", p)
		if !useTLS {
			if err := http.ListenAndServe(p, r); err != nil {
				fmt.Printf("Server failed! scheme=%s, addr=%s, err=%v\n", scheme, p, err)
				os.Exit(1)
			}
			return
		}

		if err := http.ListenAndServeTLS(p, cert, key, r); err != nil {
			fmt.Printf("TLS Server failed! scheme=%s, addr=%s, err=%v\n", scheme, p, err)
			os.Exit(1)
		}
	}()

	wait := watchDir(files, reload)
	<- wait
}

type injectWriter struct {
	w http.ResponseWriter
	h http.Header
	scheme string
}

func newInjectWriter(w http.ResponseWriter, scheme string) *injectWriter {
	return &injectWriter{
		w: w,
		h: make(http.Header),
		scheme: scheme,
	}
}

func (i *injectWriter) Header() http.Header {
	return i.h
}

func (i *injectWriter) Write(b []byte) (int, error) {
	//port    := viper.GetInt("ListenPort")
	replace := []byte(
	`	<script type="text/javascript">
			host = (location.hostname || 'localhost');
			document.write('<script type="text/javascript" '
				+ 'src="//' + host + ':35729/livereload.js'
				+ '?mindelay=10&host=' + host + '">'
				+ '</' + 'script>');
		</script>
	</body>`,
	)
	tag := []byte("</body>")
	
	out := bytes.Replace(b, tag, replace, -1)
	if len(out) == len(b) {
		tag := []byte("</BODY>")
		out  = bytes.Replace(b, tag, replace, -1)
	}

	i.h["Content-Length"] = []string{strconv.Itoa(len(out))}
	for k, v := range i.h {
		i.w.Header()[k] = v
	}

	return i.w.Write(out)
}

func (i *injectWriter) WriteHeader(status int) {
	i.w.WriteHeader(status)
}

func injectReload(root, files, scheme string) http.Handler {
	fileHnd := http.StripPrefix(root, http.FileServer(http.Dir(files)))

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Sneakily drop code into any html files; but if they are something
			// else, just let them pass through
			if len(r.URL.Path) == 1 || strings.HasSuffix(r.URL.Path, ".html") {
				inject := newInjectWriter(w, scheme)
				fileHnd.ServeHTTP(inject, r)
				return
			}
			fileHnd.ServeHTTP(w, r)
		},
	)
}

// watchDir fires off a goroutine and hands back a signalling channel that
// can be used to tell it to gracefully die or be used to block a function
// until it dies. The channel is closed automatically via defer when the 
// goroutine is terminated or encounters an unrecoverable error
func watchDir(dir string, lrserv *lr.Server) chan signal {
	die := make(chan signal)
	/*
		The code below has been copied with modifications from github.com/spf13/hugo
		// Copyright © 2013-2015 Steve Francia <spf@spf13.com>.
		//
		// Licensed under the Simple Public License, Version 2.0 (the "License");
		// you may not use this file except in compliance with the License.
		// You may obtain a copy of the License at
		// http://opensource.org/licenses/Simple-2.0
		//
		// Unless required by applicable law or agreed to in writing, software
		// distributed under the License is distributed on an "AS IS" BASIS,
		// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
		// See the License for the specific language governing permissions and
		// limitations under the License.
	*/
	/* 
		TODO: See if we actually need to smash and grab tweakLimit() from hugo.go
		to bump up OSX file descriptors
	*/
	watcher, err := watcher.New(250 * time.Millisecond)
	if err != nil {
		fmt.Printf("Couldn't launch filesystem watcher: %v\n", err)
		os.Exit(1)
	}

	addPaths(watcher)

	lrserv.SetLiveCSS(true)
	go func() {
		reload := make(chan string)
		defer func() {
			close(reload)
			close(die)
			watcher.Close()
		}()

		for {
			select {
			case events := <- watcher.Events:
				for _, event := range events {
					ext := filepath.Ext(event.Name)
					temp := strings.HasSuffix(ext, "~") || (ext == ".swp") || (ext == ".swx") || (ext == ".tmp") || strings.HasPrefix(ext, ".goutputstream")
					if temp {
						continue
					}
					// renames are always followed with Create/Modify
					if event.Op&fsnotify.Rename == fsnotify.Rename {
						continue
					}

					// add new directory to watch list
					if s, err := os.Stat(event.Name); err == nil && s.Mode().IsDir() {
						if event.Op&fsnotify.Create == fsnotify.Create {
							watcher.Add(event.Name)
						}
					}
					// Drop the changed name into the queue to be reloaded
					go func() {
						reload <- event.Name
					}()
				}

			case file := <- reload:
				absFile, err := filepath.Abs(file)
				if err != nil {
					fmt.Printf("Unable to convert %s to absolute path %#v\n", file, err)
					continue
				}

				lrserv.Reload(absFile)

			case err := <- watcher.Errors:
				if err != nil { fmt.Println("Received filesystem error: ", err) }
			}
		}
	}()

	return die
}

func addPaths(watch *watcher.Batcher) {
	/*
		As above, copied with minor modifications:
		// Copyright © 2013-2015 Steve Francia <spf@spf13.com>.
		//
		// Licensed under the Simple Public License, Version 2.0 (the "License");
		// you may not use this file except in compliance with the License.
		// You may obtain a copy of the License at
		// http://opensource.org/licenses/Simple-2.0
		//
		// Unless required by applicable law or agreed to in writing, software
		// distributed under the License is distributed on an "AS IS" BASIS,
		// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
		// See the License for the specific language governing permissions and
		// limitations under the License.
	*/
	var paths []string
	dir := viper.GetString("StaticFiles")
	if filepath.IsAbs(dir) {
		dir = filepath.Clean(dir)
	} else {
		dir = filepath.Join(viper.GetString("WorkingDir"), dir)
	}

	walker := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Error walking directory: ", dir, err)
			return nil
		}

		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			link, err := filepath.EvalSymlinks(path)
			if err != nil {
				fmt.Printf("Cannot read symbolic link '%s': %v\n", path, err)
				return nil
			}

			linkfi, err := os.Stat(link)
			if err != nil {
				fmt.Printf("Cannot stat '%s', error was: %s", link, err)
				return nil
			}
			if !linkfi.Mode().IsRegular() {
				fmt.Printf("Symbolic links for directories not supported, skipping '%s'", path)
				return nil
			}
		}

		if fi.IsDir() {
			if fi.Name() == ".git"         || 
			   fi.Name() == ".gitignore"   || 
			   fi.Name() == "node_modules" || 
			   fi.Name() == "bower_components" {

				return filepath.SkipDir
			}
			paths = append(paths, path)
		}

		return nil
	}

	filepath.Walk(dir, walker)

	for _, path := range paths {
		watch.Add(path)
	}
}