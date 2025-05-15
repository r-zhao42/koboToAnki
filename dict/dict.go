package dict

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"unicode"
)

type WordDef struct {
	Word       string
	DefEntries []DefEntry
}

type DefEntry struct {
	Defs            []string
	FunctionalLabel string
	Stems           []string
	Pronunciations  []Pronunciation
}

type Pronunciation struct {
	Pronunciation string
	AudioUrl      string
}

type DictAPI interface {
	GetDefinition(word string) WordDef

	GetDefinitions(words []string) []string
}

func NewMerriam() (Merriam, error) {
	api_key := os.Getenv("MERRIAM_API_KEY")
	if api_key == "" {
		return Merriam{}, errors.New("Merriam Webster API key was not set")

	}
	return Merriam{api_key}, nil
}

func PrintDefinitions(defs []WordDef) {
	for _, w := range defs {
		fmt.Println(w.Word)
		for _, entry := range w.DefEntries {
			fmt.Println(entry.FunctionalLabel, ",", entry.Pronunciations)
			for _, dt := range entry.Defs {
				fmt.Println("   ", dt)
			}
			fmt.Println(entry.Stems)
		}
		fmt.Println()
	}
}

// Merriam Webster API implementation
type Merriam struct {
	api_key string
}

func (m Merriam) getUrl(word string) string {
	return fmt.Sprintf("https://dictionaryapi.com/api/v3/references/collegiate/json/%s?key=%s", word, m.api_key)
}

type MerriamEntry struct {
	Fl       string   `json:"fl"`
	Shortdef []string `json:"shortdef"`
	Def      []struct {
		Sseq [][][]json.RawMessage `json:"sseq"`
	} `json:"def"`
	Hwi struct {
		Prs []struct {
			Mw    string `json:"mw"`
			Sound struct {
				AudioUrl string `json:"audio"`
			} `json:"sound"`
		} `json:"prs"`
	} `json:"hwi"`
	Meta struct {
		Stems []string `json:"stems"`
		Word  string   `json:"id"`
	} `json:"meta"`
}

func (m Merriam) GetDefinition(word string) (WordDef, error) {
	url := m.getUrl(word)
	resp, err := http.Get(url)

	if err != nil {
		panic("error")
	}

	body, err := io.ReadAll(resp.Body)

	var entries []MerriamEntry

	err = json.Unmarshal(body, &entries)
	if err != nil {
		fmt.Println("Error getting definition for ", word, ", skipping this word:", err)
		return WordDef{}, err
	}

	res := WordDef{Word: word}
	for _, e := range entries {
		defEntry := DefEntry{
			Defs:            e.Shortdef,
			FunctionalLabel: e.Fl,
			Stems:           e.Meta.Stems,
		}

		for _, pr := range e.Hwi.Prs {
			var audioUrl string

			if pr.Sound.AudioUrl != "" {
				audioUrl, err = m.getAudio(pr.Sound.AudioUrl)
				if err != nil {
					continue
				}
			}
			defEntry.Pronunciations = append(defEntry.Pronunciations, Pronunciation{pr.Mw, audioUrl})
		}
		res.DefEntries = append(res.DefEntries, defEntry)
	}
	return res, nil
}

func (m Merriam) GetDefinitions(words []string) []WordDef {
	res := make([]WordDef, 0)
	for _, w := range words {
		def, err := m.GetDefinition(w)
		if err != nil {
			continue
		}
		res = append(res, def)
	}

	return res
}

func (m Merriam) getAudio(fileRef string) (string, error) {
	url := "https://media.merriam-webster.com/audio/prons/en/us/mp3/%s/%s.mp3"

	var subdirectory string
	if strings.HasPrefix(fileRef, "bix") {
		subdirectory = "bix"
	} else if strings.HasPrefix(fileRef, "gg") {
		subdirectory = "gg"
	} else if unicode.IsDigit(rune(fileRef[0])) || unicode.IsPunct(rune(fileRef[0])) {
		subdirectory = "number"
	} else if unicode.IsLetter(rune(fileRef[0])) {
		subdirectory = string(fileRef[0])
	} else {
		return "", errors.New("Invalid audio file name")
	}

	return fmt.Sprintf(url, subdirectory, fileRef), nil
}
