package flat

import "github.com/hajimehoshi/ebiten/v2"

func SliceToSet[T comparable](s []T) map[T]struct{} {
	ret := map[T]struct{}{}
	for _, elem := range s {
		ret[elem] = struct{}{}
	}
	return ret
}

func FrameTime() float64 {
	return 1.0 / float64(ebiten.TPS())
}
