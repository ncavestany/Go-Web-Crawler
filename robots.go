package main

import (
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Urlset struct {
	Urls []string `xml:"url>loc"`
}

func (ebook *Index) createRobotMap(currentUrl string) {
	ebook.robots = make(map[string]rules)
	if parsedUrl, err := url.Parse(currentUrl); err == nil {
		// creating robots.txt url
		// localhost:8080/robots.txt
		robotsUrl := parsedUrl.Scheme + "://" + parsedUrl.Host + "/robots.txt"
		if rsp, err := http.Get(robotsUrl); err == nil {
			if robotsData, err := io.ReadAll(rsp.Body); err == nil {
				lines := strings.Split(string(robotsData), "\n")
				currentUser := ""

				for _, line := range lines {
					pieces := strings.Split(line, " ")
					if len(pieces) < 2 {
						continue
					}
					UAD := strings.ToLower(pieces[0])
					value := strings.TrimSpace(pieces[1])
					value = strings.ReplaceAll(value, "*", ".*")

					switch UAD {
					case "user-agent:":
						// fmt.Println("User agent:", value)
						currentUser = value
					case "allow:":
						// fmt.Println("Allow:", value)
						currentRules := ebook.robots[currentUser]
						currentRules.allowed = append(currentRules.allowed, value)
						ebook.robots[currentUser] = currentRules
					case "disallow:":
						// fmt.Println("Disallow:", value)
						currentRules := ebook.robots[currentUser]
						currentRules.disallowed = append(currentRules.disallowed, value)
						ebook.robots[currentUser] = currentRules
					case "crawl-delay:":
						// fmt.Println("Delay:", value)
						if num, err := strconv.Atoi(value); err == nil {
							currentRules := ebook.robots[currentUser]
							currentRules.delay = num
							ebook.robots[currentUser] = currentRules
						}
					case "sitemap:":
						ebook.downloadSitemap(value)
					}
				}
			} else {
				log.Fatalf("Could not read data %v", err)
			}
		} else {
			log.Fatalf("Could not download robots.txt %v", err)
		}
	} else {
		log.Fatalf("Error parsing %v", err)
	}
}

func (ebook *Index) downloadSitemap(sitemap string) {
	if response, err := http.Get(sitemap); err == nil {
		defer response.Body.Close()
		if xmlData, err := io.ReadAll(response.Body); err == nil {
			var urlset Urlset

			err = xml.Unmarshal(xmlData, &urlset)
			if err != nil {
				log.Fatalf("Could not unmarshal xml: %v", err)
			}

			for _, url := range urlset.Urls {
				ebook.crawlDatabase(url)
			}

		} else {
			log.Fatalf("Could not read response body %v", err)
		}

	} else {
		log.Fatalf("Could not http get sitemap %v", err)
	}
}
