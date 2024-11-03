package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt" // go get golang.org/x/crypto/bcrypt
	_ "github.com/mattn/go-sqlite3"
)

const (
	// ユーザーテーブル作成SQL
	createUserTable = `
		CREATE TABLE IF NOT EXISTS users(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			pw_hash TEXT NOT NULL
		)
	`
	// スレッドテーブル作成SQL
	createThreadTable = `
		CREATE TABLE IF NOT EXISTS threads(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			created_at TEXT NOT NULL,
			owner_id TEXT NOT NULL
		)
	`
	// コメントテーブル作成SQL
	createCommentTable = `
		CREATE TABLE IF NOT EXISTS comments(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			thread_id INTEGER NOT NULL,
			message TEXT NOT NULL,
			created_at TEXT NOT NULL
		)
	`
	// ユーザー追加SQL
	addUser = "INSERT INTO users (name, pw_hash) VALUES (?, ?)"
	// コメント追加SQL
	addComment = "INSERT INTO comments (user_id, thread_id, message, created_at) VALUES (?, ?, ?, ?)"
	// コメント取得SQL
	getCommentsQuery = "SELECT * FROM comments WHERE thread_id = ? ORDER BY created_at DESC"
)

// ユーザー情報を格納する構造体
type User struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	PwHash string `json:"pw_hash"`
}

// コメント情報を格納する構造体
type Comment struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	ThreadID  int    `json:"thread_id"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

func init() {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	defer db.Close()
}

func main() {
	// データベース接続
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	defer db.Close()

	// テーブル作成
	_, err = db.Exec(createUserTable)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(createThreadTable)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(createCommentTable)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/api/users", HandleCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			createUser(w, r, db)
		case http.MethodGet:
			getUsers(w, r, db)
		}
	}))

	http.HandleFunc("/api/login", HandleCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			login(w, r, db)
		}
	}))

	http.HandleFunc("/api/comments", HandleCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			createComment(w, r, db)
		case http.MethodGet:
			getComments(w, r, db)
		}
	}))

	// サーバーの起動、ポート番号は8080
	fmt.Println("http://localhost:8080 でサーバーを起動します")
	http.ListenAndServe(":8080", nil)
}

// ユーザー追加ハンドラ
func createUser(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var user User
	if err := decodeBody(r, &user); err != nil {
		responseJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// ユーザー名の重複チェック
	row := db.QueryRow("SELECT * FROM users WHERE name = ?", user.Name)
	var dbUser User
	err := row.Scan(&dbUser.ID, &dbUser.Name, &dbUser.PwHash)
	if err == nil {
		responseJSON(w, http.StatusConflict, "User already exists")
		return
	}

	// パスワードのハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PwHash), bcrypt.DefaultCost) 
	if err != nil {
		responseJSON(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// ユーザー追加
	_, err = db.Exec(addUser, user.Name, string(hashedPassword))
	if err != nil {
		responseJSON(w, http.StatusInternalServerError, "Failed to add user")
		return
	}
}

// ログインハンドラ
func login(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var user User
	if err := decodeBody(r, &user); err != nil {
		responseJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// ユーザー名からユーザー情報を取得
	row := db.QueryRow("SELECT * FROM users WHERE name = ?", user.Name)
	var dbUser User
	err := row.Scan(&dbUser.ID, &dbUser.Name, &dbUser.PwHash)
	if err != nil {
		responseJSON(w, http.StatusUnauthorized, "Invalid user")
		return
	}

	// パスワードの照合
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.PwHash), []byte(user.PwHash))
	if err != nil {
		responseJSON(w, http.StatusUnauthorized, "Invalid password")
		return
	}

	responseJSON(w, http.StatusOK, "Login successful")
}

// ユーザー取得ハンドラ
func getUsers(w http.ResponseWriter, _ *http.Request, db *sql.DB) {
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		panic(err)
	}

	var users []User

	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.PwHash)
		if err != nil {
			panic(err)
		}
		users = append(users, user)
	}

	responseJSON(w, http.StatusOK, users)
}

// コメント追加ハンドラ
func createComment(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var comment Comment
	if err := decodeBody(r, &comment); err != nil {
		responseJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	now := time.Now()
	// ユーザー１，スレッド１の想定
	_, err := db.Exec(addComment, 1, 1, comment.Message, now)
	if err != nil {
		responseJSON(w, http.StatusInternalServerError, "Failed to add comment")
		return
	}

	responseJSON(w, http.StatusCreated, "Comment created successfully")
}

// コメント取得ハンドラ
func getComments(w http.ResponseWriter, _ *http.Request, db *sql.DB) {
	// スレッド１の投稿を取ってくる
	rows, err := db.Query(getCommentsQuery, 1)
	if err != nil {
		panic(err)
	}

	var comments []Comment

	for rows.Next() {
		var comment Comment
		err := rows.Scan(&comment.ID, &comment.UserID, &comment.ThreadID, &comment.Message, &comment.CreatedAt)
		if err != nil {
			panic(err)
		}
		comments = append(comments, comment)
	}

	responseJSON(w, http.StatusOK, comments)
}

// CORS設定ミドルウェア
func HandleCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		h(w, r)
	}
}

// JSONをデコードする関数
func decodeBody(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}
	return nil
}

// JSONにエンコードして返す
func responseJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		panic(err)
	}
}
