package collect

import (
	"gopkg.in/ini.v1"
)

var (
	Conf  *Config
	Debug bool
)

// Config app config
type Config struct {
	Mode   string
	DB     *DBConfig
	AppKey string
	Cache  string
	Output string
}

// DBConfig database config
type DBConfig struct {
	Driver     string
	DataSource string
}

// DefaultConfig get default config
func DefaultConfig() (config *Config) {
	config = &Config{}
	db := &DBConfig{}
	db.Driver = "sqlite3"
	db.DataSource = "./postcode.db"

	config.Mode = "development"
	config.AppKey = ""
	config.DB = db
	config.Cache = "./cache"
	config.Output = "./postcode.json"
	return
}

// Load load config from file override default config
func (config *Config) Load(path string) (err error) {

	// load from ini file
	cfg, err := ini.Load(path)
	if err != nil {
		return err
	}

	// run mode
	mode := cfg.Section("").Key("mode").In("development", []string{"development", "production", "testing"})

	// database
	db := &DBConfig{}
	db.Driver = cfg.Section("db").Key("driver").In("sqlite3", []string{"sqlite3", "mysql"})
	db.DataSource = cfg.Section("db").Key("dataSource").String()
	if db.DataSource == "" {
		db.DataSource = config.DB.DataSource
	}

	// appkey
	appKey := cfg.Section("").Key("app_key").String()

	// cache
	cache := cfg.Section("").Key("cache").String()

	// output
	output := cfg.Section("").Key("output").String()

	config.Mode = mode
	config.DB = db
	config.AppKey = appKey
	config.Cache = cache
	config.Output = output

	return
}
