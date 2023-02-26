package game

import (
	"github.com/go-gl/mathgl/mgl32"

	"breakout/src/render"
	"breakout/src/texture"
)

type Object struct {
	Position  mgl32.Vec2
	Size      mgl32.Vec2
	Velocity  mgl32.Vec2
	Color     mgl32.Vec3
	Rotation  float32
	IsSolid   bool
	Destroyed bool

	Sprite *texture.Texture2D
}

func NewObject(
	position, size mgl32.Vec2,
	sprite *texture.Texture2D,
	color *mgl32.Vec3,
	velocity *mgl32.Vec2,
) *Object {
	if color == nil {
		color = &mgl32.Vec3{1, 1, 1}
	}

	if velocity == nil {
		velocity = &mgl32.Vec2{0, 0}
	}

	return &Object{
		Position:  position,
		Size:      size,
		Velocity:  *velocity,
		Color:     *color,
		Rotation:  0,
		IsSolid:   false,
		Destroyed: false,
		Sprite:    sprite,
	}
}

func (g *Object) Draw(renderer *render.SpriteRenderer) {
	renderer.DrawSprite(g.Sprite, &g.Position, &g.Size, g.Rotation, &g.Color)
}
