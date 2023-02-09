package db

import (
	"log"

	"github.com/genjidb/genji"
)

func NewDB(path string) (*genji.DB, error) {
	db, err := genji.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	err = initTables(db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func initTables(db *genji.DB) error {
	err := db.Exec(`
	CREATE TABLE IF NOT EXISTS messages (
		username TEXT,
		firstname TEXT,
		lastname TEXT,
		message TEXT                        
	)
`)
	if err != nil {
		return err
	}

	return nil
}
