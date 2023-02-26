package resource

import (
	"fmt"
	"io"
	"os"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/nicholasblaskey/stbi"

	"breakout/src/shader"
	"breakout/src/texture"
)

var (
	shaders  = make(map[string]*shader.Shader)
	textures = make(map[string]*texture.Texture2D)
)

func LoadShader(vShaderFileName, fShaderFileName, gShaderFileName, name string) (*shader.Shader, error) {
	var err error

	shaders[name], err = loadShaderFromFile(vShaderFileName, fShaderFileName, gShaderFileName)
	if err != nil {
		return nil, fmt.Errorf("faled to load shader from file: %w", err)
	}

	return shaders[name], nil
}

func GetShader(name string) *shader.Shader {
	return shaders[name]
}

func LoadTexture(fileName string, alpha bool, name string) (*texture.Texture2D, error) {
	var err error

	textures[name], err = loadTextureFromFile(fileName, alpha)
	if err != nil {
		return nil, fmt.Errorf("failed to load texture from file: %w", err)
	}

	return textures[name], nil
}

func GetTexture(name string) *texture.Texture2D {
	return textures[name]
}

func Cleanup() {
	for name := range shaders {
		gl.DeleteProgram(shaders[name].ID)
	}

	for name := range textures {
		gl.DeleteTextures(1, &textures[name].ID)
	}
}

func loadShaderFromFile(vShaderFileName, fShaderFileName, gShaderFileName string) (*shader.Shader, error) {
	if vShaderFileName == "" || fShaderFileName == "" {
		return nil, fmt.Errorf("vertex shader and fragment shader file names should not be empty")
	}

	var (
		vertexSource, fragmentSource, geometrySource []byte
		err                                          error
	)

	vertexSource, err = loadShaderSourceFromFile(vShaderFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to load vertex shader source: %w", err)
	}

	fragmentSource, err = loadShaderSourceFromFile(fShaderFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to load fragment shader source: %w", err)
	}

	if gShaderFileName != "" {
		geometrySource, err = loadShaderSourceFromFile(gShaderFileName)
		if err != nil {
			return nil, fmt.Errorf("failed to load geometry shader source: %w", err)
		}
	}

	s := shader.NewShader()
	err = s.Compile(string(vertexSource), string(fragmentSource), string(geometrySource))
	if err != nil {
		return nil, fmt.Errorf("failed to compile shader: %w", err)
	}

	return s, nil
}

func loadShaderSourceFromFile(fileName string) ([]byte, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	source, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read source: %w", err)
	}

	return source, nil
}

func loadTextureFromFile(fileName string, alpha bool) (*texture.Texture2D, error) {
	t := texture.NewTexture2D()
	if alpha {
		t.InternalFormat = gl.RGBA
		t.ImageFormat = gl.RGBA
	}

	data, width, height, _, cleanup, err := stbi.Load(fileName, true, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}
	defer cleanup()

	t.Generate(width, height, data)

	return t, nil
}
