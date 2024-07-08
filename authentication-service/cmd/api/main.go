package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/justamanpop/microservices_authentication/data"
)

const webPort = "80"

type Config struct {
	Repo   data.Repository
	Client *http.Client
}

func main() {
	log.Println("Starting authentication service")

	conn := connectToDb()
	if conn == nil {
		log.Panic("Cannot connect to Postgres")
	}

	app := Config{
		Client: &http.Client{},
	}
	app.setupRepository(conn)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Panic("Could not start auth server", err)
	}
}

func openDb(connString string) (*sql.DB, error) {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func connectToDb() *sql.DB {
	connString := os.Getenv("DB_CONNECTION_STRING")
	counts := 0
	for {
		connection, err := openDb(connString)
		if err != nil {
			log.Println("Postgres not yet up...")
		} else {
			log.Println("Connected to Postgres!")
			return connection
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Retrying in 2 seconds")
		time.Sleep(2 * time.Second)
	}
}

func (cfg *Config) setupRepository(conn *sql.DB) {
	db := data.NewPostgresRepository(conn)
	cfg.Repo = db
}
