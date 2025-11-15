package fcache

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bhmj/goblocks/log"
)

var serviceToken struct { //nolint:gochecknoglobals
	sync.RWMutex
	PrevToken     string
	PrevExpiresAt time.Time
	Token         string
}

func (s *Service) tokenRefresh(ctx context.Context) {
	// init
	token, expAt := s.getToken() //nolint:contextcheck
	serviceToken.Token = token
	serviceToken.PrevToken = token
	serviceToken.PrevExpiresAt = expAt

	timer := time.NewTimer(time.Until(expAt))
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			s.logger.Info("getToken timer")
			token, expAt := s.getToken() //nolint:contextcheck
			serviceToken.Lock()
			serviceToken.PrevToken = serviceToken.Token
			serviceToken.PrevExpiresAt = time.Now().Add(5 * time.Minute) //nolint:mnd
			serviceToken.Token = token
			serviceToken.Unlock()
			timer = time.NewTimer(time.Until(expAt))
		}
	}
}

func (s *Service) getToken() (string, time.Time) {
	start := time.Now()
	for time.Since(start) < s.cfg.APITokenTimeout {
		token, expAt, err := s.pullToken()
		if err == nil {
			return token, expAt
		}
		s.logger.Error("getToken", log.Error(err))
		time.Sleep(5 * time.Second) //nolint:mnd
	}
	s.logger.Fatal("getToken: no response")
	return "", time.Time{}
}

func (s *Service) pullToken() (string, time.Time, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !s.cfg.Production},
	}
	client := &http.Client{Transport: transport} //nolint:exhaustruct
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.cfg.AuthDomain+"/api/token/get/", nil)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Api-Token", s.cfg.APIToken)
	resp, err := client.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var data struct {
		Token     string `json:"token"`
		ExpiresIn int    `json:"expires_in"`
	}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&data); err != nil {
		return "", time.Time{}, fmt.Errorf("response: %w", err)
	}
	now := time.Now()
	expAt := now.Add(time.Duration(data.ExpiresIn) * time.Second)
	s.logger.Info("getToken", log.Int("expiresIn", data.ExpiresIn), log.Time("now", now), log.Time("expires at", expAt))
	return data.Token, expAt, nil
}

func (s *Service) isTokenValid(token string) bool {
	serviceToken.RLock()
	defer serviceToken.RUnlock()
	return token == serviceToken.Token || (token == serviceToken.PrevToken && time.Until(serviceToken.PrevExpiresAt) > 0)
}
