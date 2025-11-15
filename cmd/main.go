package main

import (
	"fmt"
	syslog "log"

	"github.com/bhmj/fcache/internal/fcache"
	"github.com/bhmj/goblocks/app"
)

var appVersion = "local" //nolint:gochecknoglobals

func FcacheFactory(config any, options app.Options) (app.Service, error) {
	cfg, _ := config.(*fcache.Config)
	svc, err := fcache.New(cfg, options.Logger, options.MetricsRegistry, options.ServiceReporter, options.Production)
	if err != nil {
		return nil, fmt.Errorf("create cman service: %w", err)
	}
	return svc, nil
}

func main() {
	app := app.New("fcache app", appVersion)
	err := app.RegisterService("fcache", &fcache.Config{}, FcacheFactory) //nolint:exhaustruct
	if err != nil {
		syslog.Fatalf("register service: %v", err)
	}
	app.Run(nil)
}
