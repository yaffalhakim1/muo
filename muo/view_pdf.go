package muo

import (
	"image"

	"github.com/egoist/mygo/ui"
)

// The PDF surface. A page is rasterized off the UI thread at the window's own
// pixel density and shown in a column wide enough for the page; the reading
// research applies to PDFs too, so a page never stretches to the window's full
// width.

// pdfContent draws the current page, its navigation and, when the file cannot be
// read, the reason.
//
// It is drawn inside the pane's scroll container, not its own: the page is a
// picture, and the reader scrolls it exactly as they scroll a page of text. The
// top padding clears the title bar, which floats over this pane.
func (r *Reader) pdfContent(c *ui.Context, pal pagePalette, top float32) {
	if r.doc.Err != "" {
		ui.Column(c).Fill().Center().Padding(48).Gap(c.Theme().Space(2)).Children(func() {
			ui.Text(c, r.doc.Err).TextColor(c.Theme().Danger).MaxWidth(460).TextAlign(ui.Center)
		})
		return
	}
	if r.file == nil {
		ui.Text(c, "This PDF could not be read.").TextColor(pal.mutedColor())
		return
	}

	// The page is capped by the column the type model chose and by the height
	// the window can show: a page wider than the reading column is a page the
	// eye cannot sweep in one pass.
	_, ch := c.Size()
	target := r.typ.ColumnW * 1.15
	if h := float32(ch) - 96; h > 0 {
		if pw, ph := r.file.PageSize(r.viewing); pw > 0 {
			if byHeight := h * float32(pw/ph); byHeight < target {
				target = byHeight
			}
		}
	}
	if target < 240 {
		target = 240
	}
	// The scale the page is drawn at, which turns a match's position in points
	// into a scroll offset in the view.
	if pw, _ := r.file.PageSize(r.viewing); pw > 0 {
		r.pageScale = target / float32(pw)
	}
	r.beginPage(c, target)

	ui.Column(c).FillWidth().Padding(top, 0, pageBottomMargin, 0).Children(func() {
		ui.Row(c).FillWidth().Children(func() {
			ui.Spacer(c)
			ui.Column(c).Width(target).Shrink(0).Padding(0, 0, 0, 0).Gap(c.Theme().Space(3)).Children(func() {
				if r.raster != nil && r.raster.page == r.viewing && r.raster.bitmap != nil {
					ui.Box(c).FillWidth().Radius(2).
						Shadow(0, 2, 18, 0, ui.RGBA(0, 0, 0, 0.16)).Children(func() {
						ui.Image(c, r.raster.bitmap).FillWidth()
					})
				} else {
					ui.Box(c).FillWidth().Height(target * 1.414).
						Background(ui.RGBA(0, 0, 0, 0.05)).Radius(2)
				}
				r.pageNav(c, pal.ink, pal.muted)
			})
			ui.Spacer(c)
		})
	})
	// A PDF's match is found by the page's text, so the scroll happens after the
	// page is laid out and the view knows how far it can go.
	r.followFindPdf()
}
// pageNav is the page counter and the arrows, under the page. The arrows are
// icons and the count is in tabular figures, so the control does not reflow as
// the number of pages changes.
func (r *Reader) pageNav(c *ui.Context, fg, muted rgbColor) {
	t := c.Theme()
	n := r.file.PageCount()
	ui.Row(c).FillWidth().Center().Gap(t.Space(1)).AlignItems(ui.Center).Children(func() {
		prev := iconButton(c, iconLeft, "Previous Page", 28, 26)
		if r.viewing <= 0 {
			prev.Disabled(true)
		}
		if prev.Clicked() && r.viewing > 0 {
			r.goToPage(r.viewing - 1)
		}
		ui.Textf(c, "%d of %d", r.viewing+1, n).
			FontSize(typeCallout).FontFeatures("tnum").
			TextColor(uiRGB(muted)).Padding(0, t.Space(1))
		next := iconButton(c, iconRight, "Next Page", 28, 26)
		if r.viewing >= n-1 {
			next.Disabled(true)
		}
		if next.Clicked() && r.viewing < n-1 {
			r.goToPage(r.viewing + 1)
		}
	})
	_ = fg
}

func (r *Reader) goToPage(index int) {
	if r.file == nil || index < 0 || index >= r.file.PageCount() {
		return
	}
	r.viewing = index
	r.raster = nil
}

// The page raster of one page, and the request that started the render.

var _ = image.Rectangle{}






