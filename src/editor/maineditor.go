package editor

import (
	"errors"
	"flatland/src/asset"
	"fmt"
	"io/fs"
	"os"
	"reflect"
	"strings"
	"unsafe"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"golang.org/x/exp/slices"
)

// This file contains the imgui specific workings of editor

const errorModalID = "ErrorModal##unique"

type ImguiEditor struct {
	// Link to the ebiten-imgui/renderer.Manager instance that is running the editor
	Manager *renderer.Manager
	// context should be used by custom editors to store their own data across
	// calls to the edit function
	context    map[unsafe.Pointer]map[any]any
	typeEditor *typeEditor

	fsysRead  fs.FS
	fsysWrite asset.WriteableFileSystem

	drawables []Drawable

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
	ret := &ImguiEditor{
		Manager:    manager,
		context:    map[unsafe.Pointer]map[any]any{},
		typeEditor: newTypeEditor(),

		fsysRead:  os.DirFS(path),
		fsysWrite: WriteFS(path),

		// fontAtlas is at ID 1, start high enough to avoid other IDs
		nextTextureID:    100,
		embeddedTextures: map[any]embeddedTexture{},
	}
	ret.typeEditor.ed = ret

	asset.RegisterFileSystem(ret.fsysRead, 0)
	asset.RegisterWritableFileSystem(ret.fsysWrite)

	contentWindow := newContentWindow(ret)
	ret.AddDrawable(contentWindow)

	return ret
}

// Get a Context item from the ImguiEditor.  Custom editors should
// use this function to save off context during edits
// returns true if this is the first time the context has been created
func GetContext[T any](ed *ImguiEditor, key reflect.Value) (*T, bool) {
	ptr := key.Addr().UnsafePointer()
	contexts, contextsExists := ed.context[ptr]
	if !contextsExists {
		// first level map doesn't exist yet
		contexts = map[any]any{}
		ed.context[ptr] = contexts
	}
	var zeroT T

	// found the type/value context
	if context, exists := contexts[zeroT]; exists {
		return context.(*T), false
	}

	// second level map doesn't exist
	ret := new(T)
	contexts[zeroT] = ret
	return ret, true
}

func (e *ImguiEditor) AddType(typeToAdd any, edit TypeEditorFn) {
	e.typeEditor.AddType(typeToAdd, edit)
}

func (e *ImguiEditor) Edit(obj any) {
	e.typeEditor.Edit(obj)
}

func (e *ImguiEditor) EditValue(value reflect.Value) {
	e.typeEditor.EditValue(value)
}

func (e *ImguiEditor) Update(deltaseconds float32) error {
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
	return nil
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
	aew := &assetEditWindow{
		path:   path,
		target: loaded,
		editor: e,
	}

	e.AddDrawable(aew)
}

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
		if d != nil && d.IsDir() {
			next := &fswalk{path: path}
			peek().dirs = append(peek().dirs, next)
			stack = append(stack, next)
			return nil
		}
		for {
			if strings.HasPrefix(path, peek().path) {
				break
			}
			stack = stack[:len(stack)-1]
		}
		peek().files = append(peek().files, path)

		return nil
	})
	return stack[0]
}
