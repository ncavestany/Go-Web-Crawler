package main

import (
	"database/sql"
	"sync"
)

var StopWords map[string]struct{}

type Index struct {
	robots       map[string]rules
	db           *sql.DB
	queries      prepStatements
	mu           sync.Mutex
	databaseName string
}

type rules struct {
	allowed    []string
	disallowed []string
	delay      int
}

type prepStatements struct {
	insertURLTitle        *sql.Stmt
	insertWord            *sql.Stmt
	insertURL             *sql.Stmt
	getURLID              *sql.Stmt
	getURL                *sql.Stmt
	getWordID             *sql.Stmt
	getWord               *sql.Stmt
	getTitle              *sql.Stmt
	getFreq               *sql.Stmt
	getBigramsFreq        *sql.Stmt
	updateFreq            *sql.Stmt
	insertFreq            *sql.Stmt
	updateBigramsFreq     *sql.Stmt
	insertBigramsFreq     *sql.Stmt
	getTotalDocsWithWord  *sql.Stmt
	getTotalUrlWords      *sql.Stmt
	getDocCount           *sql.Stmt
	getAllUrlsForWord     *sql.Stmt
	getTotalDocsForBigram *sql.Stmt
	getAllURLsForBigram   *sql.Stmt
	getSentenceID         *sql.Stmt
	getSentence           *sql.Stmt
	getFreqSentence       *sql.Stmt
	getBigramFreqSentence *sql.Stmt
}
