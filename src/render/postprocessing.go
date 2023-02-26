package render

import (
	"fmt"
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"

	"breakout/src/shader"
	"breakout/src/texture"
)

type PostProcessor struct {
	s *shader.Shader
	t *texture.Texture2D

	Width  int32
	Height int32

	Confuse bool
	Chaos   bool
	Shake   bool

	msfbo uint32
	fbo   uint32
	rbo   uint32
	vao   uint32
}

func NewPostProcessor(s *shader.Shader, width int, height int) (*PostProcessor, error) {
	p := &PostProcessor{s: s, t: texture.NewTexture2D(), Width: int32(width), Height: int32(height)}
	err := p.init()
	if err != nil {
		return nil, fmt.Errorf("failed to init post processor: %w", err)
	}

	return p, nil
}

func (p *PostProcessor) init() error {
	err := p.initFrameBuffers()
	if err != nil {
		return fmt.Errorf("failed to init framebuffers: %w", err)
	}

	p.initRenderData()
	p.initUniforms()

	return nil
}

func (p *PostProcessor) initFrameBuffers() error {
	gl.GenFramebuffers(1, &p.msfbo)
	gl.GenFramebuffers(1, &p.fbo)
	gl.GenRenderbuffers(1, &p.rbo)

	gl.BindFramebuffer(gl.FRAMEBUFFER, p.msfbo)
	gl.BindRenderbuffer(gl.RENDERBUFFER, p.rbo)
	gl.RenderbufferStorageMultisample(gl.RENDERBUFFER, 4, gl.RGB, p.Width, p.Height)
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.RENDERBUFFER, p.rbo)

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		return fmt.Errorf("failed to initialize MSFBO")
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, p.fbo)
	p.t.Generate(p.Width, p.Height, nil)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, p.t.ID, 0)

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		return fmt.Errorf("failed to initialize FBO")
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	return nil
}

func (p *PostProcessor) initRenderData() {
	var (
		vbo      uint32
		fType    float32
		fSize    = int(unsafe.Sizeof(fType))
		vertices = []float32{
			-1, -1, 0, 0,
			1, 1, 1, 1,
			-1, 1, 0, 1,

			-1, -1, 0, 0,
			1, -1, 1, 0,
			1, 1, 1, 1,
		}
	)

	gl.GenVertexArrays(1, &p.vao)
	gl.GenBuffers(1, &vbo)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(
		gl.ARRAY_BUFFER,
		len(vertices)*fSize,
		gl.Ptr(&vertices[0]),
		gl.STATIC_DRAW,
	)

	gl.BindVertexArray(p.vao)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, int32(4*fSize), nil)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

func (p *PostProcessor) initUniforms() {
	p.s.SetInteger("scene", 0, true)

	var offset float32 = 1.0 / 300
	offsets := [9][2]float32{
		{-offset, offset},
		{0, offset},
		{offset, offset},
		{-offset, 0},
		{0, 0},
		{offset, 0},
		{-offset, -offset},
		{0, -offset},
		{offset, -offset},
	}
	p.s.SetAnything("offsets", false, func(locationID int32) {
		gl.Uniform2fv(locationID, 9, &offsets[0][0])
	})

	edgeKernel := []int32{
		-1, -1, -1,
		-1, 8, -1,
		-1, -1, -1,
	}
	p.s.SetAnything("edge_kernel", false, func(locationID int32) {
		gl.Uniform1iv(locationID, 9, &edgeKernel[0])
	})

	blurKernel := []float32{
		1.0 / 16, 2.0 / 16, 1.0 / 16,
		2.0 / 16, 4.0 / 16, 2.0 / 16,
		1.0 / 16, 2.0 / 16, 1.0 / 16,
	}
	p.s.SetAnything("blur_kernel", false, func(locationID int32) {
		gl.Uniform1fv(locationID, 9, &blurKernel[0])
	})
}

func (p *PostProcessor) BeginRender() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, p.msfbo)
	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)
}

func (p *PostProcessor) EndRender() {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, p.msfbo)
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, p.fbo)
	gl.BlitFramebuffer(
		0,
		0,
		p.Width,
		p.Height,
		0,
		0,
		p.Width,
		p.Height,
		gl.COLOR_BUFFER_BIT,
		gl.NEAREST,
	)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

func (p *PostProcessor) Render(time float64) {
	p.s.Use()
	p.s.SetFloat("time", float32(time), false)
	p.s.SetInteger("confuse", boolToInt(p.Confuse), false)
	p.s.SetInteger("chaos", boolToInt(p.Chaos), false)
	p.s.SetInteger("shake", boolToInt(p.Shake), false)

	gl.ActiveTexture(gl.TEXTURE0)
	p.t.Bind()
	gl.BindVertexArray(p.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)
}

func boolToInt(val bool) int {
	if val {
		return 1
	}

	return 0
}
