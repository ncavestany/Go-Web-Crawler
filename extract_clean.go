package main

import (
	"bytes"
	"log"
	"net/url"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

func extract(exInC *DownloadResult, exOutC chan ExtractResult) {
	var result ExtractResult

	reader := bytes.NewReader(exInC.body)

	// Parse the HTML content
	doc, err := html.Parse(reader)
	if err != nil {
		log.Fatalf("Could not parse html %v", err)
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		switch n.Type {
		case html.ElementNode:
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					result.hrefs = append(result.hrefs, attr.Val)
				}
			}
			// Extracting the title name
			if n.Data == "title" && n.Parent.Data == "head" {
				// fmt.Println("Current url title:" + strings.TrimSpace(n.FirstChild.Data))
				result.title = strings.TrimSpace(n.FirstChild.Data)
			}
		case html.TextNode:
			p := n.Parent
			if p.Type == html.ElementNode && (p.Data != "style" && p.Data != "script") {
				newWords := strings.FieldsFunc(n.Data, func(r rune) bool {
					return !unicode.IsLetter(r) && !unicode.IsNumber(r)
				})
				result.words = append(result.words, newWords...)
			}
		}
		// go through the child nodes recursively
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	// Put the results into the extract output channel
	exOutC <- result
}

func clean(host string, href string) string {
	var err error

	if len(href) == 0 {
		return "error"
	}

	// If the link has backslash at the end, remove it for concatenation later
	if href[len(href)-1] == '/' {
		href = href[:len(href)-1]
	}

	// I cannot figure out how to delete DS_Store.
	if strings.Contains(href, "DS_Store") {
		return "error"
	}

	if hostUrl, err := url.Parse(host); err == nil {
		if parsedUrl, err := url.Parse(href); err == nil {
			// If the link starts with a backslash (is part of a complete link)
			// then just add the hostUrl's scheme and host
			if len(href) != 0 && href[0] == '/' {
				parsedUrl.Scheme = hostUrl.Scheme
				parsedUrl.Host = hostUrl.Host
				return parsedUrl.String()
			}

			// Do not crawl pictures.
			if strings.Contains(href, "jpg") || strings.Contains(href, "png") {
				return "error"
			}

			// if it is an incomplete url (html file), add it to the host URL
			if parsedUrl.Scheme == "" && parsedUrl.Path != "" {
				return host + "/" + parsedUrl.String()
			}

			// ADDED: if the new url is different from the initial host url, ignore it
			// to prevent crawling beyond the initial host and return an error
			if parsedUrl.Host != hostUrl.Host {
				return "error"
			}

			return parsedUrl.String()
		}
	}

	if err != nil {
		log.Fatalf("Url could not be parsed %v", err)
	}
	return ""
}
