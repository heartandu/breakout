package shader

import (
	"fmt"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Shader struct {
	ID uint32
}

func NewShader() *Shader {
	return &Shader{}
}

func (s *Shader) Use() *Shader {
	gl.UseProgram(s.ID)
	return s
}

func (s *Shader) Compile(vertexSource, fragmentSource, geometrySource string) error {
	shaders := make([]uint32, 0, 3)

	sVertex, err := compileShader(vertexSource, gl.VERTEX_SHADER)
	if err != nil {
		return fmt.Errorf("failed to compile vertex shader: %w", err)
	}

	shaders = append(shaders, sVertex)

	sFragment, err := compileShader(fragmentSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return fmt.Errorf("failed to compile fragment shader: %w", err)
	}

	shaders = append(shaders, sFragment)

	if geometrySource != "" {
		sGeometry, err := compileShader(geometrySource, gl.GEOMETRY_SHADER)
		if err != nil {
			return fmt.Errorf("failed to compile geometry shader: %w", err)
		}

		shaders = append(shaders, sGeometry)
	}

	s.ID, err = createAndLinkProgram(shaders...)
	if err != nil {
		return fmt.Errorf("failed to create and link shader program: %w", err)
	}

	for _, s := range shaders {
		gl.DeleteShader(s)
	}

	return nil
}

func (s *Shader) SetFloat(name string, value float32, useShader bool) {
	if useShader {
		s.Use()
	}

	gl.Uniform1f(s.uniformLocation(name), value)
}

func (s *Shader) SetInteger(name string, value int, useShader bool) {
	if useShader {
		s.Use()
	}

	gl.Uniform1i(s.uniformLocation(name), int32(value))
}

func (s *Shader) SetVector2f(name string, x, y float32, useShader bool) {
	if useShader {
		s.Use()
	}

	gl.Uniform2f(s.uniformLocation(name), x, y)
}

func (s *Shader) SetVector2fv(name string, value *mgl32.Vec2, useShader bool) {
	if useShader {
		s.Use()
	}

	gl.Uniform2f(s.uniformLocation(name), value.X(), value.Y())
}

func (s *Shader) SetVector3f(name string, x, y, z float32, useShader bool) {
	if useShader {
		s.Use()
	}

	gl.Uniform3f(s.uniformLocation(name), x, y, z)
}

func (s *Shader) SetVector3fv(name string, value *mgl32.Vec3, useShader bool) {
	if useShader {
		s.Use()
	}

	gl.Uniform3f(s.uniformLocation(name), value.X(), value.Y(), value.Z())
}

func (s *Shader) SetVector4f(name string, x, y, z, w float32, useShader bool) {
	if useShader {
		s.Use()
	}

	gl.Uniform4f(s.uniformLocation(name), x, y, z, w)
}

func (s *Shader) SetVector4fv(name string, value *mgl32.Vec4, useShader bool) {
	if useShader {
		s.Use()
	}

	gl.Uniform4f(s.uniformLocation(name), value.X(), value.Y(), value.Z(), value.W())
}

func (s *Shader) SetMatrix4(name string, matrix *mgl32.Mat4, useShader bool) {
	if useShader {
		s.Use()
	}

	gl.UniformMatrix4fv(s.uniformLocation(name), 1, false, &matrix[0])
}

func (s *Shader) uniformLocation(name string) int32 {
	return gl.GetUniformLocation(s.ID, gl.Str(name+"\x00"))
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	cSource, free := gl.Strs(source + "\x00")
	gl.ShaderSource(shader, 1, cSource, nil)
	free()
	gl.CompileShader(shader)

	if err := checkCompileError(shader, shaderCompileType); err != nil {
		return 0, err
	}

	return shader, nil
}

func createAndLinkProgram(shaders ...uint32) (uint32, error) {
	prog := gl.CreateProgram()

	for i := range shaders {
		gl.AttachShader(prog, shaders[i])
	}

	gl.LinkProgram(prog)

	if err := checkCompileError(prog, programCompileType); err != nil {
		return 0, err
	}

	return prog, nil
}

type compileType int

const (
	shaderCompileType compileType = iota
	programCompileType

	logLength = 1024
)

func checkCompileError(ID uint32, t compileType) error {
	var (
		success    int32
		logMessage = string(make([]byte, logLength))
	)

	switch t {
	case shaderCompileType:
		gl.GetShaderiv(ID, gl.COMPILE_STATUS, &success)
		if success == gl.FALSE {
			gl.GetShaderInfoLog(ID, logLength, nil, gl.Str(logMessage))

			return fmt.Errorf("failed to compile shader: %v", logMessage)
		}
	case programCompileType:
		gl.GetProgramiv(ID, gl.LINK_STATUS, &success)
		if success == gl.FALSE {
			gl.GetProgramInfoLog(ID, logLength, nil, gl.Str(logMessage))

			return fmt.Errorf("failed to link program: %v", logMessage)
		}
	}

	return nil
}
