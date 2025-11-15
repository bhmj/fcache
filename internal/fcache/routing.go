package fcache

import (
	"fmt"
	"strings"

	"github.com/bhmj/goblocks/app"
)

type ContextKey string

const ContextValueRequestID ContextKey = "requestID"

// GetHandlers returns a list of handlers for the server
func (s *Service) GetHandlers() []app.HandlerDefinition {
	apiBase := strings.Trim(s.cfg.APIBase, "/")
	api := func(path string) string {
		return fmt.Sprintf("/%s/%s", apiBase, strings.TrimPrefix(path, "/"))
	}
	return []app.HandlerDefinition{
		{Endpoint: "/get_file", Method: "POST", Path: api("/get_file/"), Func: s.GetCachedFile},
		{Endpoint: "/stream_file", Method: "GET", Path: api("/stream/"), Func: s.StreamFile},
	}
}
