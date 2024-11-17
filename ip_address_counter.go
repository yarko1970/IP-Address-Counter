package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

func bitMask(bitPosition int) (int, int64) {
	if bitPosition < 64 {
		return 0, 1 << bitPosition
	} else if bitPosition < 128 {
		return 1, 1 << (bitPosition - 64)
	} else if bitPosition < 192 {
		return 2, 1 << (bitPosition - 128)
	}
	return 3, 1 << (bitPosition - 192)
}

func numOnesInInt64(n int64) int {
	var count int = 0
	var tempInt int64 = n
	for i := 0; i < 64 && tempInt != 0; i++ {
		if tempInt%2 != 0 {
			count++
			tempInt = (tempInt - 1) / 2
		} else {
			tempInt = tempInt / 2
		}
	}
	return count
}

func lineProcessor(line string, ipRegistry *[256][256][256][4]int64) {
	result := strings.Split(line, ".")
	i0, _ := strconv.Atoi(result[0])
	i1, _ := strconv.Atoi(result[1])
	i2, _ := strconv.Atoi(result[2])
	i3, _ := strconv.Atoi(result[3])
	groupIndex, mask := bitMask(i3)
	ipRegistry[i0][i1][i2][groupIndex] |= mask
}

func subBatchProcessor(lines []string, ipRegistry *[256][256][256][4]int64) {
	if len(lines) == 0 {
		return
	}
	for i := 0; i < len(lines); i++ {
		lineProcessor(lines[i], ipRegistry)
	}
}

func batchProcessor(lines []string, ipRegistry *[256][256][256][4]int64) {

	if len(lines) < 100 {
		subBatchProcessor(lines, ipRegistry)
		return
	}

	var wg sync.WaitGroup

	var numOfProcs int = max(runtime.GOMAXPROCS(0)-1, 2)

	for i := 1; i < numOfProcs; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			subBatch := lines[(len(lines)*(i-1))/numOfProcs : (len(lines)*i)/numOfProcs]
			subBatchProcessor(subBatch, ipRegistry)
		}(i)
	}

	subBatch := lines[(len(lines)*(numOfProcs-1))/numOfProcs : (len(lines)*numOfProcs)/numOfProcs]
	subBatchProcessor(subBatch, ipRegistry)

	wg.Wait()
}

func uniqueIpCount(filename string) int {

	file, _ := os.Open(filename)
	defer file.Close()

	ipRegistry := [256][256][256][4]int64{}

	scanner := bufio.NewScanner(file)

	var lineCounter int = 0
	var reportFactor int = 100000000
	var batchMaxSize int = 1000000
	var currentlineInBatch int = 0
	var batchIsReady bool = true
	lines := []string{}

	for scanner.Scan() {

		var wg sync.WaitGroup

		batch := make([]string, len(lines))
		_ = copy(batch, lines)
		currentlineInBatch = 0
		lines = []string{}
		batchIsReady = false

		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			batchProcessor(batch, &ipRegistry)
		}(1)

		for !batchIsReady {
			ipAddress := scanner.Text()
			lines = append(lines, ipAddress)
			currentlineInBatch++
			if currentlineInBatch == batchMaxSize || !scanner.Scan() {
				batchIsReady = true
			}
			lineCounter++
			if lineCounter%reportFactor == 0 {
				fmt.Printf("Lines Scanned: %d Million\n", lineCounter/1000000)
			}
		}

		wg.Wait()
	}

	if len(lines) != 0 {
		batchProcessor(lines, &ipRegistry)
	}
	fmt.Printf("Lines Scanned: %d\n", lineCounter)

	var ipCounter int = 0
	for i0 := 0; i0 < 256; i0++ {
		for i1 := 0; i1 < 256; i1++ {
			for i2 := 0; i2 < 256; i2++ {
				for j := 0; j < 4; j++ {
					ipCounter += numOnesInInt64(ipRegistry[i0][i1][i2][j])
				}
			}
		}
	}

	return ipCounter
}

func main() {

	start := time.Now()

	uniqueIps := uniqueIpCount(os.Args[1])
	fmt.Printf("Total Unique IPs: %d\n", uniqueIps)

	elapsedTime := time.Since(start)
	fmt.Printf("Elapsed Time: %s\n", elapsedTime)

}
