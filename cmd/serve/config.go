package serve

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
	Srv    *SrvConfig
	Site   *SiteConfig
	Secret *SecretConfig
	Source string
}

// DBConfig database config
type DBConfig struct {
	Driver     string
	DataSource string
}

// SrvConfig server config
type SrvConfig struct {
	Host string
	Port string
}

// SiteConfig site config
type SiteConfig struct {
	Name        string
	BaseURL     string
	SessionName string
}

type SecretConfig struct {
	Session  string
	Password string
}

// DefaultConfig get default config
func DefaultConfig() (config *Config) {
	config = &Config{}
	db := &DBConfig{}
	db.Driver = "sqlite3"
	db.DataSource = "./bootstrap.db"

	srv := &SrvConfig{}
	srv.Host = "0.0.0.0"
	srv.Port = "1323"

	site := &SiteConfig{}
	site.Name = "Bootstrap"
	site.BaseURL = "http://127.0.0.1/"
	site.SessionName = "bs_sess"

	secret := &SecretConfig{}
	secret.Session = "secret"
	secret.Password = "secret"

	config.Mode = "development"
	config.DB = db
	config.Srv = srv
	config.Site = site
	config.Secret = secret
	config.Source = "./postcode.json"
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

	// server
	srv := &SrvConfig{}
	srv.Host = cfg.Section("srv").Key("host").String()
	if srv.Host == "" {
		srv.Host = config.Srv.Host
	}
	srv.Port = cfg.Section("srv").Key("port").String()
	if srv.Port == "" {
		srv.Port = config.Srv.Port
	}

	// site
	site := &SiteConfig{}
	site.Name = cfg.Section("site").Key("name").String()
	if site.Name == "" {
		site.Name = config.Site.Name
	}
	site.BaseURL = cfg.Section("site").Key("base_url").String()
	if site.BaseURL == "" {
		site.BaseURL = config.Site.BaseURL
	}
	site.SessionName = cfg.Section("site").Key("session_name").String()
	if site.SessionName == "" {
		site.SessionName = config.Site.SessionName
	}

	// secret
	secret := &SecretConfig{}
	secret.Session = cfg.Section("secret").Key("session").String()
	if secret.Session == "" {
		secret.Session = config.Secret.Session
	}
	secret.Password = cfg.Section("secret").Key("password").String()
	if secret.Password == "" {
		secret.Password = config.Secret.Password
	}

	// output
	source := cfg.Section("").Key("source").String()

	config.Mode = mode
	config.DB = db
	config.Srv = srv
	config.Site = site
	config.Secret = secret
	config.Source = source

	return
}
