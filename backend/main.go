package main

import (
	"database/sql"
	"log"
	"fmt"
	"net/http"

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

	// サーバーの起動、ポート番号は8080
	fmt.Println("http://localhost:8080 でサーバーを起動します")
	http.ListenAndServe(":8080", nil)
}