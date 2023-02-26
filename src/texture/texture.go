package texture

import (
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
)

type Texture2D struct {
	ID             uint32
	Width          int32
	Height         int32
	InternalFormat int32
	ImageFormat    uint32
	WrapS          int32
	WrapT          int32
	FilterMin      int32
	FilterMax      int32
}

func NewTexture2D() *Texture2D {
	var id uint32
	gl.GenTextures(1, &id)

	return &Texture2D{
		ID:             id,
		InternalFormat: gl.RGB,
		ImageFormat:    gl.RGB,
		WrapS:          gl.REPEAT,
		WrapT:          gl.REPEAT,
		FilterMin:      gl.LINEAR,
		FilterMax:      gl.LINEAR,
	}
}

func (t *Texture2D) Generate(width, height int32, data unsafe.Pointer) {
	t.Width = width
	t.Height = height

	gl.BindTexture(gl.TEXTURE_2D, t.ID)
	gl.TexImage2D(gl.TEXTURE_2D,
		0,
		t.InternalFormat,
		t.Width,
		t.Height,
		0,
		t.ImageFormat,
		gl.UNSIGNED_BYTE,
		data,
	)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, t.WrapS)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, t.WrapT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, t.FilterMin)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, t.FilterMax)

	gl.BindTexture(gl.TEXTURE_2D, 0)
}

func (t *Texture2D) Bind() {
	gl.BindTexture(gl.TEXTURE_2D, t.ID)
}
