# Quickstart

You'll need [Go](https://go.dev/dl/) installed.  The editor uses CGO, so you will need gcc installed.  On Windows I use https://jmeubank.github.io/tdm-gcc/.  On OSX I use the XCode installed toolchain.
I intend to keep the game binaries pure Go so that WASM can be targetted. 

From the root directory:
`go run cmd/editor/editor.go`  - run the test editor

`cd examples/fruitroids`  
`go run cmd/fruitroids/main.go` - run Fruitroids game
`go run cmd/editor/editor.go ` - run the editor for Fruitroids

Use VSCode - there are debug targets setup for each of the above.  You'll need to have the Go extension enabled.


# Testing
The `asset` package is pretty well tested.  Everything else, not so much.  
`go test ./...` from the root

I personally like seeing code coverage
`go test ./... -coverprofile=coverage.out`  
`go tool cover -html=coverage.out`