package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

func recordFile(filename, hashsum string, timeout time.Duration) {
	if timeout > 0 {
		fmt.Printf("%s\t%s\n", time.Now().Add(timeout).Format(time.UnixDate), filename)
		time.AfterFunc(timeout, func() {
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
						log.Panicln(err)
					}
					fmt.Sscanf(hash, hdir+"%s", &hash)

					if hash == hashsum && filename != file.Name() {
						return //
					}
				}

				if err := os.Remove(hdir + hashsum); err != nil {
					log.Println(err)
				}
			}
		})
	}
}

func regenFrom(inp io.Reader) {
	input := bufio.NewScanner(inp)
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
}
