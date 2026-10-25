package pdf

import (
	"fmt"
	"image"
	"sync"
)

// Document is a parsed file kept between frames and shared by the goroutine
// that renders a page, so that turning a page does not parse the file again.
type Document struct {
	file *pdfFile
	mu   sync.Mutex
}

var (
	pdfCacheMu sync.Mutex
	pdfCache   = map[string]*Document{}
)

// Load returns the parsed file for path, opening it once.
func Load(path string) (*Document, bool) {
	pdfCacheMu.Lock()
	defer pdfCacheMu.Unlock()
	if d, ok := pdfCache[path]; ok {
		return d, true
	}
	data, err := readFile(path)
	if err != nil {
		return nil, false
	}
	f, err := openPDF(data)
	if err != nil {
		return nil, false
	}
	d := &Document{file: f}
	pdfCache[path] = d
	if len(pdfCache) > 4 {
		for k := range pdfCache {
			if k != path {
				delete(pdfCache, k)
				break
			}
		}
	}
	return d, true
}

// Forget drops a cached file, so that reloading re-reads it.
func Forget(path string) {
	pdfCacheMu.Lock()
	delete(pdfCache, path)
	pdfCacheMu.Unlock()
}

func (d *Document) PageCount() int { return len(d.file.pages) }

// HasPage reports whether the file has the given zero-based page.
func (d *Document) HasPage(index int) bool {
	return index >= 0 && index < len(d.file.pages)
}

func (d *Document) PageSize(index int) (w, h float64) { return d.file.PageSize(index) }

// RenderPage rasterizes one page at the given pixel width. The interpreter
// runs under the file's lock because it walks the object graph that a second
// page's render would also walk; a window's pages are sequential anyway.
func (d *Document) RenderPage(index, width int) (image.Image, error) {
	if index < 0 || index >= len(d.file.pages) {
		return nil, fmt.Errorf("no page %d", index+1)
	}
	_, _, pw, ph := d.file.PageBox(index)
	if pw <= 0 || ph <= 0 {
		return nil, fmt.Errorf("the page has no size")
	}
	height := int(float64(width) * ph / pw)
	if height < 1 || height > 20000 || width < 1 {
		return nil, fmt.Errorf("the page is too large to draw")
	}
	canvas := newCanvas(width, height, 0xff, 0xff, 0xff)

	d.mu.Lock()
	defer d.mu.Unlock()
	interp := &pageInterpreter{
		file:   d.file,
		page:   index,
		canvas: canvas,
		scale:  float64(width) / pw,
		state:  defaultGraphicsState(),
		forms:  map[int]bool{},
	}
	_ = interp.run(d.file.Contents(index), 0)
	return canvas.image(), nil
}
