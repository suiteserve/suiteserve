package seed

import (
	"bytes"
	"github.com/tmazeika/testpass/repo"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"strings"
)

var attachments = []struct {
	a     repo.UnsavedAttachmentInfo
	srcFn func() io.Reader
}{
	{
		a: repo.UnsavedAttachmentInfo{
			Filename:    "old.xml",
			ContentType: "application/xml",
		},
		srcFn: func() io.Reader {
			return strings.NewReader(
				`<parent><child a="b">Hello, world!</child></parent>`)
		},
	},
	{
		a: repo.UnsavedAttachmentInfo{
			Filename:    "plain.txt",
			ContentType: "text/plain; charset=utf-8",
		},
		srcFn: func() io.Reader {
			return strings.NewReader("Hello, world!")
		},
	},
	{
		a: repo.UnsavedAttachmentInfo{
			Filename:    "simple.json",
			ContentType: "application/json",
		},
		srcFn: func() io.Reader {
			return strings.NewReader(`{"hello": "world", "abc": [1, 2, 3]}`)
		},
	},
	{
		a: repo.UnsavedAttachmentInfo{
			Filename:    "color.png",
			ContentType: "image/png",
		},
		srcFn: func() io.Reader {
			const width, height = 256, 256
			img := image.NewNRGBA(image.Rect(0, 0, width, height))
			for y := 0; y < height; y++ {
				for x := 0; x < width; x++ {
					img.Set(x, y, color.NRGBA{
						R: uint8((x + y) & 255),
						G: uint8((x + y) << 1 & 255),
						B: uint8((x + y) << 2 & 255),
						A: 255,
					})
				}
			}

			var b bytes.Buffer
			if err := png.Encode(&b, img); err != nil {
				log.Fatalln(err)
			}
			return &b
		},
	},
}
