# Fruitroids

Fruitroids is a simple game that shows how to build Flatland projects.

To run the editor `go run cmd/editor/editor.go`

To run the game `go run cmd/fruitroids/main.go`

Fruitroids also runs on the web!  You will need to install wasmserve  
`go install github.com/hajimehoshi/wasmserve@latest`   
then   
`cd cmd/fruitroids/`  
`wasmserve`  
Open `localhost:8080`

# Package Dependency
The dependency tree looks like
cmd/editor
cmd/src/
flatland

The editor must depend on the game so it can edit the game-specific assets and run the game inside an editor window.