package game

import "github.com/go-gl/mathgl/mgl32"

func CheckCollision(a, b *Object) bool {
	collisionX := a.Position.X()+a.Size.X() >= b.Position.X() &&
		b.Position.X()+b.Size.X() >= a.Position.X()
	collisionY := a.Position.Y()+a.Size.Y() >= b.Position.Y() &&
		b.Position.Y()+b.Size.Y() >= a.Position.Y()

	return collisionX && collisionY
}

func CheckBallCollision(a *Ball, b *Object) bool {
	center := a.Position.Add(mgl32.Vec2{a.Radius, a.Radius})

	aabbHalfExtents := b.Size.Mul(0.5)
	aabbCenter := b.Position.Add(aabbHalfExtents)

	diff := center.Sub(aabbCenter)
	clamped := mgl32.Vec2{
		mgl32.Clamp(diff.X(), -aabbHalfExtents.X(), aabbHalfExtents.X()),
		mgl32.Clamp(diff.Y(), -aabbHalfExtents.Y(), aabbHalfExtents.Y()),
	}

	closest := aabbCenter.Add(clamped)

	diff = closest.Sub(center)

	return diff.Len() < a.Radius
}
