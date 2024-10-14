package main

import (
	"database/sql"
	"log"

	_"github.com/mattn/go-sqlite3"
)

func init(){
	db, err := sql.Open("sqlite3","./database.db")
	if err != nil{
		log.Fatal(err)
		panic(err)
	}
	defer db.Close()

}

func main(){
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
		panic(err);
	}
	defer db.Close()
}