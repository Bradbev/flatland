# Flatland
Flatland is an attempt at writing a game-dev IDE using Go, imgui and Ebitengine.
I chose 2D because it is at least an order of magnitude less effort than 3D.

I've shamelessly stolen key ideas from Unreal, such as:

# Separate builds for standalone games and editor builds
Packages are organized such that `editor` depends on the game.  The game part
does not need to use anything from the editor package, or from `imgui` if you
don't want to.

# Easy editing of user defined assets
If you register a struct with the `asset` package (`asset.RegisterAsset`), then
the default `editor` package will let you create and edit the exported fields of
the struct.
Likewise it is trivial to load assets with `asset.Load`, this loading will
recursively load any assets that are linked.

# Custom asset editors are easy to write
If you know a little imgui, you can write custom editors for your types, see
`image_ed.go` for an example.