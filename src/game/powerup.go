package game

import (
	"github.com/go-gl/mathgl/mgl32"

	"breakout/src/texture"
)

var (
	powerUpSize     = mgl32.Vec2{60, 20}
	powerUpVelocity = mgl32.Vec2{0, 150}
)

type PowerUp struct {
	Type      string
	Duration  float64
	Activated bool

	Object
}

func NewPowerUp(
	t string,
	color mgl32.Vec3,
	duration float64,
	position mgl32.Vec2,
	tex *texture.Texture2D,
) PowerUp {
	return PowerUp{
		Type:      t,
		Duration:  duration,
		Activated: false,
		Object: *NewObject(
			position,
			powerUpSize,
			tex,
			&color,
			&powerUpVelocity,
		),
	}
}
