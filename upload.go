package main

import (
	"crypto/sha256"
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
	Files       []*File `json:"files,omitempty"`
}

type File struct {
	Uname   string `json:"-"`
	Name    string `json:"name"`
	Url     string `json:"url"`
	Hash    string `json:"hash"`
	Size    uint64 `json:"size"`
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
	hash := sha256.New()
	_, err := io.Copy(hash, in)
	if err != nil {
		return "", "", err
	}
	var hbuf strings.Builder
	for _, b := range hash.Sum(nil) {
		hbuf.WriteString(fmt.Sprintf("%x", b))
	}
	hsum := hbuf.String()

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

func uploadData(file *os.File, orig, name, hash string, size uint64) (*File, error) {
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
			Url:   u.String(),
			Hash:  hash,
			Size:  uint64(size),
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
		Url:   u.String(),
		Hash:  hash,
		Size:  uint64(size),
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

	return uploadData(tmp, "paste.txt", name, hash, uint64(size))
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
	switch {
	case strings.HasSuffix(fh.Filename, ".tar.gz"):
		name += ".tar.gz"
	case strings.HasSuffix(fh.Filename, ".tar.bz2"):
		name += ".tar.bz2"
	case strings.HasSuffix(fh.Filename, ".tar.xz"):
		name += ".tar.xz"
	default:
		name += path.Ext(fh.Filename)
	}

	return uploadData(tmp, fh.Filename, name, hash, uint64(size))
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
			res.Description = "File above size limit " + byteSize(maxf)
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
			case ".exe", ".bat", ".cmd", ".msi", ".vbs", ".scr":
				res.Success = false
				res.Errorcode = 403
				res.Description = fh.Filename + " if of an invalid filetype"
				break
			}

		}

		if res.Success {
			for _, fh := range form.File["files"] {
				file, err := uploadFile(fh)
				if err != nil {
					res.Success = false
					res.Errorcode = 500
					res.Description = err.Error()
					break
				}
				res.Files = append(res.Files, file)
			}

			var fsum uint64
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
				res.Description = "File(s) above size limit " + byteSize(maxf)
				res.Files = nil
			}
		}
	}

	if res.Success {
		for _, f := range res.Files {
			f.timeout = timeout
			files <- f
		}

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
			for _, f := range res.Files {
				fmt.Fprintln(w, f.Url)
			}
		case "html":
			w.Header().Set("Content-Type", "text/html")
			t.ExecuteTemplate(w, "files", res)
		case "csv":
			w.Header().Set("Content-Type", "text/csv")
			fmt.Fprint(w, "name,url,hash,size\n")
			csvw := csv.NewWriter(w)
			for _, f := range res.Files {
				csvw.Write([]string{
					f.Name,
					f.Url,
					f.Hash,
					fmt.Sprintf("%d", f.Size),
				})
			}
		case "tsv":
			w.Header().Set("Content-Type", "text/tag-seperated-values")
			for _, f := range res.Files {
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", f.Name, f.Url, f.Hash, f.Size)
			}
		default:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(res)
		}
		return
	}

	if req.URL.Query().Get("output") == "html" {
		w.Header().Set("Content-Type", "text/html")
		t.ExecuteTemplate(w, "files", res)
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}
}
