package asset_test

import (
	"flatland/src/asset"
	"io/fs"
	"testing"

	"github.com/psanford/memfs"
	"github.com/stretchr/testify/assert"
)

type testAsset struct {
	Anykey         string
	postLoadCalled bool
}

func (t *testAsset) PostLoad() {
	t.postLoadCalled = true
}

func TestAssetLoad(t *testing.T) {
	data := `{
	"Type": "flatland/src/asset_test.testAsset",
	"Inner": {
		"anykey": "hi"
	}
}`
	testAssetLoad := func(callback func()) {
		defer asset.Reset()
		name := "asset.json"
		callback()
		newTestAsset(name, data)

		a, err := asset.Load(name)
		assert.NoError(t, err)
		toTest, ok := a.(*testAsset)
		assert.True(t, ok, "Unable to convert %v to testType", a)
		assert.Equal(t, toTest.Anykey, "hi")
		assert.True(t, toTest.postLoadCalled, "PostLoad was not called")
	}
	testAssetLoad(func() {
		t.Log("Testing RegisterAsset")
		asset.RegisterAsset(testAsset{})
	})
	testAssetLoad(func() {
		t.Log("Testing RegisterAssetFactory")
		asset.RegisterAssetFactory(testAsset{}, func() asset.Asset { return &testAsset{} })
	})
}

func newTestAsset(name string, data string) {
	rootFS := memfs.New()
	rootFS.WriteFile(name, []byte(data), 0777)
	asset.RegisterFileSystem(rootFS, 0)
}

type writeFS struct {
	fs *memfs.FS
}

func newWriteFS() *writeFS {
	return &writeFS{
		fs: memfs.New(),
	}
}

func (f *writeFS) WriteFile(path string, data []byte) error {
	return f.fs.WriteFile(path, data, 0777)
}

func TestAssetSave(t *testing.T) {
	defer asset.Reset()
	wfs := newWriteFS()
	asset.RegisterWritableFileSystem(wfs)
	asset.RegisterAsset(testAsset{})

	name := "asset.json"
	a := &testAsset{Anykey: "saved"}
	err := asset.Save(name, a)
	assert.NoError(t, err)

	back, err := fs.ReadFile(wfs.fs, name)
	assert.NoError(t, err)
	expected := `{
 "Type": "flatland/src/asset_test.testAsset",
 "Inner": {
  "Anykey": "saved"
 }
}`
	assert.Equal(t, expected, string(back))

	asset.RegisterFileSystem(wfs.fs, 0)
	loaded, err := asset.Load(name)
	assert.NoError(t, err)

	expectedObj := &testAsset{Anykey: "saved", postLoadCalled: true}
	assert.Equal(t, expectedObj, loaded)
}
