package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/kljensen/snowball"
	_ "github.com/mattn/go-sqlite3"
)

func extractSubdomain(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return ""
	}

	// Split the hostname into parts using dots as separators
	parts := strings.Split(parsedURL.Hostname(), ".")
	// openai com
	if len(parts) == 2 {
		return parts[0]
	}

	// www openai com
	if len(parts) > 2 {
		return parts[1]
	}

	// If there's only one part, it's already the subdomain
	return parts[0]
}

// Create all the initial tables if they do not exist.
func (ebook *Index) initializeDatabase(url string) error {
	subDomain := extractSubdomain(url)
	ebook.databaseName = subDomain
	db, err := sql.Open("sqlite3", subDomain+".db")
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Create the table if it doesn't already exist.
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			id INTEGER NOT NULL PRIMARY KEY,
			name TEXT,
			title TEXT
		)
	`)
	if err != nil {
		log.Fatalf("Could not create urls table %v", err)
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS words (
			id INTEGER NOT NULL PRIMARY KEY,
			name TEXT
		)
	`)
	if err != nil {
		log.Fatalf("Could not create words table %v", err)
		return err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS sentences (
			id INTEGER NOT NULL PRIMARY KEY,
			sentence TEXT,
			url_id INTEGER,
			UNIQUE(sentence, url_id),
			FOREIGN KEY (url_id) REFERENCES urls(id)
		)
	`)
	if err != nil {
		log.Fatalf("Could not create or open sentences table %v", err)
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS frequency (
			id INTEGER NOT NULL PRIMARY KEY,
			url_id INTEGER,
			word_id INTEGER,
			sentence_id INTEGER,
			occurrences INTEGER,
			FOREIGN KEY (url_id) REFERENCES urls(id),
			FOREIGN KEY (word_id) REFERENCES words(id),
			FOREIGN KEY (sentence_id) REFERENCES sentences(id)
		)
	`)
	if err != nil {
		log.Fatalf("Could not create or open frequency table %v", err)
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bigrams (
			id INTEGER NOT NULL PRIMARY KEY,
			url_id INTEGER,
			word1_id INTEGER,
			word2_id INTEGER,
			sentence_id INTEGER,
			occurrences INTEGER,
			FOREIGN KEY (url_id) REFERENCES urls(id),
			FOREIGN KEY (word1_id) REFERENCES words(id),
			FOREIGN KEY (word2_id) REFERENCES words(id),
			FOREIGN KEY (sentence_id) REFERENCES sentences(id)
		)
	`)
	if err != nil {
		log.Fatalf("Could not create or open bigrams table %v", err)
		return err
	}

	ebook.db = db
	ebook.prepareStatements()

	return nil
}

func (ebook *Index) prepareStatements() {
	stmt := "UPDATE urls SET title=? WHERE name =?"
	addTitleStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare statement %v", err)
	}
	ebook.queries.insertURLTitle = addTitleStmt

	stmt = "INSERT INTO words (name) VALUES (?)"
	insertWordStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare statement %v", err)
	}
	ebook.queries.insertWord = insertWordStmt

	stmt = "INSERT INTO urls (name) VALUES (?)"
	insertURLStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare statement %v", err)
	}
	ebook.queries.insertURL = insertURLStmt

	stmt = "SELECT id FROM urls WHERE name=?"
	getURLIDStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare statement %v", err)
	}
	ebook.queries.getURLID = getURLIDStmt

	stmt = "SELECT name FROM urls WHERE id=?"
	getURLStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare statement %v", err)
	}
	ebook.queries.getURL = getURLStmt

	stmt = "SELECT id FROM words WHERE name=?"
	getWordIDStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare statement %v", err)
	}
	ebook.queries.getWordID = getWordIDStmt

	stmt = "SELECT name FROM words WHERE id=?"
	getWordStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare statement %v", err)
	}
	ebook.queries.getWord = getWordStmt

	stmt = "SELECT title FROM urls WHERE id=?"
	getTitleStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare statement %v", err)
	}
	ebook.queries.getTitle = getTitleStmt

	stmt = "SELECT occurrences FROM frequency WHERE url_id=? AND word_id=?"
	getFreqStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare statement %v", err)
	}
	ebook.queries.getFreq = getFreqStmt

	stmt = "SELECT occurrences FROM bigrams WHERE url_id=? AND word1_id=? AND word2_id=?"
	getBigramFreqStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare statement %v", err)
	}
	ebook.queries.getBigramsFreq = getBigramFreqStmt

	stmt = "UPDATE frequency SET occurrences=? WHERE url_id=? AND word_id=?"
	updateFreqStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare statement: %v", err)
	}
	ebook.queries.updateFreq = updateFreqStmt

	stmt = "INSERT INTO frequency (occurrences, url_id, word_id, sentence_id) VALUES (1, ?, ?, ?)"
	insertOccurrencesStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare statement: %v", err)
	}
	ebook.queries.insertFreq = insertOccurrencesStmt

	stmt = "INSERT INTO bigrams (occurrences, url_id, word1_id, word2_id, sentence_id) VALUES (1, ?, ?, ?, ?)"
	insertBigramFreqStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Error preparing insert statement: %v", err)
	}
	ebook.queries.insertBigramsFreq = insertBigramFreqStmt

	stmt = "UPDATE bigrams SET occurrences=? WHERE url_id=? AND word1_id=? AND word2_id=?"
	updateBigramFreqStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Error preparing update statement: %v", err)
	}
	ebook.queries.updateBigramsFreq = updateBigramFreqStmt

	stmt = "SELECT COUNT(*) FROM frequency WHERE word_id=?"
	getTotalDocsWithWordStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare statement: %v", err)
	}
	ebook.queries.getTotalDocsWithWord = getTotalDocsWithWordStmt

	stmt = "SELECT SUM(occurrences) FROM frequency WHERE url_id = ?"
	getTotalUrlWordsStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare statement: %v", err)
	}
	ebook.queries.getTotalUrlWords = getTotalUrlWordsStmt

	stmt = "SELECT COUNT(*) FROM urls"
	getDocCountStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare statement: %v", err)
	}
	ebook.queries.getDocCount = getDocCountStmt

	stmt = "SELECT url_id FROM frequency WHERE word_id = ?"
	getAllUrlsForWordStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare statement: %v", err)
	}
	ebook.queries.getAllUrlsForWord = getAllUrlsForWordStmt

	stmt = "SELECT url_id FROM bigrams WHERE word1_id = ? AND word2_id = ?"
	getAllUrlsForBigramStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare statement: %v", err)
	}
	ebook.queries.getAllURLsForBigram = getAllUrlsForBigramStmt

	stmt = "SELECT COUNT(*) FROM bigrams WHERE word1_id=? AND word2_id=?"
	getTotalDocsWithBigramStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare statement: %v", err)
	}
	ebook.queries.getTotalDocsForBigram = getTotalDocsWithBigramStmt

	stmt = "SELECT id FROM sentences WHERE sentence=?"
	getSentenceIDStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare statement: %v", err)
	}
	ebook.queries.getSentenceID = getSentenceIDStmt

	stmt = "SELECT sentence FROM sentences WHERE id=?"
	getSentenceStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare sentence select stmt: %v", err)
	}
	ebook.queries.getSentence = getSentenceStmt

	stmt = "SELECT sentence_id FROM frequency WHERE url_id=? AND word_id=?"
	getFreqSentenceStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare statement: %v", err)
	}
	ebook.queries.getFreqSentence = getFreqSentenceStmt

	stmt = "SELECT sentence_id FROM bigrams WHERE url_id=? AND word1_id=? AND word2_id=?"
	getBigramFreqSentenceStmt, err := ebook.db.Prepare(stmt)
	if err != nil {
		log.Fatalf("Could not prepare statement: %v", err)
	}
	ebook.queries.getBigramFreqSentence = getBigramFreqSentenceStmt

}

// Insert a unique word or url into the corresponding table.
func (ebook *Index) addNewWordorUrl(tableName string, name string) error {
	// First, check if the row already exists in the table
	var exists bool
	existsQuery := "SELECT EXISTS(SELECT 1 FROM " + tableName + " WHERE name=?)"
	err := ebook.db.QueryRow(existsQuery, name).Scan(&exists)
	if err != nil {
		log.Fatalf("Error during checking for existing value: %v", err)
		return err
	}

	if !exists {
		if tableName == "words" {
			_, err := ebook.queries.insertWord.Exec(name)
			if err != nil {
				log.Fatalf("Could not insert word %v", err)
			}
		} else if tableName == "urls" {
			_, err := ebook.queries.insertURL.Exec(name)
			if err != nil {
				log.Fatalf("Could not insert url %v", err)
			}
		}
		// fmt.Printf("Successfully inserted %s into the %s table.\n", name, tableName)
	}

	return nil
}

// Find the id of the selected word or url.
func (ebook *Index) findID(tableName string, name string) int {
	var id int
	if tableName == "words" {
		err := ebook.queries.getWordID.QueryRow(name).Scan(&id)
		if err != nil {
			return 0
		}
	} else if tableName == "urls" {
		err := ebook.queries.getURLID.QueryRow(name).Scan(&id)
		if err != nil {
			return 0
		}
	} else if tableName == "sentences" {
		err := ebook.queries.getSentenceID.QueryRow(name).Scan(&id)
		if err != nil {
			return 0
		}
	}
	return id
}

func (ebook *Index) getWord(wordID int) string {
	var word string
	err := ebook.queries.getWord.QueryRow(wordID).Scan(&word)
	if err != nil {
		log.Fatalf("Error getting word: %v", err)
	}

	return word
}

// Insert a new row of occurrences or update the amount of occurrences for an
// existing word.
func (ebook *Index) addOccurrence(urlID int, word, sentence string) error {
	wordID := ebook.findID("words", word)
	sentenceID := ebook.findID("sentences", sentence)
	// For testing purposes
	// url := ebook.getURL(urlID)

	// Check if the row already exists in the frequency table
	var hits int
	err := ebook.queries.getFreq.QueryRow(urlID, wordID).Scan(&hits)

	if err != nil {
		// If the word does not exist on the current url, create a new row for the new entry.
		if err == sql.ErrNoRows {
			_, err = ebook.queries.insertFreq.Exec(urlID, wordID, sentenceID)
			if err != nil {
				log.Fatalf("Could not insert into frequency table %v", err)
				return err
			}
			// Only noting the first sentence in the url the word was found on
			// fmt.Println("Word: " + word + " Sentence: " + sentence)
			// fmt.Printf("Successfully added occurrence with values (Occurrences: 1, url: %s, word: %s)\n", url, word)
		}
	} else {
		// If the word does exist on the current url, increment its amount of occurrences.
		hits++
		_, err = ebook.queries.updateFreq.Exec(hits, urlID, wordID)
		if err != nil {
			log.Fatalf("Could not update frequency table %v", err)
			return err
		}
		// fmt.Printf("Successfully updated occurrence to %d for URL: %s and Word: %s\n", hits, url, word)
	}

	return nil
}

func (ebook *Index) addTitle(title string, url string) {
	_, err := ebook.queries.insertURLTitle.Exec(title, url)
	if err != nil {
		log.Fatalf("Could not add title: %v", err)
	}
	fmt.Println("Setting title: " + title + " for url: " + url)
}

func (ebook *Index) addSentence(sentence string, urlID int) {
	insertQuery := "INSERT INTO sentences (sentence, url_id) VALUES (?, ?)"
	insertStmt, err := ebook.db.Prepare(insertQuery)
	if err != nil {
		log.Fatalf("Error preparing insert statement: %v", err)
	}
	defer insertStmt.Close()

	// fmt.Println("Inserting sentence:" + sentence + " into table at: " + url)
	insertStmt.Exec(sentence, urlID)
}

// Check if either half of the bigram is a stopword - if not stem both.
func (ebook *Index) validateAndStemBigram(word1 string, word2 string) (string, string) {
	if stemmedWord1, err := snowball.Stem(word1, "english", true); err == nil {
		if _, exists := StopWords[stemmedWord1]; !exists {
			if stemmedWord2, err := snowball.Stem(word2, "english", true); err == nil {
				if _, exists := StopWords[stemmedWord2]; !exists {
					return stemmedWord1, stemmedWord2
				}
			}
		}
	}
	return "", ""
}

func (ebook *Index) insertBigram(word1, word2, sentence string, urlID int) {
	stemmedWord1, stemmedWord2 := ebook.validateAndStemBigram(word1, word2)
	if stemmedWord1 != "" {
		// No need for error handling because these words will always exist in table.
		word1ID, word2ID := ebook.findID("words", stemmedWord1), ebook.findID("words", stemmedWord2)
		sentenceID := ebook.findID("sentences", sentence)
		// fmt.Println("Word 1 id: " + fmt.Sprintf("%d", word1ID))
		// fmt.Println("Word 2 id: " + fmt.Sprintf("%d", word2ID))

		// Check if the row already exists in the frequency table
		var hits int
		err := ebook.queries.getBigramsFreq.QueryRow(urlID, word1ID, word2ID).Scan(&hits)

		if err != nil {
			// If the bigram does not exist on the current url, create a new row for the new entry.
			if err == sql.ErrNoRows {
				_, err = ebook.queries.insertBigramsFreq.Exec(urlID, word1ID, word2ID, sentenceID)
				if err != nil {
					log.Fatalf("Could not insert into bigrams table %v", err)
				}
				// fmt.Printf("Successfully added occurrence with values (Occurrences: 1, url: %s, words: %s %s)\n", url, stemmedWord1, stemmedWord2)
			}
		} else {
			// If the bigram does exist on the current url, increment its amount of occurrences.
			hits++
			_, err = ebook.queries.updateBigramsFreq.Exec(hits, urlID, word1ID, word2ID)
			if err != nil {
				log.Fatalf("Could not update bigrams table %v", err)
			}
			// fmt.Printf("Successfully updated occurrence to %d for URL: %s and Words: %s %s\n", hits, url, stemmedWord1, stemmedWord2)
		}
	}
}
