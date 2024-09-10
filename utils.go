package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

type config struct {
	pages              map[string]int
	baseURL            *url.URL
	my                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
	maxPages           int
}

func normalizeURL(inputURL string) (string, error) {
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return "", err
	}
	return parsedURL.Scheme + "://" + parsedURL.Hostname() + parsedURL.Path, nil
}

func getURLsFromNode(node *html.Node) ([]string, error) {
	ret := []string{}
	if node.Type == html.ElementNode && node.Data == "a" {
		for _, attr := range node.Attr {
			if attr.Key == "href" {
				ret = append(ret, attr.Val)
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		child_urls, err := getURLsFromNode(child)
		if err != nil {
			return nil, err
		}
		ret = append(ret, child_urls...)
	}
	return ret, nil
}

func (cfg *config) getURLsFromHTML(htmlBody string) ([]string, error) {
	htmlReader := strings.NewReader(htmlBody)
	rootNode, err := html.Parse(htmlReader)
	if err != nil {
		return nil, err
	}
	rawUrls, err := getURLsFromNode(rootNode)
	if err != nil {
		return nil, err
	}
	i := 0
	for i < len(rawUrls) {
		parsedURL, err := url.Parse(rawUrls[i])
		if err != nil {
			return nil, err
		}
		abs := cfg.baseURL.ResolveReference(parsedURL)
		rawUrls[i] = abs.String()
		i++
	}
	return rawUrls, nil
}

func getHTML(rawURL string) (string, error) {
	res, err := http.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("http error: %d (%s)", res.StatusCode, res.Status)
	}
	if !strings.Contains(res.Header.Get("content-type"), "text/html") {
		return "", fmt.Errorf("response body not html")
	}
	html, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(html), nil
}

func (cfg *config) addPageVisit(normalizedURL string) bool {
	cfg.my.Lock()
	defer cfg.my.Unlock()
	if _, ok := cfg.pages[normalizedURL]; ok {
		cfg.pages[normalizedURL]++
		return false
	} else {
		cfg.pages[normalizedURL] = 1
		return true
	}
}

func (cfg *config) crawlPage(rawCurrentURL string) {
	cfg.concurrencyControl <- struct{}{}
	defer cfg.wg.Done()
	if cfg.maxPages < len(cfg.pages) {
		_ = <-cfg.concurrencyControl
		return
	}
	if !strings.Contains(rawCurrentURL, cfg.baseURL.Hostname()) {
		_ = <-cfg.concurrencyControl
		return
	}
	// if rawURL is relative append it to the base url
	normalizedCurrentURL, err := normalizeURL(rawCurrentURL)
	if err != nil {
		// fmt.Println("error normalizing url:", err)
		_ = <-cfg.concurrencyControl
		return
	}
	if !cfg.addPageVisit(normalizedCurrentURL) {
		_ = <-cfg.concurrencyControl
		return
	}
	fmt.Println("crawling page:", rawCurrentURL)
	html, err := getHTML(rawCurrentURL)
	if err != nil {
		// fmt.Println("error getting html for page:", rawCurrentURL, "error:", err)
		_ = <-cfg.concurrencyControl
		return
	}
	urls, err := cfg.getURLsFromHTML(html)
	if err != nil {
		// fmt.Println("error getting urls from html:", err)
		_ = <-cfg.concurrencyControl
		return
	}
	fmt.Println("Got urls: ", urls)
	for _, url := range urls {
		cfg.wg.Add(1)
		go cfg.crawlPage(url)
	}
	_ = <-cfg.concurrencyControl
	return
}

type Page struct {
	url   string
	links int
}

func printReport(pages map[string]int, baseURL string) {
	var pagesl []Page
	for k, v := range pages {
		pagesl = append(pagesl, Page{k, v})
	}
	sort.Slice(pagesl, func(i, j int) bool {
		return pagesl[i].links < pagesl[j].links
	})
	fmt.Printf("=============================\nREPORT for %s\n=============================\n", baseURL)
	for _, page := range pagesl {
		fmt.Printf("Found %d internal links to %s\n", page.links, page.url)
	}
}
