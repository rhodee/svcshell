package svcshell

import (
	"fmt"

	"net/http"
	"net/http/httptest"
	"testing"
)

type app struct{}

func (a app) HandleHello(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("APP SHELL HEADER=%s\n", r.Header.Get("X-APPSHELL-ID"))
	ctx := r.Context()
	asv, ok := ctx.Value(AppShellValues).(string)

	if ok {
		fmt.Printf("APP SHELL CTX VALUE=%s\n", asv)
	}

	fmt.Println("I am called after the app shell setup logging.")
}

func (a app) HandleLogging(as *AppShell) {
	fmt.Println("I am called after the app shell setup logging.")
}

func (a app) HandleTelemetry(as *AppShell) {
	fmt.Println("I am called after the app shell setup telemetry.")
}

func (a app) HandleShutdown(err error) {
	fmt.Println("I am called when the process is forced to close due to error.")
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
	appSvc := app{}
	mux.HandleFunc("/", appSvc.HandleHello)

	svc := Prepare(":8080", mux)

	// Just for testing purposes force an error to stop the process.
	go func() {
		svc.server.Close()
	}()

	err := svc.Start(appSvc)
	req, _ := http.NewRequest("GET", "/", nil)
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
	ctxKey := AppShellValues.String()

	if ctxKey == "" {
		t.Error("Expected context string value got nil")
	}
}
