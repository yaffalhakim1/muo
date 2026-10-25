package pdf

import (
	"bytes"
	"fmt"
	"slices"
)

// The page tree. A page inherits /Resources, /MediaBox, /CropBox and /Rotate
// from its ancestors, so the walk carries them down and copies them into each
// leaf; every later lookup on a page then sees the inherited values.

// collectPages walks the tree from the catalog and keeps the pages in order.
func (f *pdfFile) collectPages() error {
	if f.hasEncrypt {
		return fmt.Errorf("%w: the file is encrypted", errPDF)
	}
	// The catalog may be a plain object or live inside an object stream, so
	// every object number the file knows of is a candidate.
	f.expandObjectStreams()
	nums := make([]int, 0, len(f.offsets)+len(f.members))
	for num := range f.offsets {
		nums = append(nums, num)
	}
	for num := range f.members {
		nums = append(nums, num)
	}
	slices.Sort(nums)
	var root pdfValue
	for _, num := range nums {
		v, err := f.object(num)
		if err != nil || v.kind != pdfDict {
			continue
		}
		if t, _ := v.get("Type"); t.is("Catalog") {
			root = v
			break
		}
	}
	if root.kind != pdfDict {
		return fmt.Errorf("%w: no document catalog", errPDF)
	}
	seen := map[int]bool{}
	f.walkPages(mustGet(root, "Pages"), pdfValue{kind: pdfDict, dict: map[string]pdfValue{}}, seen, 0)
	if len(f.pages) == 0 {
		return fmt.Errorf("%w: no pages", errPDF)
	}
	return nil
}

func (f *pdfFile) walkPages(node, inherited pdfValue, seen map[int]bool, depth int) {
	if depth > 32 {
		return
	}
	if node.kind == pdfRef {
		num := node.int()
		if seen[num] {
			return
		}
		seen[num] = true
	}
	node = f.resolve(node)
	if node.kind != pdfDict {
		return
	}
	merged := map[string]pdfValue{}
	for k, v := range inherited.dict {
		merged[k] = v
	}
	for _, k := range []string{"Resources", "MediaBox", "CropBox", "Rotate"} {
		if v, ok := node.get(k); ok {
			merged[k] = v
		}
	}
	if kids, ok := node.get("Kids"); ok && kids.kind == pdfArray {
		for _, kid := range kids.array {
			f.walkPages(kid, pdfValue{kind: pdfDict, dict: merged}, seen, depth+1)
		}
		return
	}
	page := map[string]pdfValue{}
	for k, v := range merged {
		page[k] = v
	}
	for k, v := range node.dict {
		page[k] = v
	}
	f.pages = append(f.pages, pdfValue{kind: pdfDict, dict: page})
}

// PageSize returns a page's box in points, after rotation.
func (f *pdfFile) PageSize(index int) (w, h float64) {
	if index < 0 || index >= len(f.pages) {
		return 0, 0
	}
	box, ok := f.resolve(mustGet(f.pages[index], "CropBox")).numList(4)
	if !ok {
		box, ok = f.resolve(mustGet(f.pages[index], "MediaBox")).numList(4)
	}
	if !ok {
		return 595, 842
	}
	w, h = absf(box[2]-box[0]), absf(box[3]-box[1])
	if w == 0 || h == 0 {
		return 595, 842
	}
	if rot := f.rotation(index); rot == 90 || rot == 270 {
		w, h = h, w
	}
	return w, h
}

// PageBox returns the crop box in points, which the rasterizer maps to pixels.
func (f *pdfFile) PageBox(index int) (x, y, w, h float64) {
	box, ok := f.resolve(mustGet(f.pages[index], "CropBox")).numList(4)
	if !ok {
		box, ok = f.resolve(mustGet(f.pages[index], "MediaBox")).numList(4)
	}
	if !ok {
		return 0, 0, 595, 842
	}
	return minf(box[0], box[2]), minf(box[1], box[3]), absf(box[2] - box[0]), absf(box[3] - box[1])
}

func (f *pdfFile) rotation(index int) int {
	r := f.resolve(mustGet(f.pages[index], "Rotate")).int() % 360
	if r < 0 {
		r += 360
	}
	return r
}

// Contents returns a page's content stream, the several it may have joined.
func (f *pdfFile) Contents(index int) []byte {
	if index < 0 || index >= len(f.pages) {
		return nil
	}
	v := mustGet(f.pages[index], "Contents")
	if v.kind == pdfArray {
		var buf bytes.Buffer
		for _, e := range v.array {
			if raw, err := f.decodeStream(f.resolve(e)); err == nil {
				buf.Write(raw)
			}
			buf.WriteByte('\n')
		}
		return buf.Bytes()
	}
	raw, err := f.decodeStream(f.resolve(v))
	if err != nil {
		return nil
	}
	return raw
}

func absf(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func minf(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

