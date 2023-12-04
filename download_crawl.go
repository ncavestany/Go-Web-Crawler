package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/kljensen/snowball"
)

type DownloadResult struct {
	body []byte
	err  error
}

type ExtractResult struct {
	words, hrefs, sentences []string
	title                   string
}

func (ebook *Index) downloadDatabase(url string, dlOutC chan DownloadResult) {
	ebook.mu.Lock()
	// fmt.Println("Now downloading " + url)
	delay := time.Duration(ebook.robots[".*"].delay)
	// If delay was not set, set to 100ms.
	if delay == 0 {
		delay = 100
	}
	// create the row for the current url
	ebook.addNewWordorUrl("urls", url)
	time.Sleep(delay * time.Millisecond)

	// get the contents of a given URL and return a slice of bytes
	if rsp, err := http.Get(url); err == nil {
		if bts, err := io.ReadAll(rsp.Body); err == nil {
			// Put the results from download into the download output channel
			dlOutC <- DownloadResult{body: bts, err: nil}
		}
	}
	ebook.mu.Unlock()
}

func (ebook *Index) recursiveCrawlDatabase(url string, crawledUrls *map[string]struct{}, wg *sync.WaitGroup) {
	// Add the current goroutine to the waitgroup
	wg.Add(1)
	// Creating channels to store the input/output of functions
	dlInC := make(chan string, 10000)
	dlOutC := make(chan DownloadResult, 10000)
	exOutC := make(chan ExtractResult, 10000)
	// Timeout channel to know when to stop crawling
	// Time was reduced until values were no longer consistent
	timeout := time.After(1 * time.Second)

	// Put the current url into the download input channel.
	dlInC <- url

	var valid bool
	for _, disallowedUrl := range ebook.robots[".*"].disallowed {
		if matched, err := regexp.MatchString(disallowedUrl, url); err == nil {
			if matched {
				// fmt.Println(url, "is not allowed.")
				valid = false
				break
			} else {
				// fmt.Println(url, "is allowed")
				valid = true
				(*crawledUrls)[url] = struct{}{}
			}
		}
	}

	for _, allowedUrl := range ebook.robots[".*"].allowed {
		if matched, err := regexp.MatchString(allowedUrl, url); err == nil {
			if matched {
				valid = true
				(*crawledUrls)[url] = struct{}{}
				break
			} else {
				valid = false
			}
		}
	}

	go func() {
		for {
			select {
			case url := <-dlInC:
				// fmt.Println("Downloading...")
				valid = true
				if valid {
					ebook.downloadDatabase(url, dlOutC)
				}
			case dl := <-dlOutC:
				// fmt.Println("Extracting...")
				extract(&dl, exOutC)
			case ex := <-exOutC:
				var exists bool
				urlID := ebook.findID("urls", url)
				err := ebook.db.QueryRow("SELECT EXISTS(SELECT 1 FROM frequency WHERE url_id=?)", urlID).Scan(&exists)
				if err != nil {
					log.Fatalf("Error in checking for existing row %v", err)
				}

				// If the current url already exists in the frequency table,
				// do not crawl its words again.
				if !exists {
					// for _, sentence := range ex.sentences {
					// 	// fmt.Println("Current sentence:" + sentence)
					// 	ebook.addSentence(sentence, url)
					// }

					for _, word := range ex.words {
						ebook.updateDatabase(word, url)
					}
					for i := 0; i < len(ex.words)-1; i++ {
						// fmt.Println(ex.words[i] + " " + ex.words[i+1])
						ebook.insertBigram(ex.words[i], ex.words[i+1], url)
					}
					ebook.addTitle(ex.title, url)
				} else {
					fmt.Println(url, "already exists.")
				}

				// Commenting out crawling functionality
				// for _, currentUrl := range ex.hrefs {
				// 	cleanedUrl := clean(url, currentUrl)
				// 	if cleanedUrl != "error" {
				// 		crawled := false
				// 		if _, exists := (*crawledUrls)[cleanedUrl]; exists {
				// 			crawled = true
				// 			break
				// 		}
				// 		if !crawled {
				// 			ebook.recursiveCrawlDatabase(cleanedUrl, crawledUrls, wg)
				// 		}
				// 	}
				// }
			case <-timeout:
				// fmt.Println("Leaving.")
				// Only close the waitgroup after the timeout
				defer wg.Done()
				return
			}
		}
	}()
}

func (ebook *Index) updateDatabase(word string, url string) {
	if stemmedWord, err := snowball.Stem(word, "english", true); err == nil {
		// If the stemmed word is not in the stopword map, then add it.
		if _, exists := StopWords[stemmedWord]; !exists {
			ebook.addNewWordorUrl("words", stemmedWord)
			ebook.addOccurrence(url, stemmedWord)
		}
	}
}

func (ebook *Index) crawlDatabase(hostUrl string) {
	crawledUrls := make(map[string]struct{})
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		ebook.recursiveCrawlDatabase(hostUrl, &crawledUrls, &wg)
	}()

	wg.Wait()
	fmt.Println("Finished crawling " + hostUrl)
}
