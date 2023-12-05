package main

import (
	"fmt"
	"net/http"
)

func main() {
	// exit := make(chan os.Signal, 1)
	// signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	// Serve the "static" folder at the base URL ("/")
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.Handle("static/project06.css", http.FileServer(http.Dir("./")))

	// Start the HTTP server in a goroutine
	go func() {
		fmt.Println("Starting HTTP server on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Println(err)
		}

	}()

	url := "https://openai.com/robots.txt"
	StopWords = createSWmap("stopwords-en.json")
	ebook := Index{}
	ebook.initializeDatabase(url)
	ebook.createRobotMap(url)
	fmt.Println("Finished crawling all urls.")

	// when the server reaches the /search url, use the function search
	http.HandleFunc("/search", ebook.searchHandlerDatabase)

	// 	<-exit
	// 	log.Println("Shutting down server.")
}
