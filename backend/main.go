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

	"github.com/golang-jwt/jwt/v5" // go get github.com/golang-jwt/jwt/v5
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt" // go get golang.org/x/crypto/bcrypt
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
)

var jwtKey = []byte("secret_key")    // Replace with a secure key
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
	now := time.Now()
	_, err = db.Exec(addComment, userID, comment.ThreadID, comment.Message, now)
	if err != nil {
		responseJSON(w, http.StatusInternalServerError, "Failed to add comment")
		return
	}

	responseJSON(w, http.StatusCreated, "Comment created successfully")
}

func createThread(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

	responseJSON(w, http.StatusOK, comments)
}

func getThreads(w http.ResponseWriter, _ *http.Request, db *sql.DB) {
	var threads []Thread
	rows, err := db.Query("SELECT * FROM threads")
	if err != nil {
		responseJSON(w, http.StatusInternalServerError, "Failed to get threads")	
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		err := rows.Scan(&thread.ID, &thread.Name, &thread.CreatedAt, &thread.OwnerID)
		if err != nil {
			responseJSON(w, http.StatusInternalServerError, "Failed to parse threads data")
			return
		}
		threads = append(threads, thread)
	}

	responseJSON(w, http.StatusOK, threads)
}

func getThreadByID(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var thread Thread
	idStr := strings.TrimPrefix(r.URL.Path, "/api/threads/")
    threadID, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid thread ID", http.StatusBadRequest)
        return
    }
	fmt.Println("getThreadByID:",threadID)
	fmt.Println("getThreadByID:", threadID)
	row := db.QueryRow("SELECT * FROM threads WHERE id = ?", threadID)
	
	err = row.Scan(&thread.ID, &thread.Name, &thread.CreatedAt, &thread.OwnerID)
	if err != nil {
		responseJSON(w, http.StatusInternalServerError, "Failed to parse thread data")
		return
	}

	responseJSON(w, http.StatusOK, thread)
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
