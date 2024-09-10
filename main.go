package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"sync"
)

func main() {
	args := os.Args[1:]
	if len(args) < 3 {
		fmt.Println("not enough arguments provided")
		fmt.Printf("Correct usage:\ncrawler [url] [maxConcurrency] [maxPages]")
		os.Exit(1)
	} else if len(args) > 3 {
		fmt.Println("too many arguments provided")
		fmt.Printf("Correct usage:\ncrawler [url] [maxConcurrency] [maxPages]")
		os.Exit(1)
	}
	baseURL := args[0]
	for baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	url, err := url.Parse(baseURL)
	if err != nil {
		fmt.Println("error parsing url:", err)
		os.Exit(1)
	}
	maxConcurrency, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("maxConcurrency (argument 2) must be an int")
		fmt.Printf("Correct usage:\ncrawler [url] [maxConcurrency] [maxPages]")
		os.Exit(1)
	}
	maxPages, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Println("maxPages (argument 3) must be an int")
		fmt.Printf("Correct usage:\ncrawler [url] [maxConcurrency] [maxPages]")
		os.Exit(1)
	}
	fmt.Println("starting crawl of: ", baseURL, "with maxConcurrency:", maxConcurrency, "and maxPages:", maxPages)
	cfg := config{
		pages:              make(map[string]int),
		baseURL:            url,
		concurrencyControl: make(chan struct{}, maxConcurrency),
		wg:                 &sync.WaitGroup{},
		my:                 &sync.Mutex{},
		maxPages:           maxPages,
	}
	cfg.wg.Add(1)
	go cfg.crawlPage(baseURL)
	cfg.wg.Wait()
	fmt.Println("Done...")
	printReport(cfg.pages, baseURL)
}
