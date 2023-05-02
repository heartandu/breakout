package render

import (
	"fmt"
	"image"
	"image/draw"
	"os"
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"breakout/src/resource"
	"breakout/src/shader"
	"breakout/src/types"
)

const (
	vShaderFileName = "resources/shaders/text2d.vert"
	fShaderFileName = "resources/shaders/text2d.frag"
	shaderName      = "text"
	dpi             = 72
	baselineXOffset = 10
)

type TextRenderer struct {
	chars map[rune]*character
	s     *shader.Shader

	vao uint32
	vbo uint32
}

func NewTextRenderer(width, height int) (*TextRenderer, error) {
	err := resource.LoadShader(vShaderFileName, fShaderFileName, "", shaderName)
	if err != nil {
		return nil, fmt.Errorf("failed to load shader: %w", err)
	}

	s := resource.GetShader(shaderName)
	projection := mgl32.Ortho2D(0, float32(width), float32(height), 0)
	s.SetMatrix4("projection", &projection, true)
	s.SetInteger("text", 0, false)

	r := &TextRenderer{
		chars: make(map[rune]*character, 0),
		s:     s,
	}

	r.initBuffers()

	return r, nil
}

func (r *TextRenderer) initBuffers() {
	var (
		fType float32
		fSize = int(unsafe.Sizeof(fType))
	)

	gl.GenVertexArrays(1, &r.vao)
	gl.GenBuffers(1, &r.vbo)
	gl.BindVertexArray(r.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, fSize*6*4, nil, gl.DYNAMIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, int32(4*fSize), nil)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

func (r *TextRenderer) Load(fontPath string, fontSize int) error {
	r.chars = make(map[rune]*character, 0)

	ttf, err := loadFont(fontPath)
	if err != nil {
		return fmt.Errorf("failed to load font: %w", err)
	}

	fg, bg := image.White, image.Transparent

	c := freetype.NewContext()
	c.SetDPI(dpi)
	c.SetFont(ttf)
	c.SetFontSize(float64(fontSize))
	c.SetSrc(fg)
	c.SetHinting(font.HintingNone)

	for i := rune(0); i < 128; i++ {
		char, err := newCharacter(ttf, c, bg, fontSize, i)
		if err != nil {
			return fmt.Errorf("failed to create character: %w", err)
		}

		r.chars[i] = char
	}

	return nil
}

func (r *TextRenderer) RenderText(text string, x, y, scale float32, color *mgl32.Vec3) {
	r.s.Use()
	r.s.SetVector3fv("textColor", color, false)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindVertexArray(r.vao)

	for _, c := range text {
		char := r.chars[c]

		xpos := x + float32(char.bearing.X())*scale
		ypos := y
		//ypos := y + float32(char.size.Y()-char.bearing.Y())*scale

		w := float32(char.size.X()) * scale
		h := float32(char.size.Y()) * scale

		vertices := [6][4]float32{
			{xpos, ypos + h, 0, 1},
			{xpos + w, ypos, 1, 0},
			{xpos, ypos, 0, 0},

			{xpos, ypos + h, 0, 1},
			{xpos + w, ypos + h, 1, 1},
			{xpos + w, ypos, 1, 0},
		}

		gl.BindTexture(gl.TEXTURE_2D, char.textureID)

		gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
		gl.BufferSubData(gl.ARRAY_BUFFER, 0, int(unsafe.Sizeof(vertices)), gl.Ptr(&vertices[0][0]))
		gl.BindBuffer(gl.ARRAY_BUFFER, 0)

		gl.DrawArrays(gl.TRIANGLES, 0, 6)

		x += float32(char.advance) * scale
	}

	gl.BindVertexArray(0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

func loadFont(fontPath string) (*truetype.Font, error) {
	bytes, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read font file: %w", err)
	}

	ttf, err := truetype.Parse(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font file: %w", err)
	}

	return ttf, nil
}

type character struct {
	textureID uint32
	size      types.IVec2
	bearing   types.IVec2
	advance   int
}

func newCharacter(ttf *truetype.Font, c *freetype.Context, bg *image.Uniform, fontSize int, r rune) (*character, error) {
	glyphBounds := ttf.Bounds(fixed.Int26_6((fontSize)))
	glyphWidth := int(glyphBounds.Max.X - glyphBounds.Min.X)
	glyphHeight := int(glyphBounds.Max.Y - glyphBounds.Min.Y)

	img := image.NewRGBA(image.Rect(0, 0, glyphWidth, glyphHeight))
	draw.Draw(img, img.Bounds(), bg, image.Point{}, draw.Src)

	c.SetClip(img.Bounds())
	c.SetDst(img)

	baselineX, baselineY := baselineXOffset, int(c.PointToFixed(float64(fontSize))>>6)

	pt := freetype.Pt(baselineX, baselineY)

	_, err := c.DrawString(string(r), pt)
	if err != nil {
		return nil, fmt.Errorf("failed to draw character: %w", err)
	}

	return &character{
		textureID: newCharacterTexture(img),
		size:      types.IVec2{img.Bounds().Dx(), img.Bounds().Dy()},
		bearing:   types.IVec2{-baselineX, baselineY},
		advance:   int(ttf.HMetric(fixed.Int26_6(fontSize), ttf.Index(r)).AdvanceWidth),
	}, nil
}

func newCharacterTexture(img *image.RGBA) uint32 {
	var id uint32
	gl.GenTextures(1, &id)
	gl.BindTexture(gl.TEXTURE_2D, id)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(img.Bounds().Dx()),
		int32(img.Bounds().Dy()),
		0,
		gl.RGBA,
		gl.UNSIGNED_INT,
		gl.Ptr(img.Pix),
	)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	return id
}
