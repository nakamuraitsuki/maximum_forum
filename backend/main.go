package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

const (
	//ユーザーテーブル作成SQL
	createUserTable = `
		CREATE TABLE IF NOT EXISTS users(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			pw_hash TEXT NOT NULL
		)
	`
	//スレッドテーブル作成SQL
	createThreadTable = `
		CREATE TABLE IF NOT EXISTS threads(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			created_at TEXT NOT NULL,
			owner_id TEXT NOT NULL
		)
	`
	//コメントテーブル作成SQL
	createCommentTable = `
		CREATE TABLE IF NOT EXISTS comments(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			thread_id INTEGER NOT NULL,
			message TEXT NOT NULL,
			created_at TEXT NOT NULL
		)
	`
	//コメント追加SQL
	addComment = "INSERT INTO comments (user_id, thread_id, message, created_at) VALUES (?, ?, ?, ?)"
)

// ユーザー情報を格納する構造体
type User struct {
	ID int `json:"id"`
	Name string `json:"name"`
	PwHash string `json:"pw_hash"`
}

// コメント情報を格納する構造体
type Comment struct {
	ID int `json:"id"`
	UserID int `json:"user_id"`
	ThreadID int `json:"thread_id"`
	Message string `json:"message"`
	CreatedAt string `json:"created_at"`
}

func init(){
	db, err := sql.Open("sqlite3","./database.db")
	if err != nil{
		log.Fatal(err)
		panic(err)
	}
	defer db.Close()

}

func main(){
	// データベース接続
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
		panic(err);
	}
	defer db.Close()

	// テーブル作成
	_, err = db.Exec(createUserTable)
	if err != nil {
		panic(err)
	}

	// テーブル作成（スレッド）
	_, err = db.Exec(createThreadTable)
	if err != nil {
		panic(err)
	}

	//デーブル作成（コメント）
	_, err = db.Exec(createCommentTable)
	if err != nil {
		panic(err)
	}

	// サーバーの起動、ポート番号は8080
	fmt.Println("http://localhost:8080 でサーバーを起動します")
	http.ListenAndServe(":8080", nil)
}

/*
	CORS設定ミドルウェア
	httpハンドラーを受け取って，CORS設定をした状態で返す
	ルーティングの際に使います
*/
func HandleCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// レスポンスヘッダーの設定
		w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
        w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// リクエストヘッダーの設定
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// ハンドラーの実行
		h(w, r)
	}
}
//JSONをデコードする関数
func decodeBody(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v);
	err != nil {
		return err
	}
	return nil
}
