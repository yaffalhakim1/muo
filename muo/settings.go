package muo

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// Reading settings. It lives in the content pane, beside the source list the
// window already has, so it must not bring a second navigation column of its
// own: a settings surface with its own sidebar beside the app's sidebar is two
// sidebars, which is what made the first version look assembled.
//
// Two things make it read as a pane rather than as a form dropped on a page.
// The groups sit on cards, so where one concern ends and the next begins is a
// surface rather than a line of bold text; and above them is a specimen, set at
// the size and measure being chosen. The settings are about how type looks, so
// the pane shows type instead of asking the reader to go back and check.

func (r *Reader) settingsView(c *ui.Context) {
	t := c.Theme()
	pal := pageColors(c)
	bar := c.TitleBar()

	ui.Column(c).Fill().Background(pal.paperColor()).Children(func() {
		// The pane's own title row, at the height of the window's, so the two
		// panes change places without the window jumping. It clears the window
		// controls, which sit over the content when the title bar is inset.
		ui.Row(c).Height(titleBarHeight).Shrink(0).AlignItems(ui.Center).
			Padding(0, bar.Right+20, 0, titleLeft).Children(func() {
				ui.Text(c, "Reading Settings").FontSize(typeTitle3).FontWeight(600).
					LetterSpacing(-0.1)
				ui.Spacer(c)
				// Done is not the accent-filled primary it was: nothing is being
				// committed here, since every change applies as it is made, so the
				// way out should not be the loudest thing in the pane.
				if ui.Button(c, "Done").Clicked() {
					r.settings = false
				}
			})

		ui.Scroll(c).Fill().Children(func() {
			ui.Row(c).FillWidth().
				Padding(t.Space(6), pageMargin, t.Space(12), pageMargin).Children(func() {
				ui.Spacer(c)
				ui.Column(c).Width(560).MaxWidthPercent(100).
					Gap(t.Space(6)).Children(func() {
					r.specimen(c, pal)
					r.settingsForm(c)
				})
				ui.Spacer(c)
			})
		})
	})
}

// specimen is the type being chosen, shown. It is the same reading measure the
// page uses, so what the reader judges here is what they will read.
func (r *Reader) specimen(c *ui.Context, pal pagePalette) {
	t := c.Theme()
	ui.Column(c).FillWidth().Padding(t.Space(5)).Radius(12).Gap(t.Space(3)).
		Background(uiRGB(mixColor(pal.paper, pal.ink, 0.035))).
		Border(1, uiRGB(mixColor(pal.paper, pal.ink, 0.05))).
		Children(func() {
			ui.Text(c, "The measure comes first").
				FontSize(r.typ.FontPx * 0.88).FontWeight(600).
				LetterSpacing(-r.typ.FontPx * 0.012).
				TextColor(pal.inkColor())
			ui.Text(c, "The largest effect on reading comfort is how many characters sit on a line. This paragraph is set at the size and measure the page would give it.").
				FontSize(r.typ.FontPx).LineHeight(r.typ.Line).
				TextColor(pal.inkColor())
			ui.Row(c).FillWidth().Gap(t.Space(2)).AlignItems(ui.Center).Children(func() {
				typeReadout(c, r.typ)
				ui.Spacer(c)
				ui.Textf(c, "%.1f pt", r.typ.FontPx).FontSize(typeFootnote).
					FontFeatures("tnum").TextColor(pal.mutedColor())
			})
		})
}

// settingsForm is the form itself: one card per concern, each field naming one
// thing the reader can change and carrying its value beside it, so a slider is
// read rather than estimated.
func (r *Reader) settingsForm(c *ui.Context) {
	t := c.Theme()
	d := r.display
	if d == nil {
		d = &Display{Label: "This Display", Distance: 45, WidthCm: 60}
	}

	distance := d.Distance
	var size float64
	switch {
	case d.FontPx > 0:
		size = float64(d.FontPx)
	case r.manual > 0:
		size = float64(r.manual)
	default:
		size = float64(r.typ.PreferredSize)
	}
	label := d.Label
	// The control writes into a stable value, so the choice survives the next
	// frame: a pointer into a slice rebuilt each frame is a different variable
	// every time and the choice is lost.
	mode := int(appearanceMode.Load())

	ui.Form(c, func() {
		settingsCard(c, "Page", func() {
			ui.Field(c, "Distance", func() {
				ui.Row(c).FillWidth().Gap(t.Space(3)).AlignItems(ui.Center).Children(func() {
					ui.StepSlider(c, &distance, 30, 100, 5).Grow(1)
					ui.Textf(c, "%.0f cm", distance).FontSize(typeCallout).
						FontFeatures("tnum").TextColor(t.TextMuted).
						Width(54).TextAlign(ui.End)
				})
			}).Description("From your eyes to the screen. The type size is solved from it.")

			ui.Field(c, "Type size", func() {
				ui.Row(c).FillWidth().Gap(t.Space(3)).AlignItems(ui.Center).Children(func() {
					ui.StepSlider(c, &size, 12, 44, 0.5).Grow(1)
					ui.Textf(c, "%.1f pt", size).FontSize(typeCallout).
						FontFeatures("tnum").TextColor(t.TextMuted).
						Width(54).TextAlign(ui.End)
				})
			}).Description(fmt.Sprintf("The model asks for %.1f pt on this screen.", r.typ.PreferredSize))

			if d.FontPx > 0 {
				ui.Field(c, "", func() {
					if ui.Button(c, "Follow the Model Again").Clicked() {
						d.FontPx = 0
						r.manual = 0
					}
				}).Description("Give the size back to the model.")
			}
		})

		settingsCard(c, "Appearance", func() {
			ui.Field(c, "Mode", func() {
				ui.Segmented(c, &mode, appearanceNames...).Label("Appearance")
			}).Description("Follows the desktop unless you choose otherwise.")
		})

		settingsCard(c, "This Screen", func() {
			ui.Field(c, "Name", func() {
				ui.TextInput(c, &label)
			}).Description("Each monitor keeps its own distance and size.")
		})
	})

	// The values are written after the form is built, so a change this frame
	// applies to the next one rather than mid-layout.
	if distance != d.Distance {
		d.Distance = distance
		d.FontPx = 0
		r.manual = 0
	}
	if float32(size) != d.FontPx && size != float64(r.typ.PreferredSize) {
		r.manual = float32(size)
		d.FontPx = float32(size)
	}
	if label != d.Label {
		d.Label = label
	}
	if mode != int(appearanceMode.Load()) {
		setAppearance(AppearanceMode(mode))
	}
}

// settingsCard is one group of fields: a small name over a surface that holds
// them. The surface is what tells the reader the fields belong together, which
// a bold line of text above a bare run of controls does not.
func settingsCard(c *ui.Context, legend string, fn func()) {
	t := c.Theme()
	ui.Column(c).FillWidth().Gap(t.Space(1.5)).Children(func() {
		if legend != "" {
			ui.Text(c, legend).FontSize(typeSubhead).FontWeight(600).
				LetterSpacing(0.3).TextColor(t.TextMuted).
				Padding(0, t.Space(1))
		}
		ui.Column(c).FillWidth().Padding(t.Space(4)).Radius(12).Gap(t.Space(3)).
			Background(t.Surface).Border(1, t.Border).Children(fn)
	})
}
