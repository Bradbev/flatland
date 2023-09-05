package flat

import "github.com/bradbev/flatland/src/asset"

func RegisterAllFlatTypes() {
	asset.RegisterAsset(Image{})
	asset.RegisterAsset(World{})
	asset.RegisterAsset(ImageComponent{})
	asset.RegisterAsset(Font{})
	asset.RegisterAsset(ActorBase{})
	asset.RegisterAsset(Text{})
}
