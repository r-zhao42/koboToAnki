package kobo

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Search for and returns the path to the `KoboReader.sqlite`. Only works if kobo is connected.
// Panics if kobo volume can't be found
func findKoboPath() string {
	var koboPath string
	err := filepath.WalkDir("/Volumes/", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		fmt.Println(path)
		if d.Name() == "KOBOeReader" {
			koboPath = path
			return filepath.SkipAll
		}

		return nil
	})

	if err != nil || koboPath == "" {
		panic("Error finding kobo path")
	}

	return filepath.Join(koboPath, ".kobo", "KoboReader.sqlite")
}

// Extracts kobo saved words from the sqlite database in `db_path`
// `db_path` must have a valid kobo sqlite files with a `WordList` column, otherwise will panic
func getWordsFromPath(db_path string) ([]string, error) {
	fmt.Println("Opening Database")
	if _, err := os.Stat(db_path); errors.Is(err, os.ErrNotExist) {
		return []string{}, err

	}
	db, err := sql.Open("sqlite", db_path)
	if err != nil {
		panic("error")
	}

	fmt.Println("Querying Database")
	rows, err := db.Query("SELECT Text FROM WordList")
	if err != nil {
		panic("error")
	}

	fmt.Println("Extracting Rows")
	texts := make([]string, 0)
	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			panic(err)
		}
		texts = append(texts, text)
	}
	db.Close()
	return texts, nil
}

// Returns a slice of strings representing the words stored in the `.kobo/KoboReader.sqlite` database on the kobo.
// If Kobo is not connected or an error occurs, program stops
func GetWords() []string {
	db_path := findKoboPath()
	words, err := getWordsFromPath(db_path)
	if err != nil {
		log.Fatalf("Kobo not connected or path to kobo is incorrect")
	}
	return words
}
