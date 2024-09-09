package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("no website provided")
		os.Exit(1)
	} else if len(args) > 1 {
		fmt.Println("too many arguments provided")
		os.Exit(1)
	}
	baseURL := args[0]
	fmt.Println("starting crawl of: ", baseURL)
	pages := make(map[string]int)
	pages, err := crawlPage(baseURL, baseURL, pages)
	if err != nil {
		panic(err)
	}
	fmt.Println("Done")
	fmt.Println(pages)
}
