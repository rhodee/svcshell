package svcshell

type appShellCtx string

func (asc appShellCtx) String() string {
	return "appShell context key " + string(asc)
}

var (
	// DummyCtxValue is a context key
	// This is the sort of thing that should be exported so callers
	// can grab the value found at the key in their code and not care,
	// at all about how it got there.
	DummyCtxValue = appShellCtx("appshell-stuff")
)
