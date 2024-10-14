package main

import (
	"database/sql"
	"log"
	"fmt"
	"net/http"

	_"github.com/mattn/go-sqlite3"
)

const (
	/*ユーザーテーブル作成SQL*/
	createUserTable = `
		CREATE TABLE IF NOT EXISTS users(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			pw_hash TEXT NOT NULL
		)
	`
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
	/*データベース接続*/
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
		panic(err);
	}
	defer db.Close()

	/*テーブル作成*/
	_, err = db.Exec(createUserTable)
	if err != nil {
		panic(err)
	}

	// サーバーの起動、ポート番号は8080
	fmt.Println("http://localhost:8080 でサーバーを起動します")
	http.ListenAndServe(":8080", nil)
}