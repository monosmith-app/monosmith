package scaffold

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/draw"
	"image/png"
)

// placeholderIcons returns the icons the bundle list names (file name →
// bytes): a plain square, so `tauri build` runs on a new app. Replace them with
// `npx tauri icon <logo.png>`.
func placeholderIcons() map[string]string {
	p256 := squarePNG(256)
	return map[string]string{
		"32x32.png":      squarePNG(32),
		"128x128.png":    squarePNG(128),
		"128x128@2x.png": p256,
		"icon.png":       squarePNG(512),
		"icon.ico":       icoOf(p256),
		"icon.icns":      icnsOf(p256),
	}
}

func squarePNG(size int) string {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.RGBA{R: 0x3b, G: 0x6e, B: 0xd8, A: 0xff}), image.Point{}, draw.Src)
	var b bytes.Buffer
	_ = png.Encode(&b, img) // writing to memory cannot fail
	return b.String()
}

// icoOf wraps a 256x256 PNG in an ICO container (a width and height of 0 mean 256).
func icoOf(p256 string) string {
	var b bytes.Buffer
	_ = binary.Write(&b, binary.LittleEndian, []uint16{0, 1, 1})
	b.Write([]byte{0, 0, 0, 0})
	_ = binary.Write(&b, binary.LittleEndian, []uint16{1, 32})
	_ = binary.Write(&b, binary.LittleEndian, []uint32{uint32(len(p256)), 22})
	b.WriteString(p256)
	return b.String()
}

// icnsOf wraps a 256x256 PNG in an ICNS container (element type ic08).
func icnsOf(p256 string) string {
	var b bytes.Buffer
	b.WriteString("icns")
	_ = binary.Write(&b, binary.BigEndian, uint32(16+len(p256)))
	b.WriteString("ic08")
	_ = binary.Write(&b, binary.BigEndian, uint32(8+len(p256)))
	b.WriteString(p256)
	return b.String()
}
