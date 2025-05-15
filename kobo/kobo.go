package kobo

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

func GetWords(uri string) ([]string, error) {
	fmt.Println("Opening Database")
	if _, err := os.Stat(uri); errors.Is(err, os.ErrNotExist) {
		return []string{}, err

	}
	db, err := sql.Open("sqlite", uri)
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
