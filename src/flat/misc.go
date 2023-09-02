package flat

import (
	"fmt"
	"log"
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
