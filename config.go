package main

import (
	"log"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

//Config represents options given in the environment
type Config struct {
	SQLDriver string //required
	SQLDSN    string //required

	Workers       int //default: 10
	WriteInterval int //in seconds; default:15s

	ListenAddr string //addr format used for net.Dial; required
	Prefix     string //url prefix to mount api to without trailing slash
}

var config = &Config{}

func checkEmpty(val, name string) {
	if val == "" {
		log.Fatalf("CHRONICLE_%s must be configured\n", name)
	}
}

func init() {
	err := envconfig.Process("CHRONICLE", config)
	if err != nil {
		log.Fatalln("Error reading configuration from environment:", err)
	}

	checkEmpty(config.SQLDriver, "SQLDriver")
	checkEmpty(config.SQLDSN, "SQLDSN")

	if config.SQLDriver == "mysql" && !strings.Contains(config.SQLDSN, "?parseTime=true") {
		log.Fatalln("mysql DSN must contain \"?parseTime=true\"")
	}

	if config.Workers == 0 {
		config.Workers = 10
	}

	if config.WriteInterval == 0 {
		config.WriteInterval = 15
	}

	checkEmpty(config.ListenAddr, "LISTENADDR")
}
