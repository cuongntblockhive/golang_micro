package main

import (
	"authentication/data"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const webPort = "80"

var counts int16

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Println("Start authentication services")

	app := Config{}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}
	// start server
	err := srv.ListenAndServe()

	if err != nil {
		log.Panicf("Err run server")
	}

}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("Pg not ready")
			counts++
		} else {
			log.Println("Connected to PG")
			return connection
		}
		if counts > 10 {
			log.Println(err)
			return nil
		}
		log.Println("Backoff in 2 seconds")
		time.Sleep(2 * time.Second)
		continue
	}
}
