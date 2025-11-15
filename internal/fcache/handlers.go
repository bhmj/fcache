package fcache

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/bhmj/goblocks/httpreply"
)

const (
	responseModeContent = "content"
	responseModeURL     = "url"
)

var (
	errBadJSON        = errors.New("error decoding input JSON")
	errInvalidURL     = errors.New("invalid URL")
	errInvalidType    = errors.New("invalid file type")
	errInvalidRequest = errors.New("invalid request")
)

type getFileRequest struct {
	URL  string `json:"url"`
	Type string `json:"content_type"` // "url" or "content"
}

// GetCachedFile returns either filename on cache.domain.com, or actual file with contentType depending on "Type" request field
func (s *Service) GetCachedFile(w http.ResponseWriter, r *http.Request) (int, error) {
	var req getFileRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		return httpreply.Error(w, errBadJSON, http.StatusBadRequest) //nolint:wrapcheck
	}
	if req.URL == "" {
		return httpreply.Error(w, errInvalidURL, http.StatusBadRequest) //nolint:wrapcheck
	}

	return s.serveFile(w, req.URL, req.Type)
}

func (s *Service) getFilePath(url string) ([]byte, error) {
	path, err := s.cache.GetURL(url)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return []byte(path), nil
}

func (s *Service) getFileContent(url string) (body []byte, contentType string, err error) { //nolint:nonamedreturns
	return s.cache.GetContent(url) //nolint:wrapcheck
}

// StreamFile normally returns a cached file as a response, but also can return a cached filename.
// Used from main page, so auth is required.
func (s *Service) StreamFile(w http.ResponseWriter, r *http.Request) (int, error) {
	// check service cookie
	serviceCookie, err := r.Cookie("XID")
	if err != nil {
		return httpreply.Error(w, errInvalidRequest, http.StatusUnauthorized) //nolint:wrapcheck
	}
	if !s.isTokenValid(serviceCookie.Value) {
		return httpreply.Error(w, errInvalidRequest, http.StatusUnauthorized) //nolint:wrapcheck
	}

	// check arguments
	urls, found := r.Form["url"] // GET argument
	if !found {
		return httpreply.Error(w, errInvalidRequest, http.StatusBadRequest) //nolint:wrapcheck
	}
	url := urls[0]

	mode := responseModeContent // content by default
	modes, found := r.Form["mode"]
	if found {
		mode = modes[0]
	}
	if mode != responseModeURL && mode != responseModeContent {
		return httpreply.Error(w, errInvalidRequest, http.StatusBadRequest) //nolint:wrapcheck
	}

	return s.serveFile(w, url, mode)
}

func (s *Service) serveFile(w http.ResponseWriter, url string, mode string) (int, error) {
	var contentType string
	var content []byte
	var err error

	// rare case: some RSS feeds contain "//cdn.com/file" links
	if strings.HasPrefix(url, "//") {
		url = "https:" + url
	}

	switch mode {
	case responseModeURL:
		content, err = s.getFilePath(url)
	case responseModeContent:
		content, contentType, err = s.getFileContent(url)
	default:
		return httpreply.Error(w, errInvalidType, http.StatusBadRequest) //nolint:wrapcheck
	}
	if err != nil {
		return httpreply.Error(w, err, http.StatusInternalServerError) //nolint:wrapcheck
	}

	return httpreply.Reply(w, http.StatusOK, contentType, content) //nolint:wrapcheck
}
