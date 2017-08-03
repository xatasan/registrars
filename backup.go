package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

func recordFile(filename string, timeout time.Duration) {
	if timeout > 0 {
		fmt.Printf("%s\t%s\n", time.Now().Add(timeout).Format(time.UnixDate), filename)
		time.AfterFunc(timeout, func() {
			if err := os.Remove(udir + filename); err != nil {
				log.Println(err)
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
