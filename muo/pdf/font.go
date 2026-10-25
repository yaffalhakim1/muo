package pdf

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"math"
)

// Fonts. A PDF text operator gives glyph codes; the font dictionary says how
// wide each is and, through a ToUnicode map, which character it is. The
// reader needs both: advances place the run, characters make it searchable
// and copyable.

type pdfFont struct {
	firstChar int
	widths    []float64 // by code - firstChar, in thousandths of an em
	missing   float64   // the default width, thousandths
	toUnicode map[uint32]rune
	twoByte   bool
	subtype   string
	// baseFont is the font's own name, as "Helvetica-Bold", which picks the
	// family the text is set in when the program itself is not interpreted.
	baseFont string
}

type decodedGlyph struct {
	code uint32
	text string
}

// loadFont reads a font dictionary into what the run reader needs.
func (f *pdfFile) loadFont(dict pdfValue) *pdfFont {
	font := &pdfFont{missing: 500, toUnicode: map[uint32]rune{}}
	font.subtype = mustGet(dict, "Subtype").name
	font.baseFont = mustGet(dict, "BaseFont").name
	if font.subtype == "Type0" {
		font.twoByte = true
	}
	if v, ok := dict.get("FirstChar"); ok {
		font.firstChar = v.int()
	}
	if v, ok := dict.get("Widths"); ok {
		if v = f.resolve(v); v.kind == pdfArray {
			for _, e := range v.array {
				if e = f.resolve(e); e.kind == pdfNumber {
					font.widths = append(font.widths, e.num)
				} else {
					font.widths = append(font.widths, font.missing)
				}
			}
		}
	}
	if v, ok := dict.get("MissingWidth"); ok {
		font.missing = v.float()
	}
	// A Type0 font takes its widths from the descendant.
	if font.subtype == "Type0" {
		desc := f.resolve(mustGet(dict, "DescendantFonts"))
		if desc.kind == pdfArray && len(desc.array) > 0 {
			child := f.resolve(desc.array[0])
			if w, ok := child.get("DW"); ok {
				font.missing = w.float()
			}
			if w := f.resolve(mustGet(child, "W")); w.kind == pdfArray {
				f.readCIDWidths(font, w.array)
			}
		}
	}
	if v, ok := dict.get("ToUnicode"); ok {
		if raw, err := f.decodeStream(f.resolve(v)); err == nil {
			font.readCMap(raw)
		}
	}
	return font
}

// readCIDWidths reads the /W array: [c [w w w] cFirst cLast w …].
func (f *pdfFile) readCIDWidths(font *pdfFont, arr []pdfValue) {
	if len(font.widths) == 0 {
		font.widths = make([]float64, 65536)
	}
	for i := 0; i < len(arr); {
		start := f.resolve(arr[i]).int()
		i++
		if i >= len(arr) {
			return
		}
		next := f.resolve(arr[i])
		if next.kind == pdfArray {
			for j, w := range next.array {
				if cid := start + j; cid < len(font.widths) {
					font.widths[cid] = f.resolve(w).float()
				}
			}
			i++
			continue
		}
		i++
		if i >= len(arr) {
			return
		}
		end, w := next.int(), f.resolve(arr[i]).float()
		i++
		for cid := start; cid <= end && cid < len(font.widths); cid++ {
			font.widths[cid] = w
		}
	}
}

// readToUnicode parses the CMap's bfchar and bfrange sections.
func (font *pdfFont) readCMap(raw []byte) {
	text := raw
	inBfchar := false
	inBfrange := false
	for i := 0; i < len(text); {
		at := skipSpace(text, i)
		if at >= len(text) {
			return
		}
		switch {
		case bytes.HasPrefix(text[at:], []byte("beginbfchar")):
			inBfchar, inBfrange = true, false
			i = at + len("beginbfchar")
		case bytes.HasPrefix(text[at:], []byte("endbfchar")):
			inBfchar = false
			i = at + len("endbfchar")
		case bytes.HasPrefix(text[at:], []byte("beginbfrange")):
			inBfrange, inBfchar = true, false
			i = at + len("beginbfrange")
		case bytes.HasPrefix(text[at:], []byte("endbfrange")):
			inBfrange = false
			i = at + len("endbfrange")
		case inBfchar || inBfrange:
			if text[at] == '<' {
				end := bytes.IndexByte(text[at:], '>')
				if end < 0 {
					return
				}
				src := asciiHexDecode(text[at+1 : at+end])
				i = at + end + 1
				next := skipSpace(text, i)
				if next >= len(text) {
					return
				}
				switch {
				case text[next] == '<':
					closeAt := bytes.IndexByte(text[next:], '>')
					if closeAt < 0 {
						return
					}
					dst := asciiHexDecode(text[next+1 : next+closeAt])
					i = next + closeAt + 1
					if inBfrange {
						// <lo> <hi> <dst>: successive code points.
						if closeAt2 := bytes.IndexByte(text[i:], '>'); closeAt2 > 0 {
							if hi := asciiHexDecode(bytes.TrimPrefix(bytes.TrimSpace(text[skipSpace(text, i):i+closeAt2+1]), []byte("<"))); len(hi) > 0 {
								lo := codeOf(src)
								high := codeOf(hi)
								base := codeOf(dst)
								for c := lo; c <= high && c-lo < 65536; c++ {
									font.toUnicode[c] = rune(base + (c - lo))
								}
								i += closeAt2 + 1
							}
						}
						continue
					}
					if r := runesOf(dst); len(r) > 0 {
						font.toUnicode[codeOf(src)] = r[0]
					}
				case text[next] == '[':
					// A range whose destinations are listed one by one.
					closeAt := bytes.IndexByte(text[next:], ']')
					if closeAt < 0 {
						return
					}
					lo := codeOf(src)
					body := text[next+1 : next+closeAt]
					k := 0
					for j := 0; j < len(body); {
						at2 := skipSpace(body, j)
						if at2 >= len(body) {
							break
						}
						if body[at2] == '<' {
							c := bytes.IndexByte(body[at2:], '>')
							if c < 0 {
								break
							}
							if r := runesOf(asciiHexDecode(body[at2+1 : at2+c])); len(r) > 0 {
								font.toUnicode[lo+uint32(k)] = r[0]
							}
							k++
							j = at2 + c + 1
							continue
						}
						j = at2 + 1
					}
					i = next + closeAt + 1
				default:
					i = next
				}
				continue
			}
			i = at + 1
		default:
			i = at + 1
		}
	}
}

// decode splits a string into glyphs, one byte per code for simple fonts and
// two for composite ones.
func (f *pdfFont) decode(b []byte) []decodedGlyph {
	if !f.twoByte {
		out := make([]decodedGlyph, 0, len(b))
		for _, c := range b {
			g := decodedGlyph{code: uint32(c)}
			if r, ok := f.toUnicode[uint32(c)]; ok {
				g.text = string(r)
			}
			out = append(out, g)
		}
		return out
	}
	out := make([]decodedGlyph, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		code := uint32(b[i])<<8 | uint32(b[i+1])
		g := decodedGlyph{code: code}
		if r, ok := f.toUnicode[code]; ok {
			g.text = string(r)
		}
		out = append(out, g)
	}
	return out
}

// width returns a glyph's advance in thousandths of an em.
func (f *pdfFont) width(code uint32) float64 {
	idx := int(code) - f.firstChar
	if f.twoByte {
		if int(code) < len(f.widths) && f.widths[code] != 0 {
			return f.widths[code]
		}
		return f.missing
	}
	if idx >= 0 && idx < len(f.widths) && f.widths[idx] != 0 {
		return f.widths[idx]
	}
	return f.missing
}

func codeOf(b []byte) uint32 {
	var v uint32
	for _, c := range b {
		v = v<<8 | uint32(c)
	}
	return v
}

func runesOf(b []byte) []rune {
	// A destination may be UTF-16BE.
	if len(b) >= 2 && len(b)%2 == 0 {
		var out []rune
		for i := 0; i+1 < len(b); i += 2 {
			out = append(out, rune(uint16(b[i])<<8|uint16(b[i+1])))
		}
		return out
	}
	return []rune(string(b))
}

// Image codecs -------------------------------------------------------------

func decodeJPEG(raw []byte) (image.Image, error) {
	img, err := jpeg.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("%w: JPEG", errPDF)
	}
	return img, nil
}

// decodeJPX returns nil rather than an error: a page with a JPEG 2000 image
// it cannot decode still shows its text and vectors.
func decodeJPX(raw []byte) (image.Image, error) { return nil, nil }

// decodeCCITT returns nil: fax-encoded scans are rare enough that a blank
// image beats a wrong one.
func decodeCCITT(dict pdfValue, raw []byte, w, h int) image.Image { return nil }

// maths helpers the interpreter shares
func radians(d float64) float64 { return d * math.Pi / 180 }

// readToUnicode is a method of the file so that a caller may decode a stream
// it has already resolved; the font's map is what it fills.
func (f *pdfFile) readToUnicode(font *pdfFont, dict pdfValue) {
	if v, ok := dict.get("ToUnicode"); ok {
		if raw, err := f.decodeStream(f.resolve(v)); err == nil {
			font.readCMap(raw)
		}
	}
}
