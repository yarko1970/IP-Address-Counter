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

// generating bit mask for updating IP registry value
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

// calculating number of 1-bits in int64
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

// checking if number is between 0 and 255
func inRange(n int) bool {
	if (n >= 0) && (n <= 255) {
		return true
	}
	return false
}

// processing single line
func lineProcessor(line string, ipRegistry *[256][256][256][4]int64) {
	result := strings.Split(line, ".")
	i0, err0 := strconv.Atoi(result[0])
	i1, err1 := strconv.Atoi(result[1])
	i2, err2 := strconv.Atoi(result[2])
	i3, err3 := strconv.Atoi(result[3])
	if err0 != nil || err1 != nil || err2 != nil || err3 != nil {
		return
	}
	if !inRange(i0) || !inRange(i1) || !inRange(i2) || !inRange(i3) {
		return
	}
	groupIndex, mask := bitMask(i3)
	ipRegistry[i0][i1][i2][groupIndex] |= mask
}

// processing sub-batch of lines
func subBatchProcessor(subBatch []string, ipRegistry *[256][256][256][4]int64) {
	if len(subBatch) == 0 {
		return
	}
	for i := 0; i < len(subBatch); i++ {
		lineProcessor(subBatch[i], ipRegistry)
	}
}

// processing batch of lines
func batchProcessor(batch []string, ipRegistry *[256][256][256][4]int64) {

	if len(batch) < 1000 {
		subBatchProcessor(batch, ipRegistry)
		return
	}

	var wg sync.WaitGroup

	// setting the number of concurrent batch processes depending on available resources
	var numOfProcs int = max(runtime.GOMAXPROCS(0)-1, 2)

	// spawning and running concurrent sub-batch processing Goroutines
	for i := 1; i < numOfProcs; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			subBatch := batch[(len(batch)*(i-1))/numOfProcs : (len(batch)*i)/numOfProcs]
			subBatchProcessor(subBatch, ipRegistry)
		}(i)
	}

	subBatch := batch[(len(batch)*(numOfProcs-1))/numOfProcs : (len(batch)*numOfProcs)/numOfProcs]
	subBatchProcessor(subBatch, ipRegistry)

	wg.Wait()
}

// calculating the number of unique IPs
func uniqueIpCount(filename string) int {

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// creating and initializing IP registry
	ipRegistry := [256][256][256][4]int64{}
	for i0 := 0; i0 < 256; i0++ {
		for i1 := 0; i1 < 256; i1++ {
			for i2 := 0; i2 < 256; i2++ {
				for j := 0; j < 4; j++ {
					ipRegistry[i0][i1][i2][j] = 0
				}
			}
		}
	}

	scanner := bufio.NewScanner(file)

	var lineCounter int = 0
	var reportFactor int = 100000000
	var batchMaxSize int = 1000000
	var currentlineInBatch int = 0
	var batchIsReady bool = true
	lines := []string{}

	// scanning the file until complete
	for scanner.Scan() {

		var wg sync.WaitGroup

		batch := make([]string, len(lines))
		_ = copy(batch, lines)
		currentlineInBatch = 0
		lines = []string{}
		batchIsReady = false

		wg.Add(1)

		// spawning Goroutine to process the ready batch
		go func(id int) {
			defer wg.Done()
			batchProcessor(batch, &ipRegistry)
		}(1)

		// scanning the file to assemble next batch
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

	// making sure leftover batch is processed
	if len(lines) != 0 {
		batchProcessor(lines, &ipRegistry)
	}
	fmt.Printf("Lines Scanned: %d\n", lineCounter)

	// calculating the total number of 1-bits in IP registry
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
