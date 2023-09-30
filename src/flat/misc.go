package flat

import (
	"fmt"
	"log"
	"reflect"
	"runtime/debug"

	"github.com/hajimehoshi/ebiten/v2"
)

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

func Assert(expr bool, message string) {
	if !expr {
		log.Fatal(message)
	}
}

// Errorf is just like fmt.Errorf, but appends a stack trace to the output
func Errorf(format string, args ...any) error {
	str := fmt.Sprintf(format, args...)
	stack := debug.Stack()
	return fmt.Errorf("%s\n------\n%s", str, stack)
}

// Check will panic if e is not nil, with e being passed to panic
func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func ApplyTransform(transform *Transform, geom *ebiten.GeoM) {
	geom.Scale(transform.ScaleX, transform.ScaleY)
	geom.Rotate(DegToRad(transform.Rotation))
	geom.Translate(transform.Location.X, transform.Location.Y)
}

// Apply all the transforms up the owning chain so that nested components
// can have relative transforms
func ApplyComponentTransforms(tr Transformer, geom *ebiten.GeoM) {
	ApplyTransform(tr.GetTransform(), geom)
	if comp, ok := tr.(Component); ok {
		if owningTransformer, ok := comp.Owner().(Transformer); ok {
			ApplyComponentTransforms(owningTransformer, geom)
		}
	}
}

func WalkUpComponentOwners(start Component, callback func(comp Component)) {
	for start != nil {
		callback(start)
		start = start.Owner()
	}
}

func TypeOf[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}
