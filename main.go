// test-log-parse parses the test log and prints the failures.
package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	runPrefix  = "=== RUN   "
	passPrefix = "--- PASS: "
	failPrefix = "--- FAIL: "
	racePrefix = "WARNING: DATA RACE"
	undefined  = "undefined:"
	urlPrefix  = "https://travis-ci.org/grpc/grpc-go/jobs/"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("log url not found")
		os.Exit(1)
	}

	logID := os.Args[1]
	if strings.HasPrefix(logID, urlPrefix) {
		logID = strings.TrimPrefix(logID, urlPrefix)
	}
	url := fmt.Sprintf("https://api.travis-ci.org/jobs/%v/log.txt?deansi=true", logID)

	resp, err := http.Get(url)
	// resp, err := http.Get("https://api.travis-ci.org/jobs//log.txt?deansi=true")
	if err != nil {
		log.Fatalf("failed to get txt log: %v", err)
	}
	defer resp.Body.Close()

	testStarted := make(map[string]bool)
	testFailed := make(map[string]int)
	raceCount := 0
	undefCount := 0

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), runPrefix) {
			testName := strings.TrimPrefix(scanner.Text(), runPrefix)
			testStarted[testName] = true
		} else if strings.HasPrefix(scanner.Text(), passPrefix) {
			testName := strings.Split(strings.TrimPrefix(scanner.Text(), passPrefix), " ")[0]
			delete(testStarted, testName)
		} else if strings.HasPrefix(scanner.Text(), failPrefix) {
			testName := strings.Split(strings.TrimPrefix(scanner.Text(), failPrefix), " ")[0]
			delete(testStarted, testName)
			testFailed[testName]++
		} else if strings.HasPrefix(scanner.Text(), racePrefix) {
			raceCount++
		} else if strings.Contains(scanner.Text(), undefined) {
			undefCount++
		}
	}
	if err := scanner.Err(); err != nil {
		log.Print("scanner.Err = ", err)
	}

	fmt.Println()
	fmt.Println("rawLog: ", url)
	fmt.Println()
	fmt.Println(fmt.Sprintf("url: https://travis-ci.org/grpc/grpc-go/jobs/%v", logID))
	fmt.Println()

	if len(testStarted) > 0 {
		fmt.Println("tests started but did not finish:")
		for t := range testStarted {
			fmt.Printf(" - %v\n", t)
		}
	}

	if len(testFailed) > 0 {
		fmt.Println("tests failed:")
		for t := range testFailed {
			fmt.Printf(" - %v\n", t)
		}
	}

	if raceCount > 0 {
		fmt.Println("number of races: ", raceCount)
	}

	if undefCount > 0 {
		fmt.Println("number of undefined: ", undefCount)
	}
}
