package databases

import (
	"database/sql"
	"eiradinner/internal/logger"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// 连接数据库
func OpenDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./db/eiradinner.db")
	if err != nil {
		return nil, err
	}
	return db, nil
}

// 初始化数据库
func InitDatabase() (*sql.DB, error) {
	createTableSQL := `CREATE TABLE IF NOT EXISTS Computers (
	session_id INTEGER PRIMARY KEY AUTOINCREMENT,
	ip_address TEXT,
	port INTEGER,
	os TEXT,
	path TEXT,
	status TEXT,
	linstener TEXT,
	user      TEXT
);
	CREATE TABLE IF NOT EXISTS Listener(
	listener_id INTEGER PRIMARY KEY AUTOINCREMENT,
	listener_name TEXT,
	listener_port INTEGER
	);`
	db, err := sql.Open("sqlite3", "./db/eiradinner.db")
	if err != nil {
		logger.LogError(err)
		return nil, err
	}
	_, err = db.Exec(createTableSQL)
	if err != nil {
		logger.LogError(err)
	}
	return db, err
}

// 增加用户
func createUser(db *sql.DB, name string, age int) {
	insertSQL := `INSERT INTO users (name, age) VALUES (?, ?)`
	_, err := db.Exec(insertSQL, name, age)
	if err != nil {
		log.Fatalf("无法插入用户: %v", err)
	}
}

// 查询用户
func readUsers(db *sql.DB) {
	rows, err := db.Query("SELECT id, name, age FROM users")
	if err != nil {
		log.Fatalf("查询失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		var age int
		if err := rows.Scan(&id, &name, &age); err != nil {
			log.Fatalf("扫描失败: %v", err)
		}
		fmt.Printf("用户ID: %d, 姓名: %s, 年龄: %d\n", id, name, age)
	}
}

// 更新用户
func updateUser(db *sql.DB, id int, name string, age int) {
	updateSQL := `UPDATE users SET name = ?, age = ? WHERE id = ?`
	_, err := db.Exec(updateSQL, name, age, id)
	if err != nil {
		log.Fatalf("无法更新用户: %v", err)
	}
}

// 删除用户
func deleteUser(db *sql.DB, id int) {
	deleteSQL := `DELETE FROM users WHERE id = ?`
	_, err := db.Exec(deleteSQL, id)
	if err != nil {
		log.Fatalf("无法删除用户: %v", err)
	}
}
