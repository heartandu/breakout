package render

import (
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"

	"breakout/src/shader"
	"breakout/src/texture"
)

type SpriteRenderer struct {
	s       *shader.Shader
	quadVAO uint32
}

func NewSpriteRenderer(s *shader.Shader) *SpriteRenderer {
	renderer := &SpriteRenderer{s: s}
	renderer.initRenderData()

	return renderer
}

func (s *SpriteRenderer) Cleanup() {
	gl.DeleteVertexArrays(1, &s.quadVAO)
}

func (s *SpriteRenderer) DrawSprite(
	t *texture.Texture2D,
	position, size *mgl32.Vec2,
	rotate float32,
	color *mgl32.Vec3,
) {
	s.s.Use()
	model := mgl32.Translate3D(position.X(), position.Y(), 0)

	model = model.Mul4(mgl32.Translate3D(0.5*size.X(), 0.5*size.Y(), 0))
	model = model.Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(rotate)))
	model = model.Mul4(mgl32.Translate3D(-0.5*size.X(), -0.5*size.Y(), 0))

	model = model.Mul4(mgl32.Scale3D(size.X(), size.Y(), 1))

	s.s.SetMatrix4("model", &model, false)

	s.s.SetVector3fv("spriteColor", color, false)

	gl.ActiveTexture(gl.TEXTURE0)
	t.Bind()

	gl.BindVertexArray(s.quadVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)
}

func (s *SpriteRenderer) initRenderData() {
	var (
		vertices = []float32{
			0, 1, 0, 1,
			1, 0, 1, 0,
			0, 0, 0, 0,

			0, 1, 0, 1,
			1, 1, 1, 1,
			1, 0, 1, 0,
		}

		vbo   uint32
		fType float32
		fSize = int(unsafe.Sizeof(fType))
	)

	gl.GenVertexArrays(1, &s.quadVAO)
	gl.GenBuffers(1, &vbo)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(
		gl.ARRAY_BUFFER,
		len(vertices)*fSize,
		gl.Ptr(&vertices[0]),
		gl.STATIC_DRAW,
	)

	gl.BindVertexArray(s.quadVAO)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, int32(4*fSize), nil)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}
