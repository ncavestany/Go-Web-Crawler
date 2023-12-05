package main

import (
	"encoding/json"
	"log"
	"os"
	"sort"

	"github.com/kljensen/snowball"
)

type TfIdfValue struct {
	URL      string
	Title    string
	Sentence string
	TfIdf    float64
}

type TfIdfSlice []TfIdfValue

func (s TfIdfSlice) Len() int           { return len(s) }
func (s TfIdfSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s TfIdfSlice) Less(i, j int) bool { return s[i].TfIdf > s[j].TfIdf }

func createSWmap(filepath string) map[string]struct{} {
	stopWordMap := make(map[string]struct{})
	var stopWords []string
	if data, err := os.ReadFile(filepath); err == nil {
		if err = json.Unmarshal(data, &stopWords); err == nil {
			for _, stopWord := range stopWords {
				stopWordMap[stopWord] = struct{}{}
			}
		} else {
			log.Fatalf("JSON could not be unmarshaled: %v", err)
		}
	} else {
		log.Fatalf("File could not be read:  %v\n", err)
	}
	return stopWordMap
}

// Returns the amount of times that the word occurs in the given url.
func (ebook *Index) getOccurrences(urlID, wordID int) int {
	var occurrences int
	err := ebook.queries.getFreq.QueryRow(urlID, wordID).Scan(&occurrences)
	if err != nil {
		log.Fatalf("Could not find total occurrences %v", err)
	}
	return occurrences
}

// Returns the total amount of words in this url.
func (ebook *Index) getTotalUrlWords(urlID int) int {
	var occurrences int
	err := ebook.queries.getTotalUrlWords.QueryRow(urlID).Scan(&occurrences)
	if err != nil {
		log.Fatalf("Could not count total words in doc %v", err)
		return 0
	}
	return occurrences
}

// Returns the total amount of docs with the given word.
func (ebook *Index) getTotalDocsWithWord(wordID int) int {
	var count int
	err := ebook.queries.getTotalDocsWithWord.QueryRow(wordID).Scan(&count)
	if err != nil {
		log.Fatalf("Word could not be found in word table %v", err)
	}
	return count
}

// Returns the total amount of documents.
func (ebook *Index) getDocumentCount() int {
	var length int
	err := ebook.queries.getDocCount.QueryRow().Scan(&length)
	if err != nil {
		log.Fatalf("Could not count document table %v", err)
	}
	return length
}

// Returns a slice of all of the url_ids that a word appears in.
func (ebook *Index) getAllURLsForWord(wordID int) []int {
	var urlIDs []int
	rows, err := ebook.queries.getAllUrlsForWord.Query(wordID)
	if err != nil {
		log.Fatalf("Could not query when getting all urls of a word %v", err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var urlID int
		err := rows.Scan(&urlID)
		if err != nil {
			log.Fatalf("Could not scan through all rows %v", err)
		}
		urlIDs = append(urlIDs, urlID)
	}

	return urlIDs
}

// Returns a slice of all of the url_ids that a bigram appears in.
func (ebook *Index) getAllURLsForBigram(word1ID, word2ID int) []int {
	var urls []int
	rows, err := ebook.queries.getAllURLsForBigram.Query(word1ID, word2ID)
	if err != nil {
		log.Fatalf("Could not query when getting all urls of a bigram %v", err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var urlID int
		err := rows.Scan(&urlID)
		if err != nil {
			log.Fatalf("Could not scan through all rows %v", err)
		}
		urls = append(urls, urlID)
	}

	return urls
}

func (ebook *Index) getBigramOccurrences(word1ID, word2ID, urlID int) int {
	var occurrences int
	err := ebook.queries.getBigramsFreq.QueryRow(urlID, word1ID, word2ID).Scan(&occurrences)
	if err != nil {
		log.Fatalf("Could not find total occurrences %v", err)
	}
	return occurrences
}

func (ebook *Index) getTotalDocsWithBigram(word1ID, word2ID int) int {
	var count int
	err := ebook.queries.getTotalDocsForBigram.QueryRow(word1ID, word2ID).Scan(&count)
	if err != nil {
		log.Fatalf("Word could not be found in word table %v", err)
	}
	return count
}

// Given a url_id, returns the url in string form.
func (ebook *Index) getURL(urlID int) string {
	var url string
	err := ebook.queries.getURL.QueryRow(urlID).Scan(&url)
	if err != nil {
		log.Fatalf("Could not find url %v", err)
	}
	return url
}

// Given a url_id, returns the title in string form.
func (ebook *Index) getTitle(urlID int) string {
	var title string
	err := ebook.queries.getTitle.QueryRow(urlID).Scan(&title)
	if err != nil {
		log.Fatalf("Could not find url %v", err)
	}
	return title
}

func (ebook *Index) getSentence(sentenceID int) string {
	var sentence string
	err := ebook.queries.getSentence.QueryRow(sentenceID).Scan(&sentence)
	if err != nil {
		log.Fatalf("Could not find sentence %v", err)
	}
	return sentence
}

func (ebook *Index) getFreqSentence(urlID, wordID int) string {
	var sentenceID int
	var sentence string
	err := ebook.queries.getFreqSentence.QueryRow(urlID, wordID).Scan(&sentenceID)
	if err != nil {
		log.Fatalf("Could not find freq sentence %v", err)
	}

	sentence = ebook.getSentence(sentenceID)
	return sentence
}

func (ebook *Index) getBigramFreqSentence(urlID, word1ID, word2ID int) string {
	var sentenceID int
	var sentence string
	err := ebook.queries.getBigramFreqSentence.QueryRow(urlID, word1ID, word2ID).Scan(&sentenceID)
	if err != nil {
		log.Fatalf("Could not find bigram freq sentence %v", err)
	}

	sentence = ebook.getSentence(sentenceID)
	return sentence
}

// Print out the tf-idf value of a specific word on a specific url.
func (ebook *Index) getTfIdf(word, url string) float64 {
	urlID := ebook.findID("urls", url)
	// fmt.Println("ID for ", url, "is ", urlID)
	wordID := ebook.findID("words", word)
	// fmt.Println("ID for ", word, "is ", wordID)

	termOccurrencesinDoc := ebook.getOccurrences(urlID, wordID)
	if termOccurrencesinDoc == 0 {
		return 0
	}
	// fmt.Println("Amount of occurrences in doc:", termOccurrencesinDoc)

	totalWordsinDoc := ebook.getTotalUrlWords(urlID)
	// fmt.Println("Total words in ", url, ": ", totalWordsinDoc)

	docsWithWord := ebook.getTotalDocsWithWord(wordID)
	// fmt.Println("Total docs with word", word, ":", docsWithWord)

	documentCount := ebook.getDocumentCount()
	// fmt.Println("Total document count: ", documentCount)

	// TF is the total amount of words in the document divided by
	// the total amount of words in the document.
	TF := float64(termOccurrencesinDoc) / float64(totalWordsinDoc)
	// fmt.Println("TF:", TF)

	// DF is the amount of times the docs the word occurs in
	// divided by total amount of documents.
	DF := float64(docsWithWord) / float64(documentCount)
	// fmt.Println("DF:", DF)
	if DF == 0 {
		return 0
	}
	IDF := float64(1 / DF)
	return TF * IDF
}

// Sorts and returns a slice of tfIdf values.
func (ebook *Index) sortTfIdf(word string) (tfIdfValues []TfIdfValue) {
	stemmedTerm, _ := snowball.Stem(word, "english", true)
	wordID := ebook.findID("words", stemmedTerm)

	validURLIDs := ebook.getAllURLsForWord(wordID)

	for _, urlID := range validURLIDs {
		url := ebook.getURL(urlID)
		title := ebook.getTitle(urlID)
		sentence := ebook.getFreqSentence(urlID, wordID)
		tfIdf := ebook.getTfIdf(stemmedTerm, url)
		tfIdfValues = append(tfIdfValues, TfIdfValue{Title: title, URL: url, TfIdf: tfIdf, Sentence: sentence})
	}
	sort.Slice(tfIdfValues, func(i, j int) bool {
		if tfIdfValues[i].TfIdf == tfIdfValues[j].TfIdf {
			return tfIdfValues[i].URL > tfIdfValues[j].URL
		}
		return tfIdfValues[i].TfIdf > tfIdfValues[j].TfIdf
	})
	return tfIdfValues
}

// Print out the tf-idf value of a specific bigram on a specific url.
func (ebook *Index) getBigramTfIdf(word1, word2, url string) float64 {
	urlID := ebook.findID("urls", url)
	// fmt.Println("ID for ", url, "is ", urlID)
	word1ID, word2ID := ebook.findID("words", word1), ebook.findID("words", word2)
	// fmt.Println("ID for ", word, "is ", wordID)

	termOccurrencesinDoc := ebook.getBigramOccurrences(word1ID, word2ID, urlID)
	if termOccurrencesinDoc == 0 {
		return 0
	}
	// fmt.Println("Amount of occurrences in doc:", termOccurrencesinDoc)

	totalWordsinDoc := ebook.getTotalUrlWords(urlID)
	// fmt.Println("Total words in ", url, ": ", totalWordsinDoc)

	docsWithWord := ebook.getTotalDocsWithBigram(word1ID, word2ID)
	// fmt.Println("Total docs with bigram", word1, word2, ":", docsWithWord)

	documentCount := ebook.getDocumentCount()
	// fmt.Println("Total document count: ", documentCount)

	// TF is the total amount of words in the document divided by
	// the total amount of words in the document.
	TF := float64(termOccurrencesinDoc) / float64(totalWordsinDoc)
	// fmt.Println("TF:", TF)

	// DF is the amount of times the docs the word occurs in
	// divided by total amount of documents.
	DF := float64(docsWithWord) / float64(documentCount)
	// fmt.Println("DF:", DF)
	if DF == 0 {
		return 0
	}
	IDF := float64(1 / DF)
	return TF * IDF
}

// Sorts and returns a slice of tfIdf values.
func (ebook *Index) sortBigramTfIdf(word1, word2 string) (tfIdfValues TfIdfSlice) {
	word1ID, word2ID := ebook.findID("words", word1), ebook.findID("words", word2)

	validURLIDs := ebook.getAllURLsForBigram(word1ID, word2ID)

	for _, urlID := range validURLIDs {
		url := ebook.getURL(urlID)
		title := ebook.getTitle(urlID)
		sentence := ebook.getBigramFreqSentence(urlID, word1ID, word2ID)
		tfIdf := ebook.getBigramTfIdf(word1, word2, url)
		// fmt.Println("Current url:", title, "Current bigram:", word1, word2, "Current TF-IDF:", tfIdf)
		tfIdfValues = append(tfIdfValues, TfIdfValue{Title: title, URL: url, TfIdf: tfIdf, Sentence: sentence})
	}
	sort.Slice(tfIdfValues, func(i, j int) bool {
		if tfIdfValues[i].TfIdf == tfIdfValues[j].TfIdf {
			return tfIdfValues[i].URL > tfIdfValues[j].URL
		}
		return tfIdfValues[i].TfIdf > tfIdfValues[j].TfIdf
	})
	return tfIdfValues
}
