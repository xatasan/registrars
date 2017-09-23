package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func fileTimeout(filename, hashsum string) func() {
	return func() {
		if err := os.Remove(udir + filename); err != nil {
			log.Println(err)
		}

		if !keeptf {
			files, err := ioutil.ReadDir(udir)
			if err != nil {
				log.Panicln(err)
			}
			for _, file := range files {
				hash, err := os.Readlink(udir + file.Name())
				if err != nil {
					log.Println(err)
					continue
				}
				fmt.Sscanf(hash, hdir+"%s", &hash)

				if hash == hashsum && filename != file.Name() {
					return
				}
			}

			if err := os.Remove(hdir + hashsum); err != nil {
				log.Println(err)
			}
		}
	}
}

func recordFile(filename, hashsum string, timeout time.Duration) {
	if timeout > 0 {
		fmt.Printf("%d\t%s\t%s\n", time.Now().Add(timeout).Unix(), filename, hashsum)
		time.AfterFunc(timeout, fileTimeout(filename, hashsum))
	}
}

func regenFrom(inp io.Reader) {
	input := bufio.NewScanner(inp)
	line := 1
	for input.Scan() {
		parts := strings.SplitN(input.Text(), "\t", 2)
		if len(parts) != 3 {
			log.Printf("Line %d invalid: more or less than 3 columns\n", line)
		}

		rto, err := strconv.Atoi(parts[0])
		if err != nil {
			log.Printf("Line %d invalid: %s\n", line, err.Error())
			continue
		}
		filename := parts[1]
		hashsum := parts[2]

		to := time.Unix(int64(rto), 0)
		if time.Now().Before(to) {
			fmt.Printf("%d\t%s\t%s\n", rto, filename, hashsum)
		}

		time.AfterFunc(to.Sub(time.Now()), fileTimeout(filename, hashsum))
		line++
	}
}
