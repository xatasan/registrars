package main

import (
	"crypto/sha1"
	"encoding/base32"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	Success     bool   `json:"success"`
	Errorcode   int    `json:"errorcode"`
	Description string `json:"description"`
	Files       []File `json:"files"`
}

type File struct {
	Name string `json:"name"`
	Url  string `json:"url"`
	Hash string `json:"hash"`
	Size int    `json:"size"`
}

func processFile(in io.Reader) (string, string, error) { // name, hash
	hash := sha1.New()
	_, err := io.Copy(hash, in)
	if err != nil {
		return "", "", err
	}
	hsum := strings.ToLower(base32.StdEncoding.EncodeToString(hash.Sum(nil)))

	for {
		name := ""
		for i := 0; i < flen; i++ {
			name += string(alph[rand.Intn(len(alph))])

		}
		if _, err := os.Stat(udir + name); os.IsNotExist(err) {
			return name, hsum, nil
		}
	}

}

func uploadFile(fh *multipart.FileHeader, to time.Duration) (file File, err error) {
	tmp, err := ioutil.TempFile("", fh.Filename)
	if err != nil {
		return
	}
	defer os.Remove(tmp.Name())

	filehandler, err := fh.Open()
	if err != nil {
		return
	}
	size, err := io.Copy(tmp, filehandler)
	if err != nil {
		return
	}

	tmp.Close()
	tmp, err = os.Open(tmp.Name())
	name, hash, err := processFile(tmp)
	if err != nil {
		return
	}
	name += path.Ext(fh.Filename)

	if _, err = os.Stat(hdir + hash); err == nil {
		if err = os.Symlink(hdir+hash, udir+name); err != nil {
			return
		}

		if to > 0 {
			fmt.Printf("%s\t%s\n", time.Now().Add(to).Format(time.UnixDate), name)
			time.AfterFunc(to, func() {
				if err := os.Remove(udir + name); err != nil {
					log.Println(err)
				}
			})
		}

		return File{
			Name: fh.Filename,
			Url:  uurl + name,
			Hash: hash,
			Size: int(size),
		}, nil
	}

	hfile, err := os.Create(hdir + hash)
	if err != nil {
		return
	}

	tmp.Close()
	tmp, err = os.Open(tmp.Name())
	if _, err = io.Copy(hfile, tmp); err != nil {
		return
	}

	if err = os.Symlink(hdir+hash, udir+name); err != nil {
		return
	}

	if to > 0 {
		fmt.Printf("%s\t%s\n", time.Now().Add(to).Format(time.UnixDate), name)
		time.AfterFunc(to, func() {
			if err := os.Remove(udir + name); err != nil {
				log.Println(err)
			}
		})
	}

	return File{
		Name: fh.Filename,
		Url:  uurl + name,
		Hash: hash,
		Size: int(size),
	}, nil
}

func upload(w http.ResponseWriter, req *http.Request) {
	mpr, err := req.MultipartReader()
	if err != nil {
		t.Lookup("error.gtml").Execute(w, err.Error())
		return
	}

	form, err := mpr.ReadForm(1e10)
	if err != nil {
		t.ExecuteTemplate(w, "error.gtml", err.Error())
		return
	}

	var res Response
	res.Success = true

	if len(form.File["files"]) == 0 {
		res.Success = false
		res.Errorcode = 400
		res.Description = "No input file(s)"
	} else {
		var to, unit time.Duration

		if _, ok := form.Value["tunit"]; ok {
			switch form.Value["tunit"][0] {
			case "w":
				unit = time.Hour * 24 * 7
			case "d":
				unit = time.Hour * 24
			case "h":
				unit = time.Hour
			case "m":
				fallthrough
			default:
				unit = time.Minute

			}
		}

		if raw_to, ok := form.Value["timeout"]; ok {
			if len(raw_to) == 1 {
				nto, err := strconv.Atoi(raw_to[0])
				if err == nil {
					to = time.Duration(nto) * unit
				}
			}
		}

		for _, fh := range form.File["files"] {
			switch path.Ext(fh.Filename) {
			case ".exe", ".bat", ".cmd", ".msi", ".vbs", ".scr", "":
				res.Success = false
				res.Errorcode = 403
				res.Description = fh.Filename + "not allowed to be uploaded"
				break
			}

			file, err := uploadFile(fh, to)
			if err != nil {
				res.Success = false
				res.Errorcode = 500
				res.Description = err.Error()
				break
			}
			res.Files = append(res.Files, file)
		}

		fsum := 0
		for _, f := range res.Files {
			fsum += f.Size
		}

		if fsum > maxf {
			for _, f := range res.Files {
				var name string
				fmt.Sscanf(f.Url, uurl+"%s", &name)
				go os.Remove(udir + name)
				go os.Remove(hdir + f.Hash)
			}

			res.Success = false
			res.Errorcode = 400
			res.Description = "File(s) above size limit"
			res.Files = nil
		}
	}

	if res.Success {
		switch req.URL.Query().Get("output") {
		case "gyazo":
			w.Header().Set("Content-Type", "text/plain")
			var urls []string
			for _, f := range res.Files {
				urls = append(urls, f.Url)
			}
			fmt.Fprintf(w, "%s", strings.Join(urls, "\n"))
		case "text":
			w.Header().Set("Content-Type", "text/plain")
			var urls []string
			for _, f := range res.Files {
				urls = append(urls, f.Url)
			}
			fmt.Fprintf(w, "%s\n", strings.Join(urls, "\n"))
		case "html":
			w.Header().Set("Content-Type", "text/html")
			htmlop.Execute(w, res)
		case "csv":
			w.Header().Set("Content-Type", "text/cvs")
			fmt.Fprint(w, "name,url,hash,size\n")
			csvw := csv.NewWriter(w)
			for _, f := range res.Files {
				csvw.Write([]string{f.Name, f.Url, f.Hash, fmt.Sprintf("%d", f.Size)})
			}
		default:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(res)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
