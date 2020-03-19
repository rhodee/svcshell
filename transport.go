package svcshell

import "net/http"

type muxer interface {
	http.Handler
}

// The event interfaces to be managed:
// Logging, Telemetry and Shutdown
type interceptTelemetry interface {
	HandleLogging(s *Shell) *Shell
}

type interceptLogging interface {
	HandleTelemetry(s *Shell) *Shell
}

type interceptHandler interface {
	HandleShutdown(error)
}

// Handler provides the basic handlers this shell is
// concerned with: Logging, Telemetry and Shutdown.
type Handler interface {
	interceptLogging
	interceptTelemetry
	interceptHandler
}
