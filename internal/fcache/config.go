package fcache

import (
	"time"

	"github.com/bhmj/goblocks/dbase"
)

const DefaultPort int = 80

// Config contains all parameters
type Config struct {
	APIBase         string        `yaml:"api_base"`
	DBase           dbase.Config  `yaml:"dbase" group:"Database configuration"`
	CacheDir        string        `yaml:"cache_dir" description:"Cache directory"`
	AuthDomain      string        `yaml:"auth_domain" description:"Auth domain"`
	APIToken        string        `yaml:"api_token" description:"API token"`
	APITokenTimeout time.Duration `yaml:"api_token_timeout" description:"API token pull timeout (until panic)" default:"5m"`
	Production      bool
}
