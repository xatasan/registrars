package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"
)

type pQueue []*File

// implement heap.Interface (and sort.Interface)
func (pq *pQueue) Len() int {
	return len(*pq)
}

func (pq *pQueue) Less(i, j int) bool {
	if i >= 0 && j >= 0 {
		return (*pq)[i].timeout.Before((*pq)[j].timeout)
	}
	return false
}

func (pq *pQueue) Swap(i, j int) {
	if i >= 0 && j >= 0 {
		(*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i]
	}
}

func (pq *pQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*File))
}

func (pq *pQueue) Pop() interface{} {
	if len(*pq) == 0 {
		return nil
	}
	top := (*pq)[len(*pq)-1]
	*pq = (*pq)[0 : len(*pq)-1]
	return top
}

// implement io.{ReaderFrom, WriterTo}
func (pq pQueue) ReaderFrom(r io.Reader) (n int64, err error) {
	scan := bufio.NewScanner(r)
	if err = scan.Err(); err != nil {
		return
	}
	for scan.Scan() {
		if err = scan.Err(); err != nil {
			return
		}
		file := toFile(scan.Text())
		if time.Now().Before(file.timeout) {
			pq = append(pq, file)
			n += int64(len(scan.Text()))
		}
	}
	heap.Init(&pq)
	return
}

func (pq pQueue) WriteTo(w io.Writer) (n int64, err error) {
	writer := bufio.NewWriter(w)
	var N int
	for _, f := range pq {
		N, err = writer.WriteString(f.String())
		if err != nil {
			return
		}
		n += int64(N)
	}
	return
}

var (
	files  chan<- *File
	hashes map[string][]string
	queue  pQueue
	tolog  *os.File
)

func init() {
	istat, err := os.Stdin.Stat()
	if err != nil {
		log.Fatalln(err)
	}
	if (istat.Mode() & os.ModeCharDevice) == 0 {
		_, err := queue.ReaderFrom(os.Stdin)
		if err != nil {
			log.Println(err)
		}
	} // only read if piped input

	ostat, err := os.Stdin.Stat()
	if err != nil {
		log.Fatalln(err)
	}
	if (ostat.Mode() & os.ModeCharDevice) != 0 {
		tolog, err = os.Create(fmt.Sprintf("timeout.%d", os.Getpid()))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		tolog = os.Stdout
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		tolog.Close()
		os.Exit(1)
	}()

	fc := make(chan *File)
	files = fc
	go timeoutWorker(fc, &queue)

	hashes = make(map[string][]string)
	udirf, err := ioutil.ReadDir(udir)
	if err != nil {
		log.Panicln(err)
	}

	for _, file := range udirf {
		hash, err := os.Readlink(udir + file.Name())
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Sscanf(hash, hdir+"%s", &hash)
		hashes[hash] = append(hashes[hash], file.Name())
	}
}

func timeoutWorker(files <-chan *File, q *pQueue) {
	var (
		timer *time.Timer
		next  time.Time

		waitOnNext func()
	)

	waitOnNext = func() {
		deleteFile(heap.Pop(q).(*File))
		if q.Len() > 0 {
			time.AfterFunc(time.Until((*q)[0].timeout), waitOnNext)
		}
	}

	if q.Len() > 0 {
		timer = time.AfterFunc(time.Until((*q)[0].timeout), waitOnNext)
	}

	for file := range files {
		hashes[file.Hash] = append(hashes[file.Hash], file.Uname)
		if file.timeout.Before(time.Now()) {
			continue
		}

		if q.Len() >= 1 {
			next = (*q)[0].timeout
		} else {
			timer = time.AfterFunc(time.Until(file.timeout), waitOnNext)
		}
		heap.Push(q, file)

		if q.Len() >= 2 && (*q)[0].timeout.Before(next) {
			timer.Reset(time.Until((*q)[0].timeout))
		}
		tolog.WriteString(file.String())
	}
}

func deleteFile(file *File) {
	if err := os.Remove(udir + file.Uname); err != nil {
		log.Println(err)
	}

	if !keeptf {
		if len(hashes[file.Hash]) <= 1 {
			if err := os.Remove(hdir + file.Hash); err != nil {
				log.Println(err)
			}
			delete(hashes, file.Hash)
		} else {
			var pos int
			for i, f := range hashes[file.Hash] {
				if f == file.Uname {
					pos = i
					break
				}
			}

			l := len(hashes[file.Hash])
			if pos < l-1 {
				hashes[file.Hash][pos] = hashes[file.Hash][l-1]
			}
			hashes[file.Hash] = hashes[file.Hash][:l-1]
		}
	}
}
