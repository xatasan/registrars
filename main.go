package main

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
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
	hdir, udir string // hashsum directory, upload directory
	t, htmlop  *template.Template
)

func init() {
	rand.Seed(time.Now().Unix())
	htmlop = template.Must(template.New("").Parse(FILE_TABLE))
	t = template.Must(template.ParseGlob("./static/*.gtml"))

	// setup file directories
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	hdir = wd + "/hdir/"
	udir = wd + "/udir/"

	// regenerate old temporary files
	go regenFrom(os.Stdin)

}

func main() {
	fs := http.FileServer(http.Dir(udir))
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/upload.php", upload)
	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		err := t.ExecuteTemplate(w, "about.gtml", data)
		if err != nil {
			log.Fatalln(err)
		}
	})
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
	log.Printf("Starting server on %q\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
