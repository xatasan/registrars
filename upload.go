package main

import (
	"crypto/sha1"
	"encoding/base32"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	Success     bool    `json:"success"`
	Errorcode   int     `json:"errorcode"`
	Description string  `json:"description"`
	Files       []*File `json:"files"`
}

type File struct {
	Uname   string
	Name    string   `json:"name"`
	Url     *url.URL `json:"url"`
	Hash    string   `json:"hash"`
	Size    int64    `json:"size"`
	timeout time.Time
}

func init() {
	rand.Seed(time.Now().Unix())
}

func (f *File) String() string {
	return fmt.Sprintf("%d\t%s\t%s\n",
		f.timeout.Unix(), f.Uname, f.Hash)
}

func toFile(str string) *File {
	var timeout int64
	var name, hash string
	fmt.Sscanf(str, "%d\t%s\t%s", timeout, name, hash)
	return &File{
		Uname:   name,
		Hash:    hash,
		timeout: time.Unix(timeout, 0),
	}
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

func uploadData(file *os.File, orig, name, hash string, size int64) (*File, error) {
	var err error
	if _, err = os.Stat(hdir + hash); err == nil {
		if err = os.Symlink(hdir+hash, udir+name); err != nil {
			return nil, err
		}

		u, err := url.Parse(uurl + name)
		if err != nil {
			return nil, err
		}
		return &File{
			Uname: name,
			Name:  orig,
			Url:   u,
			Hash:  hash,
			Size:  int64(size),
		}, nil
	}

	hfile, err := os.Create(hdir + hash)
	if err != nil {
		return nil, err
	}

	file.Close()
	file, err = os.Open(file.Name())
	if _, err = io.Copy(hfile, file); err != nil {
		return nil, err
	}

	if err = os.Symlink(hdir+hash, udir+name); err != nil {
		return nil, err
	}

	u, err := url.Parse(uurl + name)
	if err != nil {
		return nil, err
	}
	return &File{
		Name:  orig,
		Uname: name,
		Url:   u,
		Hash:  hash,
		Size:  int64(size),
	}, nil
}

func uploadText(inp io.Reader) (file *File, err error) {
	tmp, err := ioutil.TempFile("", "")
	if err != nil {
		return
	}

	defer os.Remove(tmp.Name())

	size, err := io.Copy(tmp, inp)
	if err != nil {
		return
	}

	tmp.Close()
	tmp, err = os.Open(tmp.Name())
	name, hash, err := processFile(tmp)
	if err != nil {
		return
	}
	name += ".txt"

	return uploadData(tmp, "paste.txt", name, hash, size)
}

func uploadFile(fh *multipart.FileHeader) (file *File, err error) {
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

	return uploadData(tmp, fh.Filename, name, hash, size)
}

func upload(w http.ResponseWriter, req *http.Request) {
	var (
		res     Response
		unit    time.Duration
		timeout time.Time
	)
	res.Success = true

	mpr, err := req.MultipartReader()
	if err != nil {
		fmt.Fprintln(w, err.Error())
		return
	}

	form, err := mpr.ReadForm(1 << 32)
	if err != nil {
		fmt.Fprintln(w, err.Error())
		return
	}

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
		nto, err := strconv.Atoi(raw_to[0])
		if err == nil && nto > 0 {
			timeout = time.Now().Add(time.Duration(nto) * unit)
		}
	}

	if _, ok := form.Value["text"]; ok {
		file, err := uploadText(strings.NewReader(form.Value["text"][0]))
		if err != nil {
			res.Success = false
			res.Errorcode = 500
			res.Description = err.Error()
		}

		if file.Size > maxf {
			os.Remove(udir + file.Uname)
			os.Remove(hdir + file.Hash)

			res.Success = false
			res.Errorcode = 400
			res.Description = "File above size limit"
			res.Files = nil
		}

		res.Files = append(res.Files, file)
	} else if len(form.File["files"]) == 0 {
		res.Success = false
		res.Errorcode = 400
		res.Description = "No input file(s)"
	} else {
		for _, fh := range form.File["files"] {
			switch path.Ext(fh.Filename) {
			case ".exe", ".bat", ".cmd", ".msi", ".vbs", ".scr", "":
				res.Success = false
				res.Errorcode = 403
				res.Description = fh.Filename + "not allowed timeout be uploaded"
				break
			}

			file, err := uploadFile(fh)
			if err != nil {
				res.Success = false
				res.Errorcode = 500
				res.Description = err.Error()
				break
			}
			res.Files = append(res.Files, file)
		}

		var fsum int64
		for _, f := range res.Files {
			fsum += f.Size
		}

		if fsum > maxf {
			for _, f := range res.Files {
				os.Remove(udir + f.Uname)
				os.Remove(hdir + f.Hash)
			}

			res.Success = false
			res.Errorcode = 400
			res.Description = "File(s) above size limit"
			res.Files = nil
		}
	}

	if res.Success == true {
		for _, f := range res.Files {
			f.timeout = timeout
			files <- f
		}
	}

	if res.Success {
		switch req.URL.Query().Get("output") {
		case "gyazo":
			w.Header().Set("Content-Type", "text/plain")
			var urls []string
			for _, f := range res.Files {
				urls = append(urls, f.Url.String())
			}
			fmt.Fprintf(w, "%s", strings.Join(urls, "\n"))
		case "text":
			w.Header().Set("Content-Type", "text/plain")
			var urls []string
			for _, f := range res.Files {
				urls = append(urls, f.Url.String())
			}
			fmt.Fprintf(w, "%s\n", strings.Join(urls, "\n"))
		case "html":
			w.Header().Set("Content-Type", "text/html")
			t.ExecuteTemplate(w, "files", res)
		case "csv":
			w.Header().Set("Content-Type", "text/cvs")
			fmt.Fprint(w, "name,url,hash,size\n")
			csvw := csv.NewWriter(w)
			for _, f := range res.Files {
				csvw.Write([]string{
					f.Name,
					f.Url.String(),
					f.Hash,
					fmt.Sprintf("%d", f.Size),
				})
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
