package fcache

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/bhmj/goblocks/appstatus"
	"github.com/bhmj/goblocks/cache/dbcache"
	"github.com/bhmj/goblocks/dbase"
	"github.com/bhmj/goblocks/dbase/abstract"
	"github.com/bhmj/goblocks/file"
	"github.com/bhmj/goblocks/log"
	"github.com/bhmj/goblocks/metrics"
)

var errDBase = errors.New("error connecting to DB")

type Service struct {
	db             abstract.DB
	cfg            Config
	logger         log.MetaLogger
	cache          dbcache.Cache
	statusReporter appstatus.ServiceStatusReporter
}

// New returns fcache service instance
func New(
	cfg *Config,
	logger log.MetaLogger,
	_ *metrics.Registry,
	statusReporter appstatus.ServiceStatusReporter,
	production bool,
) (*Service, error) {
	// create database interface layer (migration inside)
	database := dbase.New(context.Background(), logger, cfg.DBase)
	if database == nil {
		return nil, fmt.Errorf("%w", errDBase)
	}

	// cache files path
	cacheDir, err := file.NormalizePath(cfg.CacheDir)
	if err != nil {
		return nil, fmt.Errorf("normalize path: %w", err)
	}

	cfg.Production = production

	svc := &Service{
		db:             database,
		cfg:            *cfg,
		logger:         logger,
		statusReporter: statusReporter,
		cache:          dbcache.New(database, logger, cacheDir),
	}
	return svc, nil
}

func (s *Service) Run(ctx context.Context) error {
	s.statusReporter.Ready()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.tokenRefresh(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.diskCleanup(ctx)
	}()

	<-ctx.Done()
	wg.Wait()

	return nil
}
