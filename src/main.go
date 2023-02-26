package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"

	"breakout/src/game"
	"breakout/src/resource"
)

const (
	ScreenWidth  = 800
	ScreenHeight = 600
	WindowTitle  = "Breakout"
)

var breakout *game.Game

func main() {
	runtime.LockOSThread()
	defer resource.Cleanup()

	breakout = game.NewGame(ScreenWidth, ScreenHeight)

	window, err := initGLFW()
	if err != nil {
		handleFatalError(fmt.Errorf("failed to init GLFW: %w", err))
	}
	defer glfw.Terminate()

	err = initOpenGL()
	if err != nil {
		handleFatalError(fmt.Errorf("failed to init OpenGL: %w", err))
	}

	breakout.Init()
	defer breakout.Cleanup()

	var deltaTime, lastTime float64

	for !window.ShouldClose() {
		currTime := glfw.GetTime()
		deltaTime = currTime - lastTime
		lastTime = currTime

		glfw.PollEvents()

		breakout.ProcessInput(deltaTime)

		breakout.Update(deltaTime)

		gl.ClearColor(0, 0, 0, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		breakout.Render()

		window.SwapBuffers()
	}
}

func handleFatalError(err error) {
	log.Fatal(err)
}

func initGLFW() (*glfw.Window, error) {
	err := glfw.Init()
	if err != nil {
		panic(err)
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(ScreenWidth, ScreenHeight, WindowTitle, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create a window: %w", err)
	}

	window.MakeContextCurrent()

	window.SetKeyCallback(keyCallback)
	window.SetFramebufferSizeCallback(framebufferSizeCallback)

	return window, nil
}

func framebufferSizeCallback(_ *glfw.Window, width, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		window.SetShouldClose(true)
	}

	if key >= 0 && key < 1024 {
		if action == glfw.Press {
			breakout.Keys[key] = true
		} else if action == glfw.Release {
			breakout.Keys[key] = false
		}
	}
}

func initOpenGL() error {
	if err := gl.Init(); err != nil {
		return fmt.Errorf("failed to init OpenGL: %w", err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	return nil
}
