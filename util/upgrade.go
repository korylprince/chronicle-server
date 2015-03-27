package main

import (
	"database/sql"
	"flag"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/korylprince/chronicle-server/api"
)

type Config struct {
	SQLDriver     string
	SQLDSN        string
	Workers       int
	WriteInterval int
	CommitAt      int
}

var config *Config

func init() {
	config = &Config{
		SQLDriver:     "mysql",
		SQLDSN:        "root@/chronicle?parseTime=true",
		Workers:       20,
		WriteInterval: 100,
		CommitAt:      5000,
	}
	flag.StringVar(&(config.SQLDriver), "driver", "mysql", "database/sql driver")
	flag.StringVar(&(config.SQLDSN), "dsn", "", "database/sql DSN")
	flag.IntVar(&(config.Workers), "workers", 10, "Number of workers")
	flag.IntVar(&(config.WriteInterval), "interval", 100, "Database write interval in ms")
	flag.IntVar(&(config.CommitAt), "commitbreak", 5000, "How often to break to allow database to write")
	flag.Parse()
}

func main() {
	log.Println("Starting")

	db, err := api.NewDB(config.SQLDriver, config.SQLDSN, config.Workers, time.Duration(config.WriteInterval)*time.Millisecond)
	if err != nil {
		panic(err)
	}

	rows, err := db.DB.Query("SELECT uid, username, fullname, serial, clientidentifier, hostname, ip, internetip, time FROM chronicle;")

	if err != nil {
		panic(err)
	}

	defer rows.Close()

	var counter int64

	for rows.Next() {
		e := new(api.Entry)
		ci := new(sql.NullString)
		err = rows.Scan(&(e.UID),
			&(e.Username),
			&(e.FullName),
			&(e.Serial),
			ci,
			&(e.Hostname),
			&(e.IP),
			&(e.InternetIP),
			&(e.Time),
		)
		if err != nil {
			panic(err)
		}
		e.ClientIdentifier = ci.String
		db.Push(e)
		counter++
		if counter%int64(config.CommitAt) == 0 {
			time.Sleep(time.Duration(config.WriteInterval) * time.Millisecond)
		}
	}

	log.Println("Waiting for Queue to clear")
	for {
		if db.QueueLen() == 0 {
			break
		}
		time.Sleep(time.Second)
		log.Println(db.QueueLen(), "left in queue")
	}
	log.Println("Done")
}
