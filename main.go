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
var model string

func initFlags() {
	flag.StringVar(&anki_path, "anki_path", "/Applications/Anki.app", "Path to Anki application. Defaults to /Applications/Anki.app")
	flag.StringVar(&kobo_path, "kobo_path", "/Volumes/KOBOeReader/.kobo/KoboReader.sqlite", "Path to Kobo application. Defaults to /Volumes/KOBOeReader/.kobo/KoboReader.sqlite")
	flag.StringVar(&deck, "deck", "My Words", "Name of anki deck to add words to")
	flag.StringVar(&model, "model", "Basic", "Anki card model to use for newly added notes")
}

func main() {
	initFlags()
	flag.Parse()
	startAnki(anki_path)
	words, err := kobo.GetWords(kobo_path)
	if err != nil {
		log.Fatalf("Kobo not connected or path to kobo is incorrect")
	}

	canAdd := AnkiCanAddWordsToDeck(words, deck, model)
	toAdd := make([]string, 0)
	for i := range words {
		if canAdd[i] {
			toAdd = append(toAdd, words[i])
		}
	}

	fmt.Println("New words to add:\n", toAdd)

	m, err := dict.NewMerriam()
	if err != nil {
		log.Fatalln(err)
	}

	defs := m.GetDefinitions(toAdd)
	notes := CreateNotes(defs, deck, model)

	AddNotes(notes)
}
