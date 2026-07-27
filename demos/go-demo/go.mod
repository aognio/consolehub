module consolehub-go-demo

go 1.25.0

require github.com/aognio/consolehub/libraries/go/consolehub v0.0.0

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/term v0.45.0 // indirect
)

replace github.com/aognio/consolehub/libraries/go/consolehub => ../../libraries/go/consolehub
