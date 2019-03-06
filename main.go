package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type CacheFetcher struct {
	Body  string
	Urls  []string
	Error error
}

func (cacheFetcher CacheFetcher) Fetch(url string) (string, []string, error) {
	return cacheFetcher.Body, cacheFetcher.Urls, cacheFetcher.Error
}

var cacheURL = make(map[string]CacheFetcher)
var mutex sync.Mutex

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	channelQuit := make(chan bool)
	go func() {
		defer close(channelQuit)
		if depth <= 0 {
			return
		}
		var body string
		var urls []string
		var err error
		mutex.Lock()
		cacheFetcher, existingCacheFetcher := cacheURL[url]
		mutex.Unlock()
		if existingCacheFetcher {
			body, urls, err = cacheFetcher.Fetch(url)
		} else {
			body, urls, err = fetcher.Fetch(url)
			mutex.Lock()
			cacheURL[url] = CacheFetcher{
				Body:  body,
				Urls:  urls,
				Error: err,
			}
			mutex.Unlock()
		}
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("found: %s %q\n", url, body)
		for _, u := range urls {
			Crawl(u, depth-1, fetcher)
		}
		return
	}()
	for <-channelQuit {
	}
}

func main() {
	Crawl("https://golang.org/", 4, fetcher)
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
