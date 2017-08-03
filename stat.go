package main

import (
	"io/ioutil"
	"log"
	"math"
	"os"
	"time"
)

type Data struct {
	S float64 // total storage used
	N int64   // amount of files hosted
	A float64 // average file size
	M float64 // mean file size
}

var data Data

func init() {
	var filec int64
	go func() {
		for {
			var storc, avgfs, meanfs float64
			all_files, err := ioutil.ReadDir(hdir)
			if err != nil {
				log.Fatalln(err)
			}
			var files []os.FileInfo
			for _, f := range all_files {
				if f.Size() > 0 {
					files = append(files, f)
				}
			}

			filec = int64(len(files))
			storc = 0
			for _, f := range files {
				storc += float64(f.Size())
			}
			avgfs = float64(storc / float64(len(files)))

			if len(files)%2 == 1 {
				midc := int(math.Ceil(float64(len(files) / 2)))
				midf := int(math.Floor(float64(len(files) / 2)))
				meanfs = float64(files[midc].Size()+files[midf].Size()) / 2
			} else if len(files) > 0 {
				mid := int(len(files) / 2)
				meanfs = float64(files[mid].Size())
			} else {
				meanfs = 0
				avgfs = 0
			}

			storc /= 1 << 20
			avgfs /= 1 << 20
			meanfs /= 1 << 20
			data = Data{storc, filec, avgfs, meanfs}

			time.Sleep(time.Minute * 30)
		}
	}()
}
