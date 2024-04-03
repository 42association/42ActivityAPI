package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
)

// InitializeDatabase はデータベース接続の初期化を行います。
// この関数は外部ファイルから呼び出されることを想定しています。
func InitializeDatabase() (*sql.DB, error) {
	dsn, err := getDSN()
	if err != nil {
		return nil, err
	}
	db, err := connectToDatabase(dsn)
	if err != nil {
		return nil, err
	}
	if err := verifyConnection(db); err != nil {
		return nil, err
	}
	return db, nil
}

// getDSN はDSN（Data Source Name）を環境変数から取得します。
func getDSN() (string, error) {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		return "", fmt.Errorf("DB_DSN environment variable is not set")
	}
	return dsn, nil
}

// connectToDatabase はデータベースへの接続を試みます。
func connectToDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// verifyConnection はデータベースへの接続を確認します。
func verifyConnection(db *sql.DB) error {
	if err := db.Ping(); err != nil {
		return err
	}
	return nil
}

// InsertUser は新しいユーザーをusersテーブルに挿入します。
func InsertUser(db *sql.DB, uid, login string) error {
	_, err := db.Exec("INSERT INTO users (uid, login) VALUES (?, ?)", uid, login)
	if err != nil {
		return err
	}
	fmt.Println("Inserted a new user into the users table.")
	return nil
}

func GetUserByUid(db *sql.DB, uid string) (User, error) {
	var user User
	query := "SELECT id, uid, login FROM users WHERE uid = $1"
	err := db.QueryRow(query, uid).Scan(&user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUsers はusersテーブルから全てのユーザーを取得し、表示します。
func GetUsers(db *sql.DB) ([]User, error) {
	rows, err := db.Query("SELECT id, uid, login FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	fmt.Println("Users:")
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.UID, &u.Login); err != nil {
			return nil, err
		}
		users = append(users, u)
		fmt.Printf("ID: %d, UID: %s, Login: %s\n", u.ID, u.UID, u.Login)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

// User はusersテーブルの行を表す構造体です。
type User struct {
	ID    int
	UID   string
	Login string
}
