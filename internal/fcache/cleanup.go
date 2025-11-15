package fcache

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/bhmj/goblocks/file"
	"github.com/bhmj/goblocks/log"
)

const expireIn = time.Hour * 24 * 4 // delete cached files 4 days and older

// disk cleanup process (once in 1 hour)
func (s *Service) diskCleanup(ctx context.Context) {
	timer := time.NewTicker(time.Hour)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			s.logger.Info("diskCleanup: timer")
			err := s.doCleanup(expireIn)
			if err != nil {
				s.logger.Error("doCleanup: query expired", log.Error(err))
			} else {
				s.logger.Info("diskCleanup: ok")
			}
		}
	}
}

type cachedFile struct {
	FilePath string `db:"file_path"`
}

// implementation
func (s *Service) doCleanup(age time.Duration) error {
	expirationTime := time.Now().Add(-age)
	var files []cachedFile
	err := s.db.Query(&files, `select file_path from file_cache where last_read_at < $1`, expirationTime)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	for _, f := range files {
		filePath := filepath.Join(s.cfg.CacheDir, f.FilePath)
		_ = file.Delete(filePath)
	}
	return s.db.Exec(`delete from file_cache where last_read_at < $1`, expirationTime) //nolint:wrapcheck
}
