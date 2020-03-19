package svcshell

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oklog/run"
)

// Prepare accepts a muxer to be wrapped in an app shell
func Prepare(hostaddr string, mux muxer) *Shell {
	asMux := appShellServer{mux}
	s := newServer(hostaddr, &asMux)
	as := prepare(hostaddr, &asMux)
	as.server = s
	return as
}

func prepare(addr string, mux muxer) *Shell {
	as := Shell{}
	as.router = mux
	as.group = run.Group{}
	as.group.Add(as.executor, as.interrupt)

	return &as
}

// Shell is a concrete value for managing
// the server and lifecycle methods.
type Shell struct {
	interceptHandler
	group          run.Group
	handleShutdown func(error)
	router         http.Handler
	server         *http.Server
}

func (as *Shell) executor() error {
	return as.server.ListenAndServe()
}

func (as *Shell) interrupt(err error) {
	// shutdown own resources below this line
	as.handleShutdown(err)
}

// AfterLogging client override logging
func (as *Shell) AfterLogging(loggingCb func(s *Shell) *Shell) {
	as = loggingCb(as)
}

// AfterTelemetry client override telemetry
func (as *Shell) AfterTelemetry(telemetryCb func(s *Shell) *Shell) {
	as = telemetryCb(as)
}

// Start client side method for monitoring startup and shutdown of the managed process.
func (as *Shell) Start(h Handler) error {
	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-shutdown
		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Asking listener to shutdown and load shed.
		err := as.server.Shutdown(ctx)
		if err != nil {
			err = as.server.Close()
		}

		// Log the status of this shutdown.
		switch {
		case sig == syscall.SIGSTOP:
			as.HandleShutdown(errors.New("integrity issue caused shutdown"))
		case err != nil:
			as.HandleShutdown(errors.New("server was not stopped gracefully"))
		}
	}()

	as.handleShutdown = h.HandleShutdown
	// Setup Default Logging below this line

	// Override Default Logging
	as.AfterLogging(h.HandleLogging)

	// Setup Default Telemetry below this line

	// Override Default Telemetry
	as.AfterTelemetry(h.HandleTelemetry)

	return as.group.Run()
}

type appShellServer struct {
	h http.Handler
}

func (s *appShellServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// For demo purposes only
	r.Header.Add("X-APPSHELL-ID", "appshell-1")

	// For demo purposes only
	ctx := context.WithValue(context.Background(), DummyCtxValue, "thunderdome")

	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	r = r.WithContext(ctx)

	s.h.ServeHTTP(w, r)
}

func newServer(addr string, mux http.Handler) *http.Server {
	s := http.Server{
		Addr:           addr,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        mux,
	}

	return &s
}
