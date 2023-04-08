package main

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

func scanMB(num int, seg *[]SegmentDBH, sdbh chan int, cmd chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	if debug == 1 {

		log.Printf("Start proc %d\n", num)
	}
	for {
		select {
		case res := <-sdbh:
			getSegmentData(&segments[res])
		case res := <-cmd:
			if res == "shutdown" {
				if debug == 1 {
					log.Printf("Shutdown %d\n", num)
				}
				return
			}
		}
	}

}
func getSegmentData(seg *SegmentDBH) {
	var err error
	start := time.Now()
	// var stdout bytes.Buffer
	// var stderr bytes.Buffer

	if debug == 1 {
		fmt.Printf("MB:%s Path:%s\n", seg.place, seg.path)
	}

	res, err := exec.Command("/usr/bin/ssh", "-Y", seg.place, "du", "-sb", seg.path, "|", "cut", "-f", "1").Output()
	//res := int(rand.Float32() * 640 * 1024 * 10124)
	if debug == 1 {
		log.Printf("res: %s,%s\n%s\n", seg.place, seg.path, string(res))
	}
	if err != nil {
		log.Fatalln(err)
	}

	res2, err := exec.Command("/usr/bin/ssh", "-Y", seg.place, "find", seg.path, "-type", "f", "|", "wc", "-l").Output()
	//res2 := int(rand.Float32() * 1000)
	if debug == 1 {
		log.Printf("res: %s,%s\n%s\n", seg.place, seg.path, string(res2))
	}
	if err != nil {
		log.Fatalln(err)
	}

	elapsed := float64(time.Since(start)) / float64(time.Second)

	seg.elapsed = elapsed
	seg.segSize, err = strconv.Atoi(strings.TrimSpace(string(res)))
	if err != nil {
		seg.segSize = 0
	}
	seg.segFiles, err = strconv.Atoi(strings.TrimSpace(string(res2)))
	if err != nil {
		seg.segFiles = 0
	}
	if debug == 1 {
		log.Printf("result: %d,%d\n", seg.segSize, seg.segFiles)
	}
}
