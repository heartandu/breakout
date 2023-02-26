package game

import "github.com/go-gl/mathgl/mgl32"

type Collision struct {
	Collided bool
	Dir      Direction
	Diff     mgl32.Vec2
}

type Direction int

const (
	UpDirection Direction = iota
	RightDirection
	DownDirection
	LeftDirection
)

func CheckCollision(a, b *Object) bool {
	collisionX := a.Position.X()+a.Size.X() >= b.Position.X() &&
		b.Position.X()+b.Size.X() >= a.Position.X()
	collisionY := a.Position.Y()+a.Size.Y() >= b.Position.Y() &&
		b.Position.Y()+b.Size.Y() >= a.Position.Y()

	return collisionX && collisionY
}

func CheckBallCollision(a *Ball, b *Object) Collision {
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

	if diff.Len() < a.Radius {
		return Collision{
			Collided: true,
			Dir:      VectorDirection(diff),
			Diff:     diff,
		}
	}

	return Collision{
		Collided: false,
		Dir:      UpDirection,
		Diff:     mgl32.Vec2{0, 0},
	}
}

func VectorDirection(target mgl32.Vec2) Direction {
	var (
		max       float32
		bestMatch = -1
		compass   = []mgl32.Vec2{
			{0, 1},
			{1, 0},
			{0, -1},
			{-1, 0},
		}
	)

	for i := 0; i < 4; i++ {
		dotProduct := target.Normalize().Dot(compass[i])
		if dotProduct > max {
			max = dotProduct
			bestMatch = i
		}
	}

	return Direction(bestMatch)
}
