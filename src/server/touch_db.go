// package main

// import (
// 	"database/sql"
// 	"fmt"
// 	// "log"

// 	_ "github.com/go-sql-driver/mysql"
// )

// // func main() {
// // 	db, err := connectToDatabase("user:user@tcp(mariadb:3306)/server?parseTime=true")
// // 	if err != nil {
// // 		log.Fatal("Failed to connect to database:", err)
// // 	}
// // 	defer db.Close()

// // 	fmt.Println("Successfully connected to the database.")

// // 	// 挿入関数（コメントアウト）
// // 	if err := insertUser(db, "some-uid", "some-login"); err != nil {
// // 		log.Fatal("Failed to insert user:", err)
// // 	}

// // 	// 参照関数
// // 	// if err := getUsers(db); err != nil {
// // 	// 	log.Fatal("Failed to get users:", err)
// // 	// }
// // }

// func connectToDatabase(dsn string) (*sql.DB, error) {
// 	db, err := sql.Open("mysql", dsn)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if err := db.Ping(); err != nil {
// 		return nil, err
// 	}
// 	return db, nil
// }

// // insertUserは新しいユーザーをusersテーブルに挿入する関数です（コメントアウト）。
// func insertUser(db *sql.DB, uid, login string) error {
// 	_, err := db.Exec("INSERT INTO users (uid, login) VALUES (?, ?)", uid, login)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println("Inserted a new user into the users table.")
// 	return nil
// }

// // getUsersはusersテーブルから全てのユーザーを取得し、表示する関数です。
// func getUsers(db *sql.DB) error {
// 	rows, err := db.Query("SELECT id, uid, login FROM users")
// 	if err != nil {
// 		return err
// 	}
// 	defer rows.Close()

// 	fmt.Println("Users:")
// 	for rows.Next() {
// 		var id int
// 		var uid, login string
// 		if err := rows.Scan(&id, &uid, &login); err != nil {
// 			return err
// 		}
// 		fmt.Printf("ID: %d, UID: %s, Login: %s\n", id, uid, login)
// 	}
// 	if err := rows.Err(); err != nil {
// 		return err
// 	}
// 	return nil
// }
