package fruitroids

import (
	"fmt"
	"math/rand"

	"github.com/bradbev/flatland/src/asset"
	"github.com/bradbev/flatland/src/flat"
	"github.com/hajimehoshi/ebiten/v2"
)

type SpawnConfig struct {
	CountMin    int
	CountMax    int
	SpeedMax    float64
	RotationMax float64
	RoidType    *Roid
}

type LevelSpawn struct {
	flat.ActorBase
	ToSpawn []SpawnConfig `flat:"inline"`
}

func (l *LevelSpawn) BeginPlay() {
	l.ActorBase.BeginPlay(l)
	if ActiveWorld == nil {
		return
	}
	for _, toSpawn := range l.ToSpawn {
		rng := toSpawn.CountMax - toSpawn.CountMin
		for i := 0; i < rng+toSpawn.CountMin; i++ {
			a, err := asset.NewInstance(toSpawn.RoidType)
			if err != nil {
				fmt.Println("terrible")
			}
			r := a.(*Roid)
			w, h := ebiten.WindowSize()
			x, y := rand.Intn(w), rand.Intn(h)
			r.Transform.Location.X = float64(x)
			r.Transform.Location.Y = float64(y)

			r.velocity.X = rand.Float64() * toSpawn.SpeedMax
			r.velocity.Y = rand.Float64() * toSpawn.SpeedMax
			r.rotationDelta = rand.Float64()*toSpawn.RotationMax - (toSpawn.RotationMax / 2.0)

			ActiveWorld.AddToWorld(r)
		}
	}
}
