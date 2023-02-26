package game

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
}

func NewGame(width, height int) *Game {
	return &Game{
		State:  GameActive,
		Width:  width,
		Height: height,
	}
}

func (g *Game) Cleanup() {

}

func (g *Game) Init() {

}

func (g *Game) ProcessInput(dt float64) {

}

func (g *Game) Update(dt float64) {

}

func (g *Game) Render() {

}
