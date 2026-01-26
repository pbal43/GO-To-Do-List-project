package internal

import (
	"cmp"
	"encoding/json"
	"flag"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
)

const (
	defDNS            = "postgres://postgres:postgres@localhost/postgres?sslmode=disable"
	defHost           = "0.0.0.0"
	defPort           = 8080
	defDebug          = false
	defMigrationsPath = "migrations"
	defTaskCapacity   = 10
	defSecureProtocol = false
)

type Config struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	DNS            string `json:"dns"`
	MigratePath    string `json:"migrate_path"`
	Debug          bool   `json:"debug"`
	TaskCapacity   int    `json:"task_capacity"`
	SecureProtocol bool   `json:"secure_protocol"`
	CertCert       string `json:"cert_cert"`
	KeyCert        string `json:"key_cert"`
}

type Flags struct {
	ConfigPath     string
	Host           string
	Port           int
	DNS            string
	MigratePath    string
	Debug          bool
	TaskCapacity   int
	SecureProtocol bool
	CertCert       string
	KeyCert        string
}

// Дефолты не указывал, так как заданы отдельно.
func parseFlags() Flags {
	var flags Flags

	flag.StringVar(&flags.ConfigPath, "c", "", "Path to config file")
	flag.StringVar(&flags.Host, "host", "", "Server host")
	flag.IntVar(&flags.Port, "port", 0, "Server port")
	flag.StringVar(&flags.DNS, "dns", "", "DB CONNECTION STRING")
	flag.StringVar(&flags.MigratePath, "migrate-path", "", "Path to migrations folder")
	flag.BoolVar(&flags.Debug, "debug", false, "Debug mode")
	flag.IntVar(&flags.TaskCapacity, "task-capacity", 0, "Task capacity")
	flag.BoolVar(&flags.SecureProtocol, "s", false, "Use HTTPS")
	flag.StringVar(&flags.CertCert, "cert", "", "Path to Cert file")
	flag.StringVar(&flags.KeyCert, "key-cert", "", "Path to Cert Key file")

	flag.Parse()

	return flags
}

func configFromFlags(flags *Flags) Config {
	return Config{
		Host:           flags.Host,
		Port:           flags.Port,
		DNS:            flags.DNS,
		MigratePath:    flags.MigratePath,
		Debug:          flags.Debug,
		TaskCapacity:   flags.TaskCapacity,
		SecureProtocol: flags.SecureProtocol,
		CertCert:       flags.CertCert,
		KeyCert:        flags.KeyCert,
	}
}

func configFromEnv() Config {
	cfg := Config{}

	cfg.Host = os.Getenv("HOST")
	cfg.Port, _ = strconv.Atoi(os.Getenv("PORT"))
	cfg.DNS = os.Getenv("DB_DNS")
	cfg.MigratePath = os.Getenv("MIGRATE_PATH")
	cfg.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	cfg.TaskCapacity, _ = strconv.Atoi(os.Getenv("TASK_CAPACITY"))
	cfg.SecureProtocol, _ = strconv.ParseBool(os.Getenv("SECURE_PROTOCOL"))
	cfg.CertCert = os.Getenv("CERT_FILE")
	cfg.KeyCert = os.Getenv("KEY_FILE")

	return cfg
}

func configFromFile(path string) Config {
	cfg := Config{}

	if path == "" {
		log.Info().Msg("Config file path is empty")
		return cfg
	}

	data, err := os.ReadFile(path)

	if err != nil {
		log.Info().Err(err).Msg("Config file read failed")
		return cfg
	}

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		log.Info().Err(err).Msg("Config file unmarshal failed")
		return cfg
	}

	return cfg
}

func defaultConfig() Config {
	return Config{
		Host:           defHost,
		Port:           defPort,
		DNS:            defDNS,
		MigratePath:    defMigrationsPath,
		Debug:          defDebug,
		TaskCapacity:   defTaskCapacity,
		SecureProtocol: defSecureProtocol,
	}
}

// ReadConfig - чтение конфига приложения.
// Может не подойти при расширении/изменении, так как про cmp.Or возвращает zero values только после проверки всех аргументов.
func ReadConfig() Config {
	config := Config{}

	flags := parseFlags()
	flagCfg := configFromFlags(&flags)
	envCfg := configFromEnv()
	fileCfg := configFromFile(flags.ConfigPath)
	defCfg := defaultConfig()

	config.Host = cmp.Or(
		flagCfg.Host,
		envCfg.Host,
		fileCfg.Host,
		defCfg.Host,
	)

	config.Port = cmp.Or(
		flagCfg.Port,
		envCfg.Port,
		fileCfg.Port,
		defCfg.Port,
	)

	config.DNS = cmp.Or(
		flagCfg.DNS,
		envCfg.DNS,
		fileCfg.DNS,
		defCfg.DNS,
	)

	config.MigratePath = cmp.Or(
		flagCfg.MigratePath,
		envCfg.MigratePath,
		fileCfg.MigratePath,
		defCfg.MigratePath,
	)

	config.Debug = cmp.Or(
		flagCfg.Debug,
		envCfg.Debug,
		fileCfg.Debug,
		defCfg.Debug,
	)

	config.TaskCapacity = cmp.Or(
		flagCfg.TaskCapacity,
		envCfg.TaskCapacity,
		fileCfg.TaskCapacity,
		defCfg.TaskCapacity,
	)

	config.SecureProtocol = cmp.Or(
		flagCfg.SecureProtocol,
		envCfg.SecureProtocol,
		fileCfg.SecureProtocol,
		defCfg.SecureProtocol,
	)

	config.CertCert = cmp.Or(
		flagCfg.CertCert,
		envCfg.CertCert,
		fileCfg.CertCert,
		defCfg.CertCert,
	)

	config.KeyCert = cmp.Or(
		flagCfg.KeyCert,
		envCfg.KeyCert,
		fileCfg.KeyCert,
		defCfg.KeyCert,
	)

	if config.CertCert == "" || config.KeyCert == "" {
		config.SecureProtocol = false
	}

	return config
}
