package game

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"

	"breakout/src/render"
	"breakout/src/resource"
)

type GameState int

const (
	GameActive GameState = iota
	GameMenu
	GameWin
)

type Game struct {
	State  GameState
	Keys   [1024]bool
	Width  int
	Height int

	Renderer *render.SpriteRenderer
}

func NewGame(width, height int) *Game {
	return &Game{
		State:  GameActive,
		Width:  width,
		Height: height,
	}
}

func (g *Game) Cleanup() {
	if g.Renderer != nil {
		g.Renderer.Cleanup()
	}
}

func (g *Game) Init() error {
	err := resource.LoadShader(
		"resources/shaders/sprite.vert",
		"resources/shaders/sprite.frag",
		"",
		"sprite",
	)
	if err != nil {
		return fmt.Errorf("failed to load shader: %w", err)
	}

	projection := mgl32.Ortho(0, float32(g.Width), float32(g.Height), 0, -1, 1)
	resource.GetShader("sprite").SetInteger("image", 0, true)
	resource.GetShader("sprite").SetMatrix4("projection", &projection, false)

	g.Renderer = render.NewSpriteRenderer(resource.GetShader("sprite"))

	err = resource.LoadTexture("resources/textures/happy.png", true, "face")
	if err != nil {
		return fmt.Errorf("failed to load texture: %w", err)
	}

	return nil
}

func (g *Game) ProcessInput(dt float64) {

}

func (g *Game) Update(dt float64) {

}

func (g *Game) Render() {
	g.Renderer.DrawSprite(
		resource.GetTexture("face"),
		&mgl32.Vec2{200, 200},
		&mgl32.Vec2{300, 400},
		45,
		&mgl32.Vec3{0, 1, 0},
	)
}
