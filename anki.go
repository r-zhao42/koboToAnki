package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"koboToAnki/m/dict"
	"net/http"
	"os/exec"
	"path"
	"text/template"
	"time"
)

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

func startAnki(path string) {
	cmd := exec.Command("open", path)
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

func CreateNote(def dict.WordDef, deck string, model string) Note {
	front := def.Word
	audios := make([]Audio, 0)

	for _, e := range def.DefEntries {
		for _, p := range e.Pronunciations {
			fmt.Println(p)
			audios = append(audios, Audio{p.AudioUrl, path.Base(p.AudioUrl), []string{"Back"}})
		}
	}
	fmt.Println(audios)

	fmt.Println(def.DefEntries[0].Defs)
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

func CreateNotes(defs []dict.WordDef, deck string, model string) []Note {
	notes := make([]Note, 0)
	for _, def := range defs {
		note := CreateNote(def, deck, model)
		notes = append(notes, note)
	}
	return notes
}

func ankiRequest(action, params string) (*http.Response, error) {
	fmt.Printf("Making Anki Request to %s\n", action)
	url := "http://127.0.0.1:8765"
	requestBody := fmt.Sprintf(`{"action": "%s", "version": 6, "params": %s}`, action, params)

	var jsonStr = []byte(requestBody)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func AnkiCreateDeck(name string) {
	ankiRequest("createDeck", fmt.Sprintf(`{"deck": "%s"}`, name))
}

func AnkiCanAddWordsToDeck(words []string, deck string, model string) []bool {

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
	return result.Result
}

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
