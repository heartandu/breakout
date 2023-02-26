package game

import (
	"github.com/go-gl/mathgl/mgl32"

	"breakout/src/texture"
)

type Ball struct {
	Radius float32
	Stuck  bool

	Sticky      bool
	PassThrough bool

	Object
}

func NewBall(
	position mgl32.Vec2,
	radius float32,
	velocity mgl32.Vec2,
	sprite *texture.Texture2D,
) *Ball {
	return &Ball{
		Radius: radius,
		Stuck:  true,
		Object: *NewObject(
			position,
			mgl32.Vec2{radius * 2, radius * 2},
			sprite,
			&mgl32.Vec3{1, 1, 1},
			&velocity,
		),
	}
}

func (b *Ball) Move(dt float64, windowWidth int) mgl32.Vec2 {
	if !b.Stuck {
		b.Position = b.Position.Add(b.Velocity.Mul(float32(dt)))

		if b.Position.X() <= 0 {
			b.Velocity[0] = -b.Velocity[0]
			b.Position[0] = 0
		} else if b.Position.X()+b.Size.X() >= float32(windowWidth) {
			b.Velocity[0] = -b.Velocity[0]
			b.Position[0] = float32(windowWidth) - b.Size.X()
		} else if b.Position.Y() <= 0 {
			b.Velocity[1] = -b.Velocity[1]
			b.Position[1] = 0
		}
	}

	return b.Position
}

func (b *Ball) Reset(position, velocity mgl32.Vec2) {
	b.Position = position
	b.Velocity = velocity
	b.Stuck = true
}
