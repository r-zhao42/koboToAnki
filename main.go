package main

import (
	"flag"
	"fmt"
	"koboToAnki/dict"
	"koboToAnki/kobo"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

var isSetDeck string
var isSetMerriam string

const MODEL = "Basic"

func initFlags() {
	//TODO: store this in some env file so people don't have to specify everytime
	flag.StringVar(&isSetDeck, "setDeck", "", "Set Anki deck name for words to be added to.")
	flag.StringVar(&isSetMerriam, "setMerriam", "", "Set Merriam api key in env")
}

func main() {
	initFlags()
	flag.Parse()

	_, err := os.Stat(".env")
	var env map[string]string
	if os.IsNotExist(err) {
		env = make(map[string]string)
		env["Deck"] = "My Words"
		godotenv.Write(env, ".env")
	} else {
		fs, err := os.Open(".env")
		if err != nil {
			log.Fatalf("Failed to open .env file")
		}

		env, err = godotenv.Parse(fs)
		if err != nil {
			log.Fatalf("Failed to parse .env file")
		}
	}

	if isSetMerriam != "" {
		if len(os.Args) < 3 {
			log.Fatalf("Must provide api key value as an arg")
		}

		env["MERRIAM_API_KEY"] = os.Args[2]
		godotenv.Write(env, ".env")
		return
	}

	if isSetDeck != "" {
		if len(os.Args) < 3 {
			log.Fatalf("Must provide deck name as an arg")
		}

		env["Deck"] = os.Args[2]
		godotenv.Write(env, ".env")
		return
	}

	// load env
	err = godotenv.Load(".env")
	if err != nil {
		panic(".env file not found. Remember to set api keys first with setKey flag")
	}

	// Find kobo path and extract words from kobo sqlite db
	words := kobo.GetWords()

	// Start Anki, create deck, and check which words are new
	startAnki()
	deck := os.Getenv("Deck")
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
