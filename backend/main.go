package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"strconv"
	"strings"
	"os"

	"github.com/golang-jwt/jwt/v5" // go get github.com/golang-jwt/jwt/v5
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt" // go get golang.org/x/crypto/bcrypt
	"github.com/joho/godotenv"
)

const (

	createUserTable = `
		CREATE TABLE IF NOT EXISTS users(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			pw_hash TEXT NOT NULL
		)
	`
	createThreadTable = `
		CREATE TABLE IF NOT EXISTS threads(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			created_at TEXT NOT NULL,
			owner_id TEXT NOT NULL
		)
	`
	createCommentTable = `
		CREATE TABLE IF NOT EXISTS comments(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			thread_id INTEGER NOT NULL,
			message TEXT NOT NULL,
			created_at TEXT NOT NULL
		)
	`

	addUser          = "INSERT INTO users (name, pw_hash) VALUES (?, ?)"
	addComment       = "INSERT INTO comments (user_id, thread_id, message, created_at) VALUES (?, ?, ?, ?)"
	addThread		 = "INSERT INTO threads (name, created_at, owner_id) VALUES (?, ?, ?)"
	getCommentsQuery = `
		SELECT comments.id, comments.message, comments.created_at, users.name
		FROM comments
		JOIN users ON comments.user_id = users.id
		WHERE comments.thread_id = ?
		ORDER BY comments.created_at DESC
	`
	//スレッド数上限
	maxThread 	= 500
	//コメント数上限
	maxComments = 1000
	//ページネーション表示件数
	pagination 	= 5
)

var jwtKey []byte    // Replace with a secure key
const jwtExpiryTime = time.Hour * 24 // Token valid for 24 hours

type User struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	PwHash string `json:"pw_hash"`
}

type Comment struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	Name      string `json:"name"`
	ThreadID  int    `json:"thread_id"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

type Thread struct {
    ID        int    `json:"id"`
    Name      string `json:"name"`
    CreatedAt string `json:"created_at"`
    OwnerID   string `json:"owner_id"`
}

type CommentResponse struct {
	Comments		[]Comment	`json:"comments"`
	IsLimitReached  bool      	`json:"is_limit_reached"`
	MaxComments		int			`json:"max_comments"`
	CommentCount	int			`json:"comment_count"`
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
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	defer db.Close()

	//.envファイルの読み込み
	err = godotenv.Load("../.env")
	if err != nil {
		fmt.Printf(".env読み込み失敗: %v", err)
	}
	//jwtKeyを環境変数から読み取り
	jwtKeyStr, ok := os.LookupEnv("JWT_SECRET_KEY")
	if !ok {
		fmt.Println("JWT_SECRET_KEY is not set")
	}
	//環境変数を[]byte型に変換
	jwtKey = []byte(jwtKeyStr)

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

	http.HandleFunc("/api/threads", HandleCORS(func(w http.ResponseWriter, r *http.Request){
		switch r.Method {
		case http.MethodPost:
			createThread(w, r, db)
		case http.MethodGet:
			getThreads(w, r, db)
		}
	}))

	http.HandleFunc("/api/threads/", HandleCORS(func(w http.ResponseWriter, r *http.Request){
		switch r.Method {
		case http.MethodGet:
			getThreadByID(w, r, db)
		case http.MethodDelete:
			deleteThreadByID(w, r, db)
		}
	}))

	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func createUser(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var user User
	if err := decodeBody(r, &user); err != nil {
		responseJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	row := db.QueryRow("SELECT * FROM users WHERE name = ?", user.Name)
	var dbUser User
	err := row.Scan(&dbUser.ID, &dbUser.Name, &dbUser.PwHash)
	if err == nil {
		responseJSON(w, http.StatusConflict, "User already exists")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PwHash), bcrypt.DefaultCost)
	if err != nil {
		responseJSON(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	_, err = db.Exec(addUser, user.Name, string(hashedPassword))
	if err != nil {
		responseJSON(w, http.StatusInternalServerError, "Failed to add user")
		return
	}
}

func login(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var user User
	if err := decodeBody(r, &user); err != nil {
		responseJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	row := db.QueryRow("SELECT * FROM users WHERE name = ?", user.Name)
	var dbUser User
	err := row.Scan(&dbUser.ID, &dbUser.Name, &dbUser.PwHash)
	if err != nil {
		responseJSON(w, http.StatusUnauthorized, "Invalid user")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.PwHash), []byte(user.PwHash))
	if err != nil {
		responseJSON(w, http.StatusUnauthorized, "Invalid password")
		return
	}

	token, err := generateJWT(dbUser)
	if err != nil {
		responseJSON(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	responseJSON(w, http.StatusOK, map[string]string{"token": token})
}

func generateJWT(user User) (string, error) {
	claims := &jwt.MapClaims{
		"user_id": user.ID,
		"name":    user.Name,
		"exp":     time.Now().Add(jwtExpiryTime).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func validateJWT(r *http.Request) (*jwt.MapClaims, error) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		return nil, fmt.Errorf("missing token")
	}

	// "Bearer " プレフィックスを取り除く
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	claims := &jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func createComment(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	claims, err := validateJWT(r)
	if err != nil {
		responseJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	userID := int((*claims)["user_id"].(float64))
	var comment Comment
	if err := decodeBody(r, &comment); err != nil {
		responseJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	//現在のコメント数の取得
	var commentCount int 
	err = db.QueryRow("SELECT COUNT(*) FROM comments WHERE thread_id = ?",comment.ThreadID).Scan(&commentCount)
	if err != nil{
		if err == sql.ErrNoRows {
			// スレッドにコメントが存在しない場合
			commentCount = 0
		} else {
			// その他のエラー
			responseJSON(w, http.StatusInternalServerError, "Failed to fetch comment count")
			return
		}
	}
	//コメント数が上限に達していないかの確認
	if commentCount >= maxComments {
		//達していたら403を返す（コメント作成を許可しない）
		responseJSON(w, http.StatusForbidden, "Comment limit reached")
		return
	}

	now := time.Now()
	_, err = db.Exec(addComment, userID, comment.ThreadID, comment.Message, now)
	if err != nil {
		responseJSON(w, http.StatusInternalServerError, "Failed to add comment")
		return
	}

	responseJSON(w, http.StatusCreated, "Comment created successfully")
}

func getComments(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	thread_id := r.URL.Query().Get(("threadID"))
	rows, err := db.Query(getCommentsQuery, thread_id)
	if err != nil {
		responseJSON(w, http.StatusInternalServerError, "Failed to retrieve comments")
		return
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		err := rows.Scan(&comment.ID, &comment.Message, &comment.CreatedAt, &comment.Name)
		if err != nil {
			responseJSON(w, http.StatusInternalServerError, "Failed to parse comment data")
			return
		}
		comments = append(comments, comment)
	}

	//現在のコメント数の取得
	var commentCount int 
	err = db.QueryRow("SELECT COUNT(*) FROM comments WHERE thread_id = ?",thread_id).Scan(&commentCount)
	if err != nil{
		if err == sql.ErrNoRows {
			// スレッドにコメントが存在しない場合
			commentCount = 0
		} else {
			// その他のエラー
			responseJSON(w, http.StatusInternalServerError, "Failed to fetch comment count")
			return
		}
	}
	//上限に達しているか否かを保持
	isLimitReached := commentCount >= maxComments
	//コメント配列と上限に達しているかどうかをまとめる
	response := CommentResponse{
		Comments: comments,
		IsLimitReached: isLimitReached,
		MaxComments: maxComments,
		CommentCount: commentCount,
	}
	responseJSON(w, http.StatusOK, response)
}

func createThread(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	//現在のスレッド数の取得
	var threadCount int 
	err := db.QueryRow("SELECT COUNT(*) FROM threads").Scan(&threadCount)
	if err != nil{
		if err == sql.ErrNoRows {
			// スレッドが存在しない場合
			threadCount = 0
		} else {
			// その他のエラー
			responseJSON(w, http.StatusInternalServerError, "Failed to fetch comment count")
			return
		}
	}
	//スレッド数が上限に達していないかの確認
	if threadCount >= maxThread {
		//達していたら403を返す（スレッド作成を許可しない）
		responseJSON(w, http.StatusForbidden, "Comment limit reached")
		return
	}

	claims, err := validateJWT(r)
	if err != nil {
		responseJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	userID := int((*claims)["user_id"].(float64))
	//DBに込める値を受け取るための変数宣言
	var thread Thread
	//デコードする
	if err := decodeBody(r, &thread); err != nil {
		responseJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	now := time.Now()
	_, err = db.Exec(addThread, thread.Name, now, userID)
	if err != nil {
		responseJSON(w, http.StatusInternalServerError, "Faled to add thread")
	}

	responseJSON(w, http.StatusCreated, "Thread created successfully")
}

type ThreadInfo struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	OwnerID   string `json:"owner_id"`
	CommentCount int `json:"comment_count"`
}

type ThreadResponse struct {
	Threads			[]ThreadInfo	`json:"threads"`
	IsLimitReached	bool		`json:"is_limit_reached"`
	MaxThreads		int			`json:"max_threads"`
	ThreadCount		int 		`json:"thread_count"`
	PageCount		int			`json:"page_count"`
}

func getThreads(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	//現在のページ数の取得
	queryParams := r.URL.Query();
	pageStr := queryParams.Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		responseJSON(w, http.StatusInternalServerError, "Invalid page")
	}

	//現在のスレッド数の取得
	var threadCount int 
	err = db.QueryRow("SELECT COUNT(*) FROM threads").Scan(&threadCount)
	if err != nil{
		if err == sql.ErrNoRows {
			// スレッドが存在しない場合
			threadCount = 0
		} else {
			// その他のエラー
			responseJSON(w, http.StatusInternalServerError, "Failed to fetch comment count")
			return
		}
	}
	
	getQuery := `
	SELECT 
		threads.id, 
		threads.name, 
		threads.created_at, 
		threads.owner_id, 
		COUNT(comments.id) AS comment_count
	FROM threads
	LEFT JOIN comments 
	ON threads.id = comments.thread_id
	GROUP BY threads.id
	LIMIT ? OFFSET ?`

	var threads []ThreadInfo
	rows, err := db.Query(getQuery, pagination, (page-1)*pagination)
	if err != nil {
		responseJSON(w, http.StatusInternalServerError, "Failed to get threads")	
		return
	}
	defer rows.Close()

	for rows.Next() {
		var thread ThreadInfo
		err := rows.Scan(&thread.ID, &thread.Name, &thread.CreatedAt, &thread.OwnerID, &thread.CommentCount)
		if err != nil {
			responseJSON(w, http.StatusInternalServerError, "Failed to parse threads data")
			return
		}
		threads = append(threads, thread)
	}

	//上限に達しているか否かを保持
	isLimitReached := threadCount >= maxThread
	//ページ数を保持(繰り上げ)
	pageCount := (threadCount+pagination-1)/pagination
	//コメント配列と上限に達しているかどうかをまとめる
	response := ThreadResponse{
		Threads: threads,
		IsLimitReached: isLimitReached,
		MaxThreads: maxThread,
		ThreadCount: threadCount,
		PageCount: pageCount,
	}
	responseJSON(w, http.StatusOK, response)
}

func getThreadByID(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var thread Thread
	idStr := strings.TrimPrefix(r.URL.Path, "/api/threads/")
    threadID, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid thread ID", http.StatusBadRequest)
        return
    }
	
	row := db.QueryRow("SELECT * FROM threads WHERE id = ?", threadID)
	err = row.Scan(&thread.ID, &thread.Name, &thread.CreatedAt, &thread.OwnerID)
	if thread.ID == 0 {
		responseJSON(w, 404, thread)
		return
	}
	if err != nil {
		
		responseJSON(w, http.StatusInternalServerError, "Failed to parse thread data")
		return
	}

	responseJSON(w, http.StatusOK, thread)
}

func deleteThreadByID(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/threads/")
    threadID, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid thread ID", http.StatusBadRequest)
        return
    }
	//コメントとスレッドの一方のみが削除されるのを防ぐトランザクション
	tx, err := db.Begin()
	if err != nil {
		responseJSON(w, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer func() {
        if err != nil {
            tx.Rollback() // エラー時にロールバックする
        }
    }()

	_, err = tx.Exec("DELETE FROM comments WHERE thread_id=?",threadID)
	if err != nil{
		tx.Rollback()
		responseJSON(w, http.StatusInternalServerError, "Failed to delete comments")
		return
	}

	_,err = tx.Exec("DELETE FROM threads WHERE id=?", threadID)
	if err != nil {
		tx.Rollback()
		responseJSON(w, http.StatusInternalServerError, "Failed to delete thread")
		return
	}

	err = tx.Commit()
	if err != nil {
		responseJSON(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	responseJSON(w, http.StatusOK, "thread and comments deleted successfully")
}

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

func decodeBody(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}
	return nil
}

func responseJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		panic(err)
	}
}
