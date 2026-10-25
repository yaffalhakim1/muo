package muo

import (
	"fmt"
	"image"

	"github.com/egoist/mygo/ui"

	"example.com/muo/muo/pdf"
)

// The page raster. MyGo's bitmap is what the view draws; the pixels come from
// the PDF interpreter on a goroutine, so a heavy page never freezes the window.

// pdfPageRaster is one page of the file being read: its size in points, the
// bitmap it currently holds, and the density that bitmap was drawn at.
type pdfPageRaster struct {
	page    int
	w, h    float32 // points
	width   int     // pixels
	height  int
	bitmap  *ui.Bitmap
	err     string
	loading bool
}

// beginPage starts a render when the page on screen needs one at the density
// the window is drawn at.
func (r *Reader) beginPage(c *ui.Context, target float32) {
	if r.file == nil || !r.file.HasPage(r.viewing) {
		return
	}
	pw, ph := r.file.PageSize(r.viewing)
	if pw <= 0 || ph <= 0 {
		return
	}
	// The page is drawn at the pixel density the window is rendered at, so a
	// 2× laptop panel gets twice the pixels of the same window on a 1× screen:
	// the text stays sharp instead of being upscaled.
	scale := float32(r.typ.EffectivePPI / 72)
	if scale < 1 {
		scale = 1
	}
	if scale > 3 {
		scale = 3
	}
	width := int(target * scale)
	if width < 200 {
		width = 200
	}
	if width > 4000 {
		width = 4000
	}
	if r.raster != nil && r.raster.page == r.viewing {
		if r.raster.loading || r.raster.width >= width-16 || r.raster.err != "" {
			return
		}
	}
	if r.raster != nil && r.raster.loading && r.raster.page == r.viewing {
		return
	}

	page, want, path := r.viewing, width, r.doc.Path
	r.raster = &pdfPageRaster{page: page, w: float32(pw), h: float32(ph), loading: true}
	win := r.window
	if win == nil {
		return
	}
	go func() {
		doc, ok := pdf.Load(path)
		if !ok {
			win.Update(func() {
				if r.raster != nil && r.raster.page == page {
					r.raster.loading = false
					r.raster.err = "This file could not be read."
				}
			})
			return
		}
		img, err := doc.RenderPage(page, want)
		win.Update(func() {
			if r.raster == nil || r.raster.page != page {
				return
			}
			r.raster.loading = false
			if err != nil {
				r.raster.err = err.Error()
				return
			}
			r.raster.bitmap = ui.NewBitmap(img)
			r.raster.width, r.raster.height = img.Bounds().Dx(), img.Bounds().Dy()
		})
	}()
}

// pageView draws the page's pixels, a placeholder while they arrive, or the
// reason they could not be drawn.
func (r *Reader) pageView(c *ui.Context, target float32) ui.Element {
	p := r.raster
	if p == nil || p.page != r.viewing {
		return ui.Box(c).Width(target).Height(target * 1.414)
	}
	switch {
	case p.err != "":
		return ui.Box(c).Width(target).Height(target * 1.414).
			Background(c.Theme().SurfaceHover).
			Children(func() {
				ui.Text(c, fmt.Sprint("Page ", p.page+1, " could not be drawn: ", p.err)).
					TextColor(c.Theme().Danger).Padding(16)
			})
	case p.bitmap != nil:
		return ui.Box(c).Width(target).Children(func() {
			ui.Image(c, p.bitmap).FillWidth()
		})
	}
	return ui.Box(c).Width(target).Height(target * 1.414).
		Background(c.Theme().SurfaceHover).Radius(2)
}

// unusedImage keeps the image import honest while the bitmap path is what the
// view draws.
var _ image.Image
