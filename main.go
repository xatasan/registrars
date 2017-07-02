package main

import (
	"bufio"
	"fmt"
	"html/template"
	"math"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	uurl = "https://s.te.rs/" // upload url
	flen = 8                  // file name length
	maxf = 32e6               // max file size (byte)

	alph = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrtsuvw0123456789"

	FILE_TABLE = `<!DOCTYPE html>
<meta charset="utf-8" />
<table border="1"><tbody>
<tr><th>URL</th><th>Name</th><th>Hash</th><th>Size</th></tr>
{{range .Files}}
<tr>
<td><a href="{{.Url}}">{{.Url}}</a></td>
<td>{{ .Name }}</td><td><tt>{{ .Hash }}</tt></td>
<td><tt>{{ .Size }}</tt></td>
</tr>
{{end}}
</tbody></table>`
)

var (
	hdir, udir   string // hashsum directory, upload directory
	t, htmlop    *template.Template
	filec, storc  int64 // file count and total storage
	avgfs, meanfs int64 // average and mean file size
)

func init() {
	rand.Seed(time.Now().Unix())
	htmlop = template.Must(template.New("").Parse(FILE_TABLE))

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	hdir = wd + "/hdir/"
	udir = wd + "/udir/"

	go func() {
		for {
			files, err := ioutil.ReadDir(hdir)
			if err != nil {
				log.Fatalln(err)
			}
			filec = int64(len(files))

			storc = 0
			for _, f := range files {
				storc += int64(f.Size())
			}
			avgfs = int64(storc / int64(len(files)))
			
			if len(files) % 2 == 1 {
				midc := int(math.Ceil(float64(len(files)/2)))
				midf := int(math.Floor(float64(len(files)/2)))
				meanfs = (files[midc].Size() + files[midf].Size()) / 2
			} else {
				mid := int(len(files)/2)
				meanfs = files[mid].Size()
			}

			storc /= 1 << 20
			avgfs /= 1 << 20
			meanfs /= 1 << 20

			time.Sleep(time.Minute * 30)
		}
	}()

	go func() {
		input := bufio.NewScanner(os.Stdin)
		for input.Scan() {
			parts := strings.SplitN(input.Text(), "\t", 2)
			rto := parts[0]
			name := parts[1]

			to, err := time.Parse(time.UnixDate, rto)
			if err != nil {
				log.Println(err)
				continue
			}

			if time.Now().Before(to) {
				fmt.Printf("%s\t%s\n", to.Format(time.UnixDate), name)
			}

			time.AfterFunc(to.Sub(time.Now()), func() {
				if err := os.Remove(udir + name); err != nil {
					log.Println(err)
				}
			})

		}
	}()

	t = template.Must(template.ParseGlob("./static/*.gtml"))
}

func main() {
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/upload.php", upload)

	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		data := struct{ S, N, A, M int64 }{storc, filec, avgfs, meanfs}
		err := t.ExecuteTemplate(w, "about.gtml", data)
		if err != nil {
			log.Fatalln(err)
		}
	})

	fs := http.FileServer(http.Dir(udir))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			t.ExecuteTemplate(w, "index.gtml", nil)
		case "/style.css", "/img.png", "/favicon.ico":
			http.ServeFile(w, r, "./static"+r.URL.Path)
		default: // preferably to be handled by a web server 
			fs.ServeHTTP(w, r)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}
	log.Printf("Starting server on " + port)
	log.Fatal(http.ListenAndServe(port, nil))
}
