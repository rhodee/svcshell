package svcshell

import (
	"fmt"
	"io"
	"log"
	"os"

	"net/http"
	"net/http/httptest"
	"testing"
)

// App is the closest thing to a knowledgeable object. It exists to be orchestrated
// by the app shell. Try not to decorate this a ton.
type App struct {
	Logger *log.Logger
	Mux    http.Handler
}

// New returns a new App service instance.
func New(mux http.Handler, logwriter io.Writer) App {
	logger := log.New(logwriter, " ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	a := App{
		Logger: logger,
		Mux:    mux,
	}
	return a
}

func (a App) HandleHello(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("APP SHELL HEADER=%s\n", r.Header.Get("X-APPSHELL-ID"))
	ctx := r.Context()
	asv, ok := ctx.Value(DummyCtxValue).(string)

	if ok {
		fmt.Printf("APP SHELL CTX VALUE=%s\n", asv)
	}

	fmt.Println("I am called after the app shell setup logging.")
}

// HandleLogging an appshell event handler for setting up the logger manually.
func (a App) HandleLogging(s *Shell) *Shell {
	a.Logger.Println("I am called after the app shell setup logging.")
	return s
}

// HandleTelemetry an appshell event handler for setting up telemetry manually.
func (a App) HandleTelemetry(s *Shell) *Shell {
	a.Logger.Println("I am called after the app shell setup telemetry.")
	return s
}

// HandleShutdown an appshell event handler called after shutdown has occurred in the app process.
func (a App) HandleShutdown(err error) {
	a.Logger.Printf("I am called when the process is forced to close due to error.")
}

func TestStart(t *testing.T) {
	// Provide an http.Handler here that dispatches to handlers.
	mux := http.NewServeMux()
	// Create a service to be managed by the AppShell.
	// This service can do whatever you need it to for satisfying
	// the logic it is responsible for. Managing the details like
	// managing telemetry and logging is provided by you or accept
	// the defaults.
	//
	// The app service has a to implement event handlers for logging,
	// telemetry and shutdown. They don't have to _do_ anything, but
	// they must be present.
	var logwriter io.Writer = os.Stdout

	appSvc := New(mux, logwriter)

	var _ Handler = (*App)(nil)

	mux.HandleFunc("/", appSvc.HandleHello)

	svc := Prepare(":8080", mux)

	go func() {
		// force a close to stop the process.
		svc.server.Close()
	}()

	err := svc.Start(appSvc)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	svc.router.ServeHTTP(rr, req)
	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if err == nil {
		t.Error("expected error but did not receive one.")
	}
}

func TestShellCtx(t *testing.T) {
	ctxKey := DummyCtxValue.String()

	if ctxKey == "" {
		t.Error("Expected context string value got nil")
	}
}
