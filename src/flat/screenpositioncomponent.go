package flat

import "github.com/hajimehoshi/ebiten/v2"

type PositionAnchor int

const (
	UpperLeft PositionAnchor = iota
	UpperCenter
	UpperRight
	CenterLeft
	CenterCenter
	CenterRight
	LowerLeft
	LowerCenter
	LowerRight
)

type ScreenPositionComponent struct {
	ComponentBase
	Anchor                 PositionAnchor
	XPercentAwayFromAnchor float32
	YPercentAwayFromAnchor float32
	XPixelsAwayFromAnchor  int32
	YPixelsAwayFromAnchor  int32
}

func (s *ScreenPositionComponent) Draw(screen *ebiten.Image) {
	bounds := screen.Bounds()
	var startX, startY float32
	dirX, dirY := float32(1), float32(1)

	switch s.Anchor {
	case UpperLeft:
		startX, startY = 0, 0
	case UpperCenter:
		startX, startY = float32(bounds.Dx())/2, 0
	case UpperRight:
		startX, startY = float32(bounds.Dx()), 0
		dirX = -1
	case CenterLeft:
		startX, startY = 0, float32(bounds.Dy())/2
	case CenterCenter:
		startX, startY = float32(bounds.Dx())/2, float32(bounds.Dy())/2
	case CenterRight:
		startX, startY = float32(bounds.Dx()), float32(bounds.Dy())/2
		dirX = -1
	case LowerLeft:
		startX, startY = 0, float32(bounds.Dy())
		dirY = -1
	case LowerCenter:
		startX, startY = float32(bounds.Dx())/2, float32(bounds.Dy())
		dirY = -1
	case LowerRight:
		startX, startY = float32(bounds.Dx()), float32(bounds.Dy())
		dirX, dirY = -1, -1
	}

	x := startX + dirX*(float32(s.XPixelsAwayFromAnchor)+float32(bounds.Dx())*s.XPercentAwayFromAnchor/100)
	y := startY + dirY*(float32(s.YPixelsAwayFromAnchor)+float32(bounds.Dy())*s.YPercentAwayFromAnchor/100)
	t := s.GetTransform()
	t.Location.X = float64(x)
	t.Location.Y = float64(y)
}
