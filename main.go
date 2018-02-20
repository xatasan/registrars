package main

//go:generate go-bindata -o assets.go assets/

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

const (
	flen = 6       // file name length
	maxf = 1 << 32 // max file size (byte)
	alph = "uncopyrightable"
)

var (
	uurl       string // upload url
	hdir, udir string // hashsum directory, upload directory
	keeptf     bool
	t          *template.Template
)

func init() {
	// setup file directories
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	hdir = wd + "/hdir/"
	udir = wd + "/udir/"

	// load template
	t = template.New("reg")
	for _, f := range []string{"index", "files"} {
		data, err := Asset("assets/" + f + ".gohtml")
		if err != nil {
			log.Fatalln(err)
		}
		t = template.Must(t.New(f).Parse(string(data)))
	}

	// check whether to keep hashfiles
	keeptf = os.Getenv("KEEPHF") != ""
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "usage: %s [uurl]", os.Args[0])
		os.Exit(1)
	}
	uurl = os.Args[1]

	// start HTTP server
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/upload.php", upload)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := t.ExecuteTemplate(w, "index", data)
		if err != nil {
			log.Fatalln(err)
		}
	})
	log.Fatal(http.ListenAndServe(os.Getenv("HOST"), nil))
}
