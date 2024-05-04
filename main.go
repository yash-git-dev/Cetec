package main

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	r := gin.Default()

	r.GET("/person/:person_id/info", PersonGET)
	r.POST("/person/create", PersonPOST)

	if err := r.Run(":8000"); err != nil {
		panic(fmt.Sprintf("failed to start the server: %v", err))
	}
}

func ConnectDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/cetec")
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
