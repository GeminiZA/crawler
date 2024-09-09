package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

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

func getURLsFromHTML(htmlBody, rawBaseURL string) ([]string, error) {
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
		if rawUrls[i] == "/" {
			rawUrls = append(rawUrls[:i], rawUrls[i+1:]...)
		}
		if rawUrls[i][0] == '/' {
			rawUrls[i] = rawBaseURL + rawUrls[i]
		}
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

func crawlPage(rawBaseURL, rawCurrentURL string, pages map[string]int) (map[string]int, error) {
	if !strings.Contains(rawCurrentURL, rawBaseURL) {
		return pages, nil
	}
	normalizedCurrentURL, err := normalizeURL(rawCurrentURL)
	if err != nil {
		return nil, err
	}
	if _, ok := pages[normalizedCurrentURL]; ok {
		pages[normalizedCurrentURL]++
		return pages, nil
	} else {
		pages[normalizedCurrentURL] = 1
	}
	html, err := getHTML(normalizedCurrentURL)
	if err != nil {
		return nil, err
	}
	fmt.Println("Got html for page:", normalizedCurrentURL)
	urls, err := getURLsFromHTML(html, normalizedCurrentURL)
	if err != nil {
		return nil, err
	}
	fmt.Println("Got urls:", urls)
	for _, url := range urls {
		crawlPage(rawBaseURL, url, pages)
	}
	return pages, nil
}
