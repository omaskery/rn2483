module github.com/omaskery/rn2483/examples

go 1.16

replace github.com/omaskery/rn2483 => ../

require (
	github.com/alecthomas/kong v0.2.17
	github.com/augustoroman/hexdump v0.0.0-20190827031536-6506f4163e93
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/stdr v0.4.0
	github.com/jacobsa/go-serial v0.0.0-20180131005756-15cf729a72d4
	github.com/omaskery/rn2483 v0.0.0-00010101000000-000000000000
	go.uber.org/multierr v1.7.0
	golang.org/x/sys v0.1.0 // indirect
)
