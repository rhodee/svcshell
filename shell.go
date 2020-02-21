package svcshell

import (
	"context"
	"net/http"
	"time"

	"github.com/oklog/run"
)

type muxer interface {
	http.Handler
}

// The "event" interfaces
// Logging, telemetry and shutdown
// were the obvious candidates
type interceptTelemetry interface {
	HandleLogging(b *AppShell)
}

type interceptLogging interface {
	HandleTelemetry(b *AppShell)
}

type interceptHandler interface {
	HandleShutdown(error)
}

// AppShellHandler provides the basic handlers this shell is
// concerned with: Logging, Telemetry and Shutdown.
type AppShellHandler interface {
	interceptLogging
	interceptTelemetry
	interceptHandler
}

// Prepare accepts a muxer to be wrapped in an app shell
// Which is provided back to the caller to run the server
func Prepare(addr string, mux muxer) *AppShell {
	asMux := appShellServer{mux}
	s := newServer(addr, &asMux)
	as := prepare(addr, &asMux)
	as.server = s
	return as
}

func prepare(addr string, mux muxer) *AppShell {
	as := AppShell{}
	as.router = mux
	as.group = run.Group{}
	as.group.Add(as.executor(addr, mux), as.interuptor)

	return &as
}

// AppShell this is an exported type for callers
// to use for binding event callbacks.
type AppShell struct {
	group          run.Group
	handleShutdown func(error)
	interceptHandler
	router http.Handler
	server *http.Server
}

func (as *AppShell) executor(addr string, mux muxer) func() error {
	return func() error {
		return as.server.ListenAndServe()
	}
}

func (as *AppShell) interuptor(err error) {
	as.handleShutdown(err)
	// shutdown own resources below this line

	// end by closing service
	as.server.Close()
}

func (as *AppShell) AfterLogging(loggingCb func(b *AppShell)) {
	loggingCb(as)
}

func (as *AppShell) AfterTelemetry(telemetryCb func(b *AppShell)) {
	telemetryCb(as)
}

func (as *AppShell) Start(ash AppShellHandler) error {
	as.handleShutdown = ash.HandleShutdown
	// Setup Default Logging below this line

	// Override Default Logging
	as.AfterLogging(ash.HandleLogging)

	// Setup Default Telemetry below this line

	// Override Default Telemetry
	as.AfterTelemetry(ash.HandleTelemetry)

	return as.group.Run()
}

type appShellCtx string

func (asc appShellCtx) String() string {
	return "appShell context key " + string(asc)
}

var (
	// AppShellValues is a context key
	// This is the sort of thing that should be exported so callers
	// can grab the value found at the key in their code and not care,
	// at all about how it got there.
	AppShellValues = appShellCtx("appshell-stuff")
)

type appShellServer struct {
	h http.Handler
}

func (s *appShellServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// For demo purposes only
	r.Header.Add("X-APPSHELL-ID", "appshell-1")

	// For demo purposes only
	ctx := context.WithValue(context.Background(), AppShellValues, "thunderdome")

	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	r = r.WithContext(ctx)

	s.h.ServeHTTP(w, r)
}

func newServer(address string, mux http.Handler) *http.Server {
	s := http.Server{
		Addr:           address,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        mux,
	}

	return &s
}
