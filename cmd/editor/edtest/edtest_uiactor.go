package edtest

import (
	"github.com/bradbev/flatland/src/flat"
)

type UiTestActor struct {
	flat.ActorBase
}

func (u *UiTestActor) BeginPlay() {
	u.ActorBase.BeginPlay(u)
}

func (u *UiTestActor) Update() {

}
