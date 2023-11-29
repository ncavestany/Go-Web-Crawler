package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {

	// Serve the "static" folder at the base URL ("/")
	http.Handle("/", http.FileServer(http.Dir("static")))

	// Start the HTTP server in a goroutine
	go func() {
		fmt.Println("Starting HTTP server on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Println(err)
		}

	}()

	url := os.Args[1]
	StopWords = createSWmap("stopwords-en.json")
	ebook := Index{}
	ebook.initializeDatabase(url)
	ebook.createRobotMap(url)
	fmt.Println("Finished crawling all urls.")

	// when the server reaches the /search url, use the function search
	http.HandleFunc("/search", ebook.searchHandlerDatabase)
	// Use a loop to keep the program running
	for {
		time.Sleep(10 * time.Millisecond)
	}
}
