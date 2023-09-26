package editor

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/bradbev/flatland/src/asset"
	"github.com/bradbev/flatland/src/editor/edgui"
	"github.com/bradbev/flatland/src/flat"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"golang.org/x/exp/slices"
)

// This file contains the imgui specific workings of editor

const errorModalID = "ErrorModal##unique"

type ImguiEditor struct {
	// Link to the ebiten-imgui/renderer.Manager instance that is running the editor
	Manager    *renderer.Manager
	typeEditor *typeEditor

	fsysRead  fs.FS
	fsysWrite asset.WriteableFileSystem

	shouldQuit bool

	menu menuManager

	// The drawables that the editor will show
	drawables []Drawable

	knownEnums enumInfo

	pie pieManager

	// texture handling
	nextTextureID    imgui.TextureID
	embeddedTextures map[any]embeddedTexture
}

type embeddedTexture struct {
	img *ebiten.Image
	id  imgui.TextureID
}

type Drawable interface {
	// Draw allows an item to render itself.
	// if the returned error is not nil the drawable
	// will be removed from the draw list
	Draw() error
}

var closeDrawable = errors.New("Close")

func New(path string, manager *renderer.Manager) *ImguiEditor {
	asset.SetEditorMode()

	ed := &ImguiEditor{
		Manager:    manager,
		typeEditor: newTypeEditor(),

		fsysRead:  os.DirFS(path),
		fsysWrite: WriteFS(path),

		// fontAtlas is at ID 1, start high enough to avoid other IDs
		nextTextureID:    100,
		embeddedTextures: map[any]embeddedTexture{},

		knownEnums: enumInfo{
			enums: map[reflect.Type]map[int32]string{},
		},
	}
	ed.typeEditor.ed = ed
	ed.pie.ed = ed

	asset.RegisterFileSystem(ed.fsysRead, 0)
	asset.RegisterWritableFileSystem(ed.fsysWrite)

	contentWindow := newContentWindow(ed)
	ed.AddDrawable(contentWindow)

	ed.createMenu(&ed.menu)

	return ed
}

func (e *ImguiEditor) AddType(typeToAdd any, edit TypeEditorFn) {
	e.typeEditor.AddType(typeToAdd, edit)
}

func (e *ImguiEditor) Update(deltaseconds float32) error {
	e.menu.Draw()

	// iterate Drawables, then remove any that closed
	toRemove := map[Drawable]bool{}
	for _, d := range e.drawables {
		err := d.Draw()
		if err == closeDrawable {
			toRemove[d] = true
		}
	}
	e.drawables = slices.DeleteFunc(e.drawables, func(d Drawable) bool {
		return toRemove[d]
	})

	e.pie.Update(deltaseconds)
	if e.shouldQuit {
		return fmt.Errorf("Quit")
	}
	return nil
}

func (e *ImguiEditor) StartGameCallback(startGame func() ebiten.Game) {
	e.pie.StartGameCallback(startGame)
}

func (e *ImguiEditor) AddDrawable(d Drawable) {
	if !slices.ContainsFunc(e.drawables, func(existing Drawable) bool {
		return d == existing
	}) {
		e.drawables = append(e.drawables, d)
	}
}

func (e *ImguiEditor) isAssetBeingEdited(path string) bool {
	return slices.ContainsFunc(e.drawables, func(d Drawable) bool {
		win, ok := d.(*assetEditWindow)
		return ok && win.path == path
	})
}

func (e *ImguiEditor) EditAsset(path string) {
	// don't edit already open windows
	if e.isAssetBeingEdited(path) {
		return
	}

	loaded, err := asset.Load(asset.Path(path))
	if err != nil {
		fmt.Println(err)
		return
	}

	aew := newAssetEditWindow(path, loaded, NewTypeEditContext(e, path, loaded))

	e.AddDrawable(aew)
}

// GetImguiTexture creates a new ebiten.Image of size width, height and registers
// the image into the imgui texture system.  The return values are
// the id that can be used with imgui.Image() and the img can be used with
// ebiten code.
// When called repeatedly with the same key, no real work will be done, and
// cached values are returned.  If the size changes then the old texture is
// disposed
func (e *ImguiEditor) GetImguiTexture(key any, width int, height int) (id imgui.TextureID, img *ebiten.Image) {
	if tex, exists := e.embeddedTextures[key]; exists {
		s := tex.img.Bounds().Size()
		if s.X == width && s.Y == height {
			return tex.id, tex.img
		}
		// size must have changed
		tex.img.Dispose()
	}

	newImg := ebiten.NewImage(width, height)
	e.Manager.Cache.SetTexture(e.nextTextureID, newImg)
	tex := embeddedTexture{
		img: newImg,
		id:  e.nextTextureID,
	}
	e.embeddedTextures[key] = tex
	e.nextTextureID++
	return tex.id, tex.img
}

func (e *ImguiEditor) DisposeImguiTexture(key any) {
	if _, exists := e.embeddedTextures[key]; exists {
		// I can't actually dispose here because the texture is still going to be
		// used on the next Draw call.  It should be fine though because images
		// are automatically disposed when they are GC'd
		//tex.img.Dispose()
		delete(e.embeddedTextures, key)
	}
}

type fswalk struct {
	path  string
	dirs  []*fswalk
	files []string
}

func (e *ImguiEditor) buildFileCache() *fswalk {
	// todo - optimise this and cache state, filling it
	// as folders are opened
	stack := []*fswalk{{path: ""}}
	peek := func() *fswalk {
		return stack[len(stack)-1]
	}

	fs.WalkDir(e.fsysRead, ".", func(path string, d fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}
		// pop the stack if the incoming path doesn't match
		for {
			if strings.HasPrefix(path, peek().path) {
				break
			}
			stack = stack[:len(stack)-1]
		}
		if d != nil && d.IsDir() {
			// push dirs to the stack
			next := &fswalk{path: path}
			peek().dirs = append(peek().dirs, next)
			stack = append(stack, next)
			return nil
		}

		peek().files = append(peek().files, path)

		return nil
	})
	return stack[0]
}

func (ed *ImguiEditor) AddMenu(menu edgui.Menu) {
	ed.menu.AddMenu(menu)
}

func (ed *ImguiEditor) createMenu(m *menuManager) {
	i := func(name string, action func()) *edgui.MenuItem {
		return &edgui.MenuItem{Text: name, Action: func(*edgui.MenuItem) { action() }}
	}
	m.AddMenu(edgui.Menu{
		Name: "File",
		Items: []*edgui.MenuItem{
			i("Quit", func() { ed.shouldQuit = true }),
		},
	})
	m.AddMenu(edgui.Menu{
		Name: "Game",
		Items: []*edgui.MenuItem{
			i("Start PIE", func() { ed.pie.StartGame() }),
		},
	})
}

func (ed *ImguiEditor) RegisterEnum(mapping map[any]string) {
	var typ reflect.Type
	first := true
	for key := range mapping {
		if first {
			typ = reflect.TypeOf(key)
			first = false
		} else {
			t := reflect.TypeOf(key)
			flat.Assert(t == typ, "All keys in enum must be the same type")
		}
	}
	ed.knownEnums.RegisterEnum(typ, mapping)
}

type editorWriteFS struct {
	base string
}

func (e *editorWriteFS) WriteFile(path asset.Path, data []byte) error {
	return os.WriteFile(filepath.Join(e.base, string(path)), data, 0777)
}

func WriteFS(base string) asset.WriteableFileSystem {
	return &editorWriteFS{base: base}
}
