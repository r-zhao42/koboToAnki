package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"koboToAnki/dict"
	"net/http"
	"os/exec"
	"path"
	"text/template"
	"time"
)

const ANKICONNECT_URL = "http://localhost:8765"

// Represents an Anki Notes
type Note struct {
	DeckName  string            `json:"deckName"`
	ModelName string            `json:"modelName"`
	Fields    map[string]string `json:"fields"`
	Options   any               `json:"options"`
	Audio     []Audio           `json:"audio"`
}

type Audio struct {
	Url      string   `json:"url"`
	Filename string   `json:"filename"`
	Fields   []string `json:"fields"`
}

type Result struct {
	Result []bool  `json:"result"`
	Error  *string `json:"error"`
}

// Start Anki application and wait until ankiconnect server is up.
func startAnki() {
	cmd := exec.Command("open", "-a", "anki")
	err := cmd.Start()
	if err != nil {
		panic("Error launching Anki")
	}
	err = cmd.Wait()
	if err != nil {
		panic("Error launching Anki")
	}
	time.Sleep(2 * time.Second)
}

// Creates an Anki notes from a `dict.WordDef` object to be put in `deck` using the `model`
func CreateNote(def dict.WordDef, deck string, model string) Note {
	front := def.Word
	audios := make([]Audio, 0)

	for _, e := range def.DefEntries {
		for _, p := range e.Pronunciations {
			fmt.Println(p)
			audios = append(audios, Audio{p.AudioUrl, path.Base(p.AudioUrl), []string{"Back"}})
		}
	}

	var back bytes.Buffer
	tmpl, err := template.ParseFiles("anki.tmpl")
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(&back, def)
	if err != nil {
		panic(err)
	}

	note := Note{
		deck,
		model,
		map[string]string{
			"Front": front,
			"Back":  back.String(),
		},
		map[string]any{
			"allowDuplicate": false,
			"duplicateScope": "deck",
		},
		audios,
	}
	return note
}

// Converts a slice of `dict.WordDef` objects into a slice of `Notes` objects. The notes will be assigned to `deck` and created using the specified `model`
func CreateNotes(defs []dict.WordDef, deck string, model string) []Note {
	notes := make([]Note, 0)
	for _, def := range defs {
		note := CreateNote(def, deck, model)
		notes = append(notes, note)
	}
	return notes
}

// Makes a request to the AnkiConnect API to the specified action. Returns the http.Response as the error
func ankiRequest(action, params string) (*http.Response, error) {
	fmt.Printf("Making Anki Request to %s\n", action)
	requestBody := fmt.Sprintf(`{"action": "%s", "version": 6, "params": %s}`, action, params)

	var jsonStr = []byte(requestBody)
	resp, err := http.Post(ANKICONNECT_URL, "application/json", bytes.NewBuffer(jsonStr))

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Makes request to AnkiConnect API `createDeck` action to create a new deck. If the deck already exists, nothing will happen
func AnkiCreateDeck(name string) {
	ankiRequest("createDeck", fmt.Sprintf(`{"deck": "%s"}`, name))
}

// Create a set of dummy Anki notes to check if words can be added to `deck`. This allows us to
// check which notes have already been added and reduce api calls to dictionary later
func createDummyNotes(words []string, deck string, model string) []Note {
	dummyNotes := make([]Note, 0)
	for _, w := range words {
		note := Note{
			deck,
			model,
			map[string]string{
				"Front": w,
				"Back":  "test",
			},
			map[string]any{
				"allowDuplicate": false,
				"duplicateScope": "deck",
			},
			[]Audio{},
		}

		dummyNotes = append(dummyNotes, note)
	}
	return dummyNotes
}

// Makes a request to AnkiConnect API `canAddNotes` action to check if words can be added to the specified `deck` with the specified `model`.
// Returns a slice of
func AnkiCanAddWordsToDeck(words []string, deck string, model string) []string {
	dummyNotes := createDummyNotes(words, deck, model)

	notesJson, err := json.Marshal(dummyNotes)
	if err != nil {
		panic(err)
	}

	resp, err := ankiRequest("canAddNotes", fmt.Sprintf(`{"notes": %s }`, notesJson))
	defer resp.Body.Close()
	if err != nil {
		panic(err)
	}

	var result Result
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		panic(err)
	}

	if result.Error != nil {
		panic(result.Error)
	}

	toAdd := make([]string, 0)
	for i := range words {
		if result.Result[i] {
			toAdd = append(toAdd, words[i])
		}
	}

	return toAdd
}

// Make call to Ankiconnect `addNotes` action
func AddNotes(notes []Note) error {
	notesJson, err := json.Marshal(notes)
	if err != nil {
		panic(err)
	}
	_, err = ankiRequest("addNotes", fmt.Sprintf(`{"notes": %s }`, notesJson))
	if err != nil {
		return err
	}

	fmt.Println("Finished Adding Notes")
	return nil
}
