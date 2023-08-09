package flat

import "flatland/src/asset"

func RegisterAllFlatTypes() {
	asset.RegisterAsset(Image{})
	asset.RegisterAsset(World{})
	asset.RegisterAsset(ImageComponent{})
}
