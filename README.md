# Web Crawler in Go

## Overview

This project is a versatile web crawler written in Go that enables users to extract and analyze textual content from websites. It utilizes goroutines for concurrent crawling, supports recursive crawling, and maintains a persistent database using SQLite. The extracted data is presented on a user-friendly HTML and CSS webpage. The key features include word extraction, bigram support, wildcard search, and TF-IDF-based result sorting.

## Features

### 1. Web Crawling

- **Concurrency:** The crawler employs goroutines to concurrently crawl websites, significantly speeding up the process.
- **Recursive Crawling:** Users can enable or disable recursive crawling, allowing for in-depth exploration of linked pages.

### 2. Database Integration

- **SQLite:** The crawler maintains a persistent database using SQLite to store extracted words and relevant metadata.

### 3. User Interface

- **HTML and CSS Webpage:** Results are presented on a simple and visually appealing HTML and CSS webpage.

### 4. Search Functionality

- **Word and Bigram Search:** Users can enter any word, including bigrams, to retrieve relevant results.
- **Wildcard Search:** A powerful feature that allows users to search for a base word and receive results that include variations (e.g., "water" yields "watercolor").

### 5. Result Sorting

- **TF-IDF Calculation:** Results are sorted using TF-IDF calculations, ensuring that the most relevant content appears first in the search results.


