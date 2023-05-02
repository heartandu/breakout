package game

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"breakout/src/render"
	"breakout/src/resource"
	"breakout/src/sound"
)

type State int

const (
	StateActive State = iota
	StateMenu
	StateWin

	playerVelocity float32 = 500
	ballRadius     float32 = 12.5
	particleAmount         = 2000

	startingLives = 3
)

var (
	playerSize       = mgl32.Vec2{100, 20}
	initBallVelocity = mgl32.Vec2{100, -350}
	shaderFiles      = map[string]struct {
		v, f, g string
	}{
		"sprite":         {"resources/shaders/sprite.vert", "resources/shaders/sprite.frag", ""},
		"particle":       {"resources/shaders/particle.vert", "resources/shaders/particle.frag", ""},
		"postprocessing": {"resources/shaders/postprocessing.vert", "resources/shaders/postprocessing.frag", ""},
	}
	textureFiles = map[string]struct {
		path  string
		alpha bool
	}{
		"background":          {"resources/textures/background.png", false},
		"face":                {"resources/textures/happy.png", true},
		"block":               {"resources/textures/block.png", false},
		"block_solid":         {"resources/textures/block_solid.png", false},
		"paddle":              {"resources/textures/paddle.png", true},
		"particle":            {"resources/textures/particle.png", true},
		"powerup_chaos":       {"resources/textures/powerup_chaos.png", true},
		"powerup_confuse":     {"resources/textures/powerup_confuse.png", true},
		"powerup_increase":    {"resources/textures/powerup_increase.png", true},
		"powerup_passthrough": {"resources/textures/powerup_passthrough.png", true},
		"powerup_speed":       {"resources/textures/powerup_speed.png", true},
		"powerup_sticky":      {"resources/textures/powerup_sticky.png", true},
	}
	levelFiles = []string{
		"resources/levels/one.lvl",
		"resources/levels/two.lvl",
		"resources/levels/three.lvl",
		"resources/levels/four.lvl",
	}
	fontFiles = map[string]int{
		"resources/fonts/ocraext.ttf": 24,
	}
)

type Game struct {
	State         State
	Keys          [1024]bool
	KeysProcessed [1024]bool
	Width         int
	Height        int

	Levels []Level
	Level  int

	PowerUps []PowerUp

	Renderer  *render.SpriteRenderer
	Effects   *render.PostProcessor
	Text      *render.TextRenderer
	Particles *ParticleGenerator

	ball       *Ball
	player     *Object
	background *Object
	lives      uint32

	soundsPlayer *sound.Player

	shakeTime float64
}

func NewGame(width, height int) *Game {
	return &Game{
		State:  StateMenu,
		Width:  width,
		Height: height,
	}
}

func (g *Game) Cleanup() error {
	if g.Renderer != nil {
		g.Renderer.Cleanup()
	}

	if g.soundsPlayer != nil {
		if err := g.soundsPlayer.Cleanup(); err != nil {
			return fmt.Errorf("failed to cleanup sounds player: %w", err)
		}
	}

	return nil
}

func (g *Game) Init() error {
	err := g.loadShaders()
	if err != nil {
		return fmt.Errorf("failed to load shaders: %w", err)
	}

	projection := mgl32.Ortho2D(0, float32(g.Width), float32(g.Height), 0)
	resource.GetShader("sprite").SetInteger("image", 0, true)
	resource.GetShader("sprite").SetMatrix4("projection", &projection, false)
	resource.GetShader("particle").SetInteger("sprite", 0, true)
	resource.GetShader("particle").SetMatrix4("projection", &projection, false)

	g.Renderer = render.NewSpriteRenderer(resource.GetShader("sprite"))
	g.Effects, err = render.NewPostProcessor(resource.GetShader("postprocessing"), g.Width, g.Height)
	if err != nil {
		return fmt.Errorf("failed to create post processor: %w", err)
	}

	g.Text, err = render.NewTextRenderer(g.Width, g.Height)
	if err != nil {
		return fmt.Errorf("failed to create text renderer: %w", err)
	}

	for path, size := range fontFiles {
		if err := g.Text.Load(path, size); err != nil {
			return fmt.Errorf("failed to load font: %w", err)
		}
	}

	err = g.loadTextures()
	if err != nil {
		return fmt.Errorf("failed to load textures: %w", err)
	}

	err = g.loadLevels()
	if err != nil {
		return fmt.Errorf("failed to load levels: %w", err)
	}

	g.PowerUps = make([]PowerUp, 0)

	g.Particles = NewParticleGenerator(
		resource.GetShader("particle"),
		resource.GetTexture("particle"),
		particleAmount,
	)

	g.background = NewObject(
		mgl32.Vec2{0, 0},
		mgl32.Vec2{float32(g.Width), float32(g.Height)},
		resource.GetTexture("background"),
		nil,
		nil,
	)

	g.player = NewObject(
		mgl32.Vec2{float32(g.Width)/2 - playerSize.X()/2, float32(g.Height) - playerSize.Y()},
		playerSize,
		resource.GetTexture("paddle"),
		nil,
		nil,
	)

	g.ball = NewBall(
		g.player.Position.Add(mgl32.Vec2{playerSize.X()/2 - ballRadius, -ballRadius * 2}),
		ballRadius,
		initBallVelocity,
		resource.GetTexture("face"),
	)

	g.soundsPlayer, err = sound.NewPlayer()
	if err != nil {
		return fmt.Errorf("failed to create sounds player: %w", err)
	}

	g.soundsPlayer.PlayBgMusic()

	g.lives = startingLives

	return nil
}

func (g *Game) ProcessInput(dt float64) {
	if g.State == StateActive {
		velocity := playerVelocity * float32(dt)

		if g.Keys[glfw.KeyA] {
			if g.player.Position.X() >= 0 {
				g.player.Position[0] -= velocity

				if g.ball.Stuck {
					g.ball.Position[0] -= velocity
				}
			}
		}

		if g.Keys[glfw.KeyD] {
			if g.player.Position.X() <= float32(g.Width)-g.player.Size.X() {
				g.player.Position[0] += velocity

				if g.ball.Stuck {
					g.ball.Position[0] += velocity
				}
			}
		}

		if g.Keys[glfw.KeySpace] {
			g.ball.Stuck = false
		}
	}

	if g.State == StateMenu {
		if g.Keys[glfw.KeyEnter] && !g.KeysProcessed[glfw.KeyEnter] {
			g.State = StateActive
			g.KeysProcessed[glfw.KeyEnter] = true
		}
		if g.Keys[glfw.KeyW] && !g.KeysProcessed[glfw.KeyW] {
			g.Level = (g.Level + 1) % len(g.Levels)
			g.KeysProcessed[glfw.KeyW] = true
		}
		if g.Keys[glfw.KeyS] && !g.KeysProcessed[glfw.KeyS] {
			if g.Level > 0 {
				g.Level--
			} else {
				g.Level = len(g.Levels) - 1
			}
			g.KeysProcessed[glfw.KeyS] = true
		}
	}

	if g.State == StateWin {
		if g.Keys[glfw.KeyEnter] {
			g.KeysProcessed[glfw.KeyEnter] = true
			g.Effects.Chaos = false
			g.State = StateMenu
		}
	}
}

func (g *Game) Render() {
	if g.State == StateActive || g.State == StateMenu || g.State == StateWin {
		g.Effects.BeginRender()

		{
			g.background.Draw(g.Renderer)

			g.Levels[g.Level].Draw(g.Renderer)

			g.player.Draw(g.Renderer)

			for i := range g.PowerUps {
				if !g.PowerUps[i].Destroyed {
					g.PowerUps[i].Draw(g.Renderer)
				}
			}

			g.Particles.Draw()
			g.ball.Draw(g.Renderer)
		}

		g.Effects.EndRender()
		g.Effects.Render(glfw.GetTime())

		g.Text.RenderText(fmt.Sprintf("Lives: %d", g.lives), 5, 5, 1, &mgl32.Vec3{1, 1, 1})
	}

	if g.State == StateMenu {
		g.Text.RenderText("Press ENTER to start", 250, float32(g.Height)/2, 1, &mgl32.Vec3{1, 1, 1})
		g.Text.RenderText("Press W or S to select level", 245, float32(g.Height)/2+20, 0.75, &mgl32.Vec3{1, 1, 1})
	}

	if g.State == StateWin {
		g.Text.RenderText("You WON!!!", 320, float32(g.Height)/2-20, 1, &mgl32.Vec3{0, 1, 0})
		g.Text.RenderText("Press ENTER to retry or ESC to quit", 130, float32(g.Height)/2, 1, &mgl32.Vec3{1, 1, 0})
	}
}

func (g *Game) Update(dt float64) {
	g.ball.Move(dt, g.Width)
	g.DoCollisions()
	g.Particles.Update(dt, &g.ball.Object, 2, mgl32.Vec2{g.ball.Radius / 2, g.ball.Radius / 2})
	g.UpdatePowerUps(dt)

	isLevelCompleted := g.Levels[g.Level].IsCompleted()
	if g.State == StateActive && isLevelCompleted {
		g.ResetLevel()
		g.ResetPlayer()
		g.Effects.Chaos = true
		g.State = StateWin
	}

	if g.ball.Position.Y() >= float32(g.Height) {
		g.lives -= 1
		if g.lives == 0 {
			g.ResetLevel()
			g.State = StateMenu
		}

		g.ResetPlayer()
	}

	if g.shakeTime > 0 {
		g.shakeTime -= dt
		if g.shakeTime <= 0 {
			g.Effects.Shake = false
		}
	}
}

func (g *Game) DoCollisions() {
	for _, brick := range g.Levels[g.Level].Bricks {
		if !brick.Destroyed {
			r := CheckBallCollision(g.ball, brick)
			if r.Collided {
				if !brick.IsSolid {
					brick.Destroyed = true
					g.SpawnPowerUps(brick)
					g.soundsPlayer.PlayNonSolidBlockBleep()
				} else {
					g.shakeTime = 0.05
					g.Effects.Shake = true
					g.soundsPlayer.PlaySolidBlockBleep()
				}

				if g.ball.PassThrough && !brick.IsSolid {
					continue
				}

				if r.Dir == LeftDirection || r.Dir == RightDirection {
					g.ball.Velocity[0] = -g.ball.Velocity[0]
					penetration := g.ball.Radius - float32(math.Abs(float64(r.Diff.X())))

					if r.Dir == LeftDirection {
						g.ball.Position[0] += penetration
					} else {
						g.ball.Position[0] -= penetration
					}
				} else {
					g.ball.Velocity[1] = -g.ball.Velocity[1]
					penetration := g.ball.Radius - float32(math.Abs(float64(r.Diff.Y())))

					if r.Dir == UpDirection {
						g.ball.Position[1] -= penetration
					} else {
						g.ball.Position[1] += penetration
					}
				}
			}
		}
	}

	r := CheckBallCollision(g.ball, g.player)
	if !g.ball.Stuck && r.Collided {
		centerBoard := g.player.Position.X() + g.player.Size.X()/2
		distance := g.ball.Position.X() + g.ball.Radius - centerBoard
		percentage := distance / (g.player.Size.X() / 2)

		var strength float32 = 2
		oldVelocity := g.ball.Velocity
		g.ball.Velocity[0] = initBallVelocity.X() * percentage * strength
		g.ball.Velocity[1] = -1 * float32(math.Abs(float64(g.ball.Velocity.Y())))
		g.ball.Velocity = g.ball.Velocity.Normalize().Mul(oldVelocity.Len())
		g.ball.Stuck = g.ball.Sticky

		g.soundsPlayer.PlayPaddleBleep()
	}

	for i := range g.PowerUps {
		if !g.PowerUps[i].Destroyed {
			if g.PowerUps[i].Position.Y() >= float32(g.Height) {
				g.PowerUps[i].Destroyed = true
			}

			if CheckCollision(g.player, &g.PowerUps[i].Object) {
				g.ActivatePowerUp(&g.PowerUps[i])
				g.PowerUps[i].Destroyed = true
				g.PowerUps[i].Activated = true
				g.soundsPlayer.PlayPowerUp()
			}
		}
	}
}

func (g *Game) ResetLevel() {
	for _, brick := range g.Levels[g.Level].Bricks {
		brick.Destroyed = false
	}

	g.lives = startingLives
}

func (g *Game) ResetPlayer() {
	g.player.Position = mgl32.Vec2{float32(g.Width)/2 - playerSize.X()/2, float32(g.Height) - playerSize.Y()}
	g.ball.Reset(
		g.player.Position.Add(mgl32.Vec2{playerSize.X()/2 - ballRadius, -ballRadius * 2}),
		initBallVelocity,
	)
}

func (g *Game) SpawnPowerUps(block *Object) {
	if shouldSpawn(75) {
		g.PowerUps = append(g.PowerUps, NewPowerUp(
			"speed",
			mgl32.Vec3{0.5, 0.5, 1},
			0,
			block.Position,
			resource.GetTexture("powerup_speed"),
		))
	}

	if shouldSpawn(75) {
		g.PowerUps = append(g.PowerUps, NewPowerUp(
			"sticky",
			mgl32.Vec3{1, 0.5, 1},
			20,
			block.Position,
			resource.GetTexture("powerup_sticky"),
		))
	}

	if shouldSpawn(75) {
		g.PowerUps = append(g.PowerUps, NewPowerUp(
			"pass-through",
			mgl32.Vec3{0.5, 1, 0.5},
			10,
			block.Position,
			resource.GetTexture("powerup_passthrough"),
		))
	}

	if shouldSpawn(75) {
		g.PowerUps = append(g.PowerUps, NewPowerUp(
			"pad-size-increase",
			mgl32.Vec3{1, 0.6, 0.4},
			0,
			block.Position,
			resource.GetTexture("powerup_increase"),
		))
	}

	if shouldSpawn(15) {
		g.PowerUps = append(g.PowerUps, NewPowerUp(
			"confuse",
			mgl32.Vec3{1, 0.3, 0.3},
			15,
			block.Position,
			resource.GetTexture("powerup_confuse"),
		))
	}

	if shouldSpawn(15) {
		g.PowerUps = append(g.PowerUps, NewPowerUp(
			"chaos",
			mgl32.Vec3{0.9, 0.25, 0.25},
			15,
			block.Position,
			resource.GetTexture("powerup_chaos"),
		))
	}
}

func (g *Game) UpdatePowerUps(dt float64) {
	for i := range g.PowerUps {
		g.PowerUps[i].Position = g.PowerUps[i].Position.Add(g.PowerUps[i].Velocity.Mul(float32(dt)))

		if g.PowerUps[i].Activated {
			g.PowerUps[i].Duration -= dt

			if g.PowerUps[i].Duration <= 0 {
				g.PowerUps[i].Activated = false
				switch g.PowerUps[i].Type {
				case "sticky":
					if !g.IsOtherPowerUpActive("sticky") {
						g.ball.Sticky = false
						g.player.Color = mgl32.Vec3{1, 1, 1}
					}
				case "pass-through":
					if !g.IsOtherPowerUpActive("pass-through") {
						g.ball.PassThrough = false
						g.player.Color = mgl32.Vec3{1, 1, 1}
					}
				case "confuse":
					if !g.IsOtherPowerUpActive("confuse") {
						g.Effects.Confuse = false
					}
				case "chaos":
					if !g.IsOtherPowerUpActive("chaos") {
						g.Effects.Chaos = false
					}
				}
			}
		}
	}

	moveIndex := 0
	for i := range g.PowerUps {
		if !g.PowerUps[i].Destroyed || g.PowerUps[i].Activated {
			g.PowerUps[moveIndex] = g.PowerUps[i]
			moveIndex++
		}
	}

	g.PowerUps = g.PowerUps[:moveIndex]
}

func (g *Game) ActivatePowerUp(powerUp *PowerUp) {
	switch powerUp.Type {
	case "speed":
		g.ball.Velocity = g.ball.Velocity.Mul(1.2)
	case "sticky":
		g.ball.Sticky = true
		g.player.Color = mgl32.Vec3{1, 0.5, 1}
	case "pass-through":
		g.ball.PassThrough = true
		g.ball.Color = mgl32.Vec3{1, 0.5, 0.5}
	case "pad-size-increase":
		g.player.Size[0] += 50
	case "confuse":
		if !g.Effects.Chaos {
			g.Effects.Confuse = true
		}
	case "chaos":
		if !g.Effects.Confuse {
			g.Effects.Chaos = true
		}
	}
}

func (g *Game) IsOtherPowerUpActive(t string) bool {
	for i := range g.PowerUps {
		if g.PowerUps[i].Activated && g.PowerUps[i].Type == t {
			return true
		}
	}

	return false
}

func (g *Game) loadShaders() error {
	for name, sFile := range shaderFiles {
		err := resource.LoadShader(sFile.v, sFile.f, sFile.g, name)
		if err != nil {
			return fmt.Errorf("failed to load %s shader: %w", name, err)
		}
	}

	return nil
}

func (g *Game) loadTextures() (err error) {
	for name, tFile := range textureFiles {
		err = resource.LoadTexture(tFile.path, tFile.alpha, name)
		if err != nil {
			return fmt.Errorf("failed to load %s texture: %w", name, err)
		}
	}

	return nil
}

func (g *Game) loadLevels() (err error) {
	g.Levels = make([]Level, 0, len(levelFiles))

	for i := range levelFiles {
		var l Level

		err = l.Load(levelFiles[i], g.Width, g.Height/2)
		if err != nil {
			return fmt.Errorf("failed to load level %s: %w", levelFiles[i], err)
		}

		g.Levels = append(g.Levels, l)
	}

	g.Level = 0

	return nil
}

func shouldSpawn(chance int) bool {
	r := rand.Int() % chance
	return r == 0
}
