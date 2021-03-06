package main

//go:generate go-bindata -o assets.go assets/

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"os/user"
)

const (
	flen = 6       // file name length
	maxf = 1 << 25 // max file size (byte)
	alph = "uncopyrightable"
)

var (
	uurl       string // upload url
	hdir, udir string // hashsum directory, upload directory
	keeptf     bool
	t          *template.Template
	index      []byte
)

func main() {
	// setup file directories
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	hdir = wd + "/hdir/"
	udir = wd + "/udir/"
	os.Mkdir(hdir, 0755)
	os.Mkdir(udir, 0755)

	// load template
	t = template.New("reg").Funcs(template.FuncMap{"bytes": byteSize})
	for _, f := range []string{"index", "files"} {
		data, err := Asset("assets/" + f + ".gohtml")
		if err != nil {
			log.Fatalln(err)
		}
		t = template.Must(t.New(f).Parse(string(data)))
	}

	// check whether to keep hashfiles
	keeptf = os.Getenv("KEEPHF") != ""

	// find out what port to listen on
	host := os.Getenv("HOST")
	u, err := user.Current()
	if err == nil && u.Uid != "0" && host == "" {
		host = ":8080"
	}

	// set uurl, if possible
	if len(os.Args) > 1 {
		uurl = os.Args[1]
	}

	// start background processes
	go timeoutSetup()
	go statWorker()

	// start HTTP server
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/upload.php", upload)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if len(os.Args) <= 1 && r.URL.Path != "/" {
			http.FileServer(http.Dir(udir)).ServeHTTP(w, r)
			return
		}

		_, err := w.Write(index)
		if err != nil {
			log.Fatalln(err)
		}
	})

	log.Fatal(http.ListenAndServe(host, nil))
}
