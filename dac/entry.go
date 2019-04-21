package dac

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

type table interface {
	CreateStatement() string
	GetAll() *sql.Rows
	ListAll()
}

type URL struct {
	Name, Schema string
	Db           *sql.DB
}

func (t URL) CreateStatement() string {
	return "CREATE TABLE IF NOT EXISTS " + t.Name + " (" + t.Schema + ");"
}

func (t URL) RemoveShortcut(shortcut string) {
	statement := "DELETE FROM mappings WHERE redirection = '" + shortcut + "';"
	_, err := t.Db.Exec(statement)
	check(err, t.Db)
}

func (t URL) GetAll() *sql.Rows {
	statement := "SELECT * FROM " + t.Name + " ORDER BY numberOfUses DESC;"
	rows, err := t.Db.Query(statement)
	check(err, t.Db)
	return rows
}

func (t URL) ListAll() {
	rows := t.GetAll()
	var id, numberOfUses int
	var url, redirection string
	for rows.Next() {
		rows.Scan(&id, &url, &redirection, &numberOfUses)
		postfix := "times"
		if numberOfUses <= 1 {
			postfix = "time"
		}
		fmt.Println(strconv.Itoa(id) + ": " + redirection + " redirect to  " + url + " has been used " + strconv.Itoa(numberOfUses) + " " + postfix + ".")
	}
}

func (t URL) Insert(url string, shortcut string) {
	statement := "INSERT INTO mappings (url, redirection, numberOfUses) VALUES('" + url + "','" + shortcut + "', 0);"
	_, err := t.Db.Exec(statement)
	check(err, t.Db)
}

func (t URL) GetURL(shortcut string) string {
	statement := "SELECT url FROM mappings WHERE redirection = '" + shortcut + "';"
	row := t.Db.QueryRow(statement, shortcut)
	var url string
	switch err := row.Scan(&url); err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
	case nil:
		return url
	default:
		check(err, t.Db)
	}

	return ""
}

func (t URL) LogRedirection(shortcut string) {
	_, err := t.Db.Exec("UPDATE mappings SET numberOfUses = numberOfUses + 1 WHERE redirection='" + shortcut + "';")
	check(err, t.Db)
}

func ConnectToDB(connection string) *sql.DB {
	db, _ := sql.Open("sqlite3", connection)
	return db
}

func Create(t table, db *sql.DB) {
	_, err := db.Exec(t.CreateStatement())
	check(err, db)
}

func check(err error, db *sql.DB) {
	if err != nil {
		log.Fatal(err)
	}
}
