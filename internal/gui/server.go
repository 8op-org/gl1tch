package gui

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/provider"
	"github.com/8op-org/gl1tch/internal/store"
)

//go:embed all:dist
var distFS embed.FS

// Server is the workflow GUI HTTP server.
type Server struct {
	addr        string
	workspace   string
	dev         bool
	store       *store.Store
	mux         *http.ServeMux
	providerReg *provider.ProviderRegistry
}

// New creates a GUI server for the given workspace.
func New(addr, workspace string, dev bool) (*Server, error) {
	st, err := store.Open()
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}
	var reg *provider.ProviderRegistry
	if home, err := os.UserHomeDir(); err == nil {
		reg, _ = provider.LoadProviders(filepath.Join(home, ".config", "glitch", "providers"))
	}
	s := &Server{
		addr:        addr,
		workspace:   workspace,
		dev:         dev,
		store:       st,
		mux:         http.NewServeMux(),
		providerReg: reg,
	}
	s.routes()
	return s, nil
}

// newTelemetry creates a telemetry client if ES is reachable.
func newTelemetry() *esearch.Telemetry {
	esClient := esearch.NewClient("http://localhost:9200")
	if err := esClient.Ping(context.Background()); err == nil {
		tel := esearch.NewTelemetry(esClient)
		tel.EnsureIndices(context.Background())
		return tel
	}
	return nil
}

func (s *Server) routes() {
	// API routes — handlers will be added in subsequent tasks
	s.mux.HandleFunc("GET /api/workflows", s.handleListWorkflows)
	s.mux.HandleFunc("GET /api/workflows/actions/{context}", s.handleGetWorkflowActions)
	s.mux.HandleFunc("GET /api/workflows/{name}", s.handleGetWorkflow)
	s.mux.HandleFunc("PUT /api/workflows/{name}", s.handlePutWorkflow)
	s.mux.HandleFunc("POST /api/workflows/{name}/run", s.handleRunWorkflow)
	s.mux.HandleFunc("GET /api/runs", s.handleListRuns)
	s.mux.HandleFunc("GET /api/runs/{id}", s.handleGetRun)
	s.mux.HandleFunc("GET /api/results/{path...}", s.handleGetResult)
	s.mux.HandleFunc("PUT /api/results/{path...}", s.handlePutResult)
	s.mux.HandleFunc("GET /api/kibana/workflow/{name}", s.handleKibanaWorkflow)
	s.mux.HandleFunc("GET /api/kibana/run/{id}", s.handleKibanaRun)
	s.mux.HandleFunc("GET /api/workspace", s.handleGetWorkspace)

	if !s.dev {
		dist, _ := fs.Sub(distFS, "dist")
		fileServer := http.FileServer(http.FS(dist))
		s.mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			// SPA fallback: if the path isn't a real file, serve index.html
			if r.URL.Path != "/" {
				_, err := fs.Stat(dist, r.URL.Path[1:])
				if err != nil {
					r.URL.Path = "/"
				}
			}
			fileServer.ServeHTTP(w, r)
		})
	}
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	srv := &http.Server{
		Addr:         s.addr,
		Handler:      s.mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	return srv.ListenAndServe()
}

// Close releases server resources.
func (s *Server) Close() error {
	return s.store.Close()
}

// workflowsDir returns the path to the workspace workflows directory.
func (s *Server) workflowsDir() string {
	return s.workspace + "/workflows"
}

// resultsDir returns the path to the workspace results directory.
func (s *Server) resultsDir() string {
	return s.workspace + "/results"
}
