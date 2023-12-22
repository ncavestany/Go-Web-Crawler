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

## Screenshots

**Homepage**
 ![Homepage](https://cdn.discordapp.com/attachments/428973514557751297/1187618245168074872/image.png?ex=65978aae&is=658515ae&hm=1ae70fb62530c18fc88301b407721e5ccc2ba726d99be29bff05fb31a389a53e& "Homepage")

**Searching**
![Searching](https://cdn.discordapp.com/attachments/428973514557751297/1187618294094639214/image.png?ex=65978aba&is=658515ba&hm=5dd2c59b761cc8b107db387c04785cffb99b266d5bac2235acda659593a732c3& "Searching")

**Bigram Search**
![Bigram Search](https://cdn.discordapp.com/attachments/428973514557751297/1187618369558564885/image.png?ex=65978acc&is=658515cc&hm=717d3a31c684011760a54c3316b339debba2b3cfe6e13ea560c1e04d6b2376d5& "Bigram Search")

**Wildcard Search**
![Wildcard Search](https://cdn.discordapp.com/attachments/428973514557751297/1187618446528229376/image.png?ex=65978ade&is=658515de&hm=496176d6025d3017661f13237bed27bff03659a0487423eb28401b2b1b366110& "Wildcard Search")
