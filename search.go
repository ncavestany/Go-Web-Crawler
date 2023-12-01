package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/kljensen/snowball"
)

func (ebook *Index) searchDatabase(searchWord string) []string {
	var displayTfIdfValues []string
	tfIdfValues := ebook.sortTfIdf(searchWord)
	for _, currentTfIdfValues := range tfIdfValues {
		displayTfIdfValues = append(displayTfIdfValues, fmt.Sprint(currentTfIdfValues.Title)+" : "+fmt.Sprint(currentTfIdfValues.TfIdf))
	}

	return displayTfIdfValues
}

func (ebook *Index) wildcardSearch(searchWord string) (allTfIdfValues TfIdfSlice) {
	query := "SELECT id FROM words WHERE name LIKE ?"
	rows, err := ebook.db.Query(query, searchWord+"%")
	if err != nil {
		log.Fatalf("Could not query during wildcard search %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var wordID int
		err := rows.Scan(&wordID)
		if err == nil {
			word := ebook.getWord(wordID)
			fmt.Println("Current word:" + word)
			tfIdfValues := ebook.sortTfIdf(word)
			allTfIdfValues = append(allTfIdfValues, tfIdfValues...)
		}
	}

	sort.Sort(allTfIdfValues)
	return allTfIdfValues
}

func isBigram(query string) bool {
	words := strings.Fields(query)
	return len(words) == 2
}

func splitBigram(query string) (word1 string, word2 string) {
	words := strings.Fields(query)
	return words[0], words[1]
}

func (ebook *Index) bigramSearch(word1 string, word2 string) []string {
	var displayTfIdfValues []string
	tfIdfValues := ebook.sortBigramTfIdf(word1, word2)
	fmt.Println(word1, word2)

	for _, currentTfIdfValues := range tfIdfValues {
		displayTfIdfValues = append(displayTfIdfValues, fmt.Sprint(currentTfIdfValues.Title)+" : "+fmt.Sprint(currentTfIdfValues.TfIdf))
	}

	return displayTfIdfValues

}

// For searching bigram wildcards - example: computer scien% gives computer science and computer scientist.
func (ebook *Index) bigramWildcardSearch(word1, word2 string) (displayTfIdfValues TfIdfSlice) {
	var similarWordIDs []int
	var allTfIdfValues TfIdfSlice
	query := "SELECT id FROM words WHERE name LIKE ?"
	rows, err := ebook.db.Query(query, word2+"%")
	if err != nil {
		log.Fatalf("Could not query during wildcard search %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var wordID int
		err := rows.Scan(&wordID)
		if err == nil {
			similarWordIDs = append(similarWordIDs, wordID)
		}
	}
	word1ID := ebook.findID("words", word1)
	bigramWildcardQuery := "SELECT occurrences FROM bigrams WHERE word1_id=? AND word2_id=?"
	for _, word2IDs := range similarWordIDs {
		word2 := ebook.getWord(word2IDs)
		fmt.Println(word2)
		similarWordOccurrences, err := ebook.db.Query(bigramWildcardQuery, word1ID, word2IDs)
		if err == nil {
			allTfIdfValues = append(allTfIdfValues, ebook.sortBigramTfIdf(word1, word2)...)
		}
		defer similarWordOccurrences.Close()
	}
	sort.Sort(allTfIdfValues)
	return allTfIdfValues
}

func (ebook *Index) searchHandlerDatabase(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("static/template.html")
	if err != nil {
		log.Fatalf("Could not parse template files %v", err)
	}
	var tfIdfValues []TfIdfValue

	// localhost:8080/search?term=query
	query := r.URL.Query().Get("term")
	wildcard := r.URL.Query().Get("wildcard")

	if isBigram(query) {
		word1, word2 := splitBigram(query)
		stemmedWord1, stemmedWord2 := ebook.validateAndStemBigram(word1, word2)
		if wildcard != "" {
			tfIdfValues = ebook.bigramWildcardSearch(stemmedWord1, stemmedWord2)
		} else {
			tfIdfValues = ebook.sortBigramTfIdf(stemmedWord1, stemmedWord2)
		}

		if len(tfIdfValues) != 0 {
			// w.Write([]byte("Word: " + query + "\n"))
			err = t.Execute(w, tfIdfValues)
			if err != nil {
				log.Fatalf("Execute: %v", err)
			}
		} else {
			http.Error(w, "Word not found.", http.StatusNotFound)
		}
	} else {
		if stemmedQuery, err := snowball.Stem(query, "english", true); err == nil {
			fmt.Println(wildcard)
			if wildcard != "" {
				tfIdfValues = ebook.wildcardSearch(stemmedQuery)
			} else {
				tfIdfValues = ebook.sortTfIdf(stemmedQuery)
			}

			if err == nil && len(tfIdfValues) != 0 {
				// w.Write([]byte("Word: " + query + "\n"))
				// for _, tfIdfValue := range tfIdfValues {
				// 	w.Write([]byte(tfIdfValue + "\n"))
				// }
				err = t.Execute(w, tfIdfValues)
				if err != nil {
					log.Fatalf("Execute: %v", err)
				}
			} else {
				http.Error(w, "Word not found.", http.StatusNotFound)
			}

		} else {
			log.Fatalf("Error in word stemming %v", err)
		}
	}

}
