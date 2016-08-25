package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type GuidanceFile struct {
	ID           int64
	Mimetype     sql.NullString
	Key          sql.NullString
	Name         sql.NullString
	GuidanceID   int64
	CreateUserID sql.NullInt64
	UpdateUserID sql.NullInt64
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

func main() {
	user := ""
	pw := ""
	dbName := ""
	host := ""
	port := "3306"
	str := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true", user, pw, host, port, dbName)

	db, err := sql.Open("mysql", str)
	if err != nil {
		fmt.Println("open error :", err)
		os.Exit(1)
	}
	defer db.Close()
	key := ""
	originFile := GuidanceFile{}
	sql := fmt.Sprintf("select id, mimetype, guidance_id, create_user_id from guidance_files  where `key` = \"%s\"", key)
	err = db.QueryRow(sql).Scan(&originFile.ID, &originFile.Mimetype, &originFile.GuidanceID, &originFile.CreateUserID)
	if err != nil {
		fmt.Println("select error ", err)
		os.Exit(1)
	}

	fmt.Println("originFile: ", originFile.ID, originFile.Mimetype, originFile.GuidanceID)
	thumbKey := fmt.Sprintf("%s_thumbnail_lambda", key)
	_, err = db.Exec("insert into guidance_files  (mimetype, `key`, name, guidance_id, create_user_id, update_user_id, created_at, updated_at) values(?, ?, ?, ?, ?, ?, NOW(), NOW())", originFile.Mimetype, thumbKey, "サムネイル", originFile.GuidanceID, originFile.CreateUserID, originFile.CreateUserID)
	if err != nil {
		fmt.Println("insert error :", err)
		os.Exit(1)
	}
}
