package main

import (
	"flag"
	"fmt"
	"koboToAnki/m/dict"
	"koboToAnki/m/kobo"
	"log"

	_ "modernc.org/sqlite"
)

var anki_path string
var kobo_path string
var deck string

const MODEL = "Basic"

func initFlags() {
	flag.StringVar(&deck, "deck", "My Words", "Name of anki deck to add words to")
}

func main() {
	initFlags()
	flag.Parse()

	// Find kobo path and extract words from kobo sqlite db
	words := kobo.GetWords()

	// Start Anki, create deck, and check which words are new
	startAnki()
	AnkiCreateDeck(deck)
	toAdd := AnkiCanAddWordsToDeck(words, deck, MODEL)

	fmt.Println("New words to add:\n", toAdd)

	// Initialize dictionary object
	m, err := dict.NewMerriam()
	if err != nil {
		log.Fatalln(err)
	}

	// Fetch definitions from dictionary
	defs := m.GetDefinitions(toAdd)
	// Create Anki Notes
	notes := CreateNotes(defs, deck, MODEL)

	// Add notes to Anki via AnkiConnect
	AddNotes(notes)
}
