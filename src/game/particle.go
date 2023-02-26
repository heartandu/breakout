package game

import (
	"math/rand"
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"

	"breakout/src/shader"
	"breakout/src/texture"
)

type Particle struct {
	Position mgl32.Vec2
	Velocity mgl32.Vec2
	Color    mgl32.Vec4
	Life     float32
}

func NewParticle() Particle {
	return Particle{
		Position: mgl32.Vec2{0, 0},
		Velocity: mgl32.Vec2{0, 0},
		Color:    mgl32.Vec4{1, 1, 1, 1},
		Life:     0,
	}
}

type ParticleGenerator struct {
	particles        []Particle
	amount           int
	lastUsedParticle int

	s   *shader.Shader
	t   *texture.Texture2D
	vao uint32
}

func NewParticleGenerator(s *shader.Shader, t *texture.Texture2D, amount int) *ParticleGenerator {
	p := &ParticleGenerator{amount: amount, s: s, t: t}
	p.init()

	return p
}

func (pg *ParticleGenerator) init() {
	var (
		vbo      uint32
		fType    float32
		fSize    = int(unsafe.Sizeof(fType))
		vertices = []float32{
			0, 1, 0, 1,
			1, 0, 1, 0,
			0, 0, 0, 0,

			0, 1, 0, 1,
			1, 1, 1, 1,
			1, 0, 1, 0,
		}
	)

	gl.GenVertexArrays(1, &pg.vao)
	gl.GenBuffers(1, &vbo)
	gl.BindVertexArray(pg.vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(
		gl.ARRAY_BUFFER,
		len(vertices)*fSize,
		gl.Ptr(&vertices[0]),
		gl.STATIC_DRAW,
	)

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, int32(4*fSize), nil)
	gl.BindVertexArray(0)

	pg.particles = make([]Particle, 0, pg.amount)
	for i := 0; i < pg.amount; i++ {
		pg.particles = append(pg.particles, NewParticle())
	}
}

func (pg *ParticleGenerator) Update(dt float64, o *Object, newParticles int, offset mgl32.Vec2) {
	for i := 0; i < newParticles; i++ {
		unusedParticle := pg.firstUnusedParticle()
		pg.respawnParticle(&pg.particles[unusedParticle], o, offset)
	}

	for i := range pg.particles {
		p := &pg.particles[i]
		p.Life -= float32(dt)

		if p.Life > 0 {
			p.Position.Sub(p.Velocity.Mul(float32(dt)))
			p.Color[3] -= float32(dt) * 2.5
		}
	}
}

func (pg *ParticleGenerator) firstUnusedParticle() int {
	for i := pg.lastUsedParticle; i < pg.amount; i++ {
		if pg.particles[i].Life <= 0 {
			pg.lastUsedParticle = i
			return i
		}
	}

	for i := 0; i < pg.lastUsedParticle; i++ {
		if pg.particles[i].Life <= 0 {
			pg.lastUsedParticle = i
			return i
		}
	}

	pg.lastUsedParticle = 0
	return 0
}

func (pg *ParticleGenerator) respawnParticle(p *Particle, o *Object, offset mgl32.Vec2) {
	random := float32((rand.Int()%100)-50) / 10
	rColor := 0.5 + float32(rand.Int()%100)/100
	p.Position = o.Position.Add(mgl32.Vec2{random, random}).Add(offset)
	p.Color = mgl32.Vec4{rColor, rColor, rColor, 1}
	p.Life = 1
	p.Velocity = o.Velocity.Mul(0.1)
}

func (pg *ParticleGenerator) Draw() {
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
	pg.s.Use()

	for i := range pg.particles {
		if pg.particles[i].Life > 0 {
			pg.s.SetVector2fv("offset", &pg.particles[i].Position, false)
			pg.s.SetVector4fv("color", &pg.particles[i].Color, false)
			pg.t.Bind()
			gl.BindVertexArray(pg.vao)
			gl.DrawArrays(gl.TRIANGLES, 0, 6)
			gl.BindVertexArray(0)
		}
	}

	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}
