package internal

import (
	"cmp"
	"flag"
	"os"
)

const (
	defDNS            = "postgres://postgres:postgres@localhost/postgres?sslmode=disable"
	defHost           = "0.0.0.0"
	defPort           = 8080
	defMigrationsPath = "migrations"
	defTaskCapacity   = 10
)

type Config struct {
	Host         string
	Port         int
	DNS          string
	MigratePath  string
	Debug        bool
	TaskCapacity int
}

func ReadConfig() Config {
	var config Config
	flag.StringVar(&config.Host, "host", defHost, "Server host")
	flag.IntVar(&config.Port, "port", defPort, "Server port")
	flag.StringVar(&config.DNS, "dns", defDNS, "DB CONNECTION STRING")
	flag.StringVar(&config.MigratePath, "migrate-path", defMigrationsPath, "Path to migrations folder")
	flag.BoolVar(&config.Debug, "debug", false, "Debug mode (уровень логирования)")
	flag.IntVar(&config.TaskCapacity, "task-capacity", defTaskCapacity, "Task capacity for delete")
	flag.Parse()

	config.DNS = cmp.Or(os.Getenv("DB_DNS"), defDNS)
	config.MigratePath = cmp.Or(os.Getenv("MIGRATE_PATH"), defMigrationsPath)

	return config
}
