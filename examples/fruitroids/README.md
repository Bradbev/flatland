# Fruitroids

Fruitroids is a simple game that shows how to build Flatland projects.

To run the editor `go run cmd/editor/editor.go`

To run the game `go run cmd/fruitroids/main.go`

# Package Dependency
The dependency tree looks like
cmd/editor
cmd/src/
flatland

The editor must depend on the game so it can edit the game-specific assets and run the game inside an editor window.