package game

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-gl/mathgl/mgl32"

	"breakout/src/render"
	"breakout/src/resource"
)

type Level struct {
	Bricks []*Object
}

func (g *Level) Load(fileName string, levelWidth, levelHeight int) error {
	g.Bricks = make([]*Object, 0)

	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("failed to open level file: %w", err)
	}
	defer file.Close()

	tileData := make([][]int, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tileCodes := strings.Split(strings.TrimSpace(scanner.Text()), " ")
		row := make([]int, len(tileCodes))

		for i := range tileCodes {
			tileCode, err := strconv.Atoi(tileCodes[i])
			if err != nil {
				return fmt.Errorf("failed to parse tile code: %w", err)
			}

			row[i] = tileCode
		}

		tileData = append(tileData, row)
	}

	if len(tileData) > 0 {
		g.init(tileData, levelWidth, levelHeight)
	}

	return nil
}

func (g *Level) Draw(renderer *render.SpriteRenderer) {
	for _, brick := range g.Bricks {
		if !brick.Destroyed {
			brick.Draw(renderer)
		}
	}
}

func (g *Level) IsCompleted() bool {
	for _, brick := range g.Bricks {
		if !brick.IsSolid && !brick.Destroyed {
			return false
		}
	}

	return true
}

func (g *Level) init(tileData [][]int, levelWidth, levelHeight int) {
	height := len(tileData)
	width := len(tileData[0])
	unitWidth := float32(levelWidth) / float32(width)
	unitHeight := float32(levelHeight) / float32(height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if tileData[y][x] == 1 {
				brickObj := NewObject(
					mgl32.Vec2{unitWidth * float32(x), unitHeight * float32(y)},
					mgl32.Vec2{unitWidth, unitHeight},
					resource.GetTexture("block_solid"),
					&mgl32.Vec3{0.8, 0.8, 0.7},
					nil,
				)
				brickObj.IsSolid = true

				g.Bricks = append(g.Bricks, brickObj)
			} else if tileData[y][x] > 1 {
				color := mgl32.Vec3{1, 1, 1}
				switch tileData[y][x] {
				case 2:
					color = mgl32.Vec3{0.2, 0.6, 1}
				case 3:
					color = mgl32.Vec3{0, 0.7, 0}
				case 4:
					color = mgl32.Vec3{0.8, 0.8, 0.4}
				case 5:
					color = mgl32.Vec3{1, 0.5, 0}
				}

				brickObj := NewObject(
					mgl32.Vec2{unitWidth * float32(x), unitHeight * float32(y)},
					mgl32.Vec2{unitWidth, unitHeight},
					resource.GetTexture("block"),
					&color,
					nil,
				)

				g.Bricks = append(g.Bricks, brickObj)
			}
		}
	}
}
