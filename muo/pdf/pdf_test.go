package pdf

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Hand-built PDFs. Each writer below emits the smallest file that exercises
// one part of the reader, so a failure names the structure it broke rather than
// "some page looks odd".

type pdfWriter struct {
	buf    bytes.Buffer
	offset int
	refs   map[int]int
	order  []int
}

func newPDFWriter() *pdfWriter {
	w := &pdfWriter{refs: map[int]int{}}
	w.buf.WriteString("%PDF-1.7\n%\xe2\xe3\xcf\xd3\n")
	return w
}

// obj writes an object and records its offset for the xref.
func (w *pdfWriter) obj(num int, body string) {
	w.refs[num] = w.buf.Len()
	w.order = append(w.order, num)
	fmt.Fprintf(&w.buf, "%d 0 obj\n%s\nendobj\n", num, body)
}

// stream writes an object whose body is a stream.
func (w *pdfWriter) stream(num int, dict string, data []byte) {
	w.refs[num] = w.buf.Len()
	w.order = append(w.order, num)
	fmt.Fprintf(&w.buf, "%d 0 obj\n<< %s /Length %d >>\nstream\n", num, dict, len(data))
	w.buf.Write(data)
	w.buf.WriteString("\nendstream\nendobj\n")
}

// flateStream writes a stream whose payload is deflated.
func (w *pdfWriter) flateStream(num int, dict string, data []byte) {
	var z bytes.Buffer
	zw := zlib.NewWriter(&z)
	zw.Write(data)
	zw.Close()
	w.stream(num, dict+" /Filter /FlateDecode", z.Bytes())
}

// finish writes the xref table and the trailer. Objects must be numbered from
// 1 up, which the helpers here do.
func (w *pdfWriter) finish(root int) []byte {
	xref := w.buf.Len()
	max := 0
	for n := range w.refs {
		if n > max {
			max = n
		}
	}
	fmt.Fprintf(&w.buf, "xref\n0 %d\n", max+1)
	w.buf.WriteString("0000000000 65535 f \n")
	for n := 1; n <= max; n++ {
		if off, ok := w.refs[n]; ok {
			fmt.Fprintf(&w.buf, "%010d 00000 n \n", off)
			continue
		}
		w.buf.WriteString("0000000000 65535 f \n")
	}
	fmt.Fprintf(&w.buf, "trailer\n<< /Size %d /Root %d 0 R >>\nstartxref\n%d\n%%%%EOF\n", max+1, root, xref)
	return w.buf.Bytes()
}

// writeTemp writes bytes to a temporary file and returns its path.
func writeTemp(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// simplePDF builds a one-page file whose content is given.
func simplePDF(content string) []byte {
	w := newPDFWriter()
	w.obj(1, "<< /Type /Catalog /Pages 2 0 R >>")
	w.obj(2, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	w.obj(3, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>")
	w.stream(4, "", []byte(content))
	w.obj(5, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /FirstChar 32 /LastChar 126 /Widths ["+helveticaWidths()+"] >>")
	return w.finish(1)
}

func helveticaWidths() string {
	// The standard Helvetica advance widths for 32..126: the interpreter only
	// needs plausible numbers, but real ones make a wrong sum visible.
	var out bytes.Buffer
	for i := 32; i <= 126; i++ {
		if i > 32 {
			out.WriteString(" ")
		}
		out.WriteString("556")
	}
	return out.String()
}

func TestOpenPDFReadsStructure(t *testing.T) {
	path := writeTemp(t, "simple.pdf", simplePDF("BT /F1 12 Tf 72 720 Td (Hello, reader) Tj ET"))
	f, ok := Load(path)
	if !ok {
		t.Fatal("Load failed")
	}
	if got := f.PageCount(); got != 1 {
		t.Fatalf("page count = %d, want 1", got)
	}
	w, h := f.PageSize(0)
	if w != 612 || h != 792 {
		t.Fatalf("page size = %vx%v, want 612x792", w, h)
	}
	if content := f.file.Contents(0); !bytes.Contains(content, []byte("Hello")) {
		t.Fatalf("content stream lost its text: %q", content)
	}
}

func TestRenderPageDrawsText(t *testing.T) {
	path := writeTemp(t, "text.pdf", simplePDF("BT /F1 48 Tf 100 700 Td (MMMM) Tj ET"))
	doc, ok := Load(path)
	if !ok {
		t.Fatal("Load failed")
	}
	img, err := doc.RenderPage(0, 400)
	if err != nil {
		t.Fatal(err)
	}
	b := img.Bounds()
	if b.Dx() != 400 || b.Dy() < 500 {
		t.Fatalf("rendered %dx%d, want 400 wide", b.Dx(), b.Dy())
	}
	// The run sits near the top: count dark pixels in that band. At 400 px for
	// 612 pt, 48 pt of type is about 31 px tall.
	dark := 0
	for y := 0; y < 120; y++ {
		for x := 0; x < b.Dx(); x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			if r>>8 < 128 && g>>8 < 128 && bl>>8 < 128 {
				dark++
			}
		}
	}
	if dark < 200 {
		t.Fatalf("only %d dark pixels where the text run should be", dark)
	}
}

func TestRenderPageDrawsRectangle(t *testing.T) {
	path := writeTemp(t, "rect.pdf", simplePDF("0 0 0 rg 100 100 200 200 re f"))
	doc, _ := Load(path)
	img, err := doc.RenderPage(0, 612)
	if err != nil {
		t.Fatal(err)
	}
	// The rectangle covers x 100..300, y 100..300 in points; the raster is 612
	// wide and flipped, so its centre lands at device y = 792-200 = 592.
	r, g, b, _ := img.At(200, 592).RGBA()
	if r>>8 > 40 || g>>8 > 40 || b>>8 > 40 {
		t.Fatalf("centre of the rectangle is (%d,%d,%d), want black", r>>8, g>>8, b>>8)
	}
	// A corner well outside stays white.
	r, g, b, _ = img.At(10, 10).RGBA()
	if r>>8 < 200 {
		t.Fatalf("outside the rectangle is (%d,%d,%d), want white", r>>8, g>>8, b>>8)
	}
}

func TestRenderPageRotatedIsTaller(t *testing.T) {
	w := newPDFWriter()
	w.obj(1, "<< /Type /Catalog /Pages 2 0 R >>")
	w.obj(2, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	w.obj(3, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Rotate 90 /Contents 4 0 R >>")
	w.stream(4, "", []byte(""))
	data := w.finish(1)
	path := writeTemp(t, "rotated.pdf", data)
	doc, ok := Load(path)
	if !ok {
		t.Fatal("Load failed")
	}
	w1, h1 := doc.PageSize(0)
	if w1 != 792 || h1 != 612 {
		t.Fatalf("rotated page = %vx%v, want 792x612", w1, h1)
	}
}

func TestObjectStreamAndXrefStream(t *testing.T) {
	// PDF 1.5 style: the objects live in a compressed object stream and the
	// cross-reference is itself a stream.
	var objs bytes.Buffer
	// Object 1: catalog, object 2: pages, object 3: page.
	catalog := "<< /Type /Catalog /Pages 2 0 R >>"
	pages := "<< /Type /Pages /Kids [3 0 R] /Count 1 >>"
	page := "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 300 400] /Contents 4 0 R >>"
	first := 0
	for i, body := range []string{catalog, pages, page} {
		fmt.Fprintf(&objs, "%d %d ", i+1, first)
		first += len(body) + 1
	}
	fmt.Fprintln(&objs)
	header := objs.Bytes()
	var payload bytes.Buffer
	payload.Write(header)
	for _, body := range []string{catalog, pages, page} {
		payload.WriteString(body)
		payload.WriteString("\n")
	}

	w := newPDFWriter()
	var z bytes.Buffer
	zw := zlib.NewWriter(&z)
	zw.Write(payload.Bytes())
	zw.Close()
	w.stream(5, "/Type /ObjStm /N 3 /First "+fmt.Sprint(len(header))+" /Filter /FlateDecode", z.Bytes())
	w.stream(4, "", []byte(""))

	// The xref stream must be object 6; build the table by hand.
	xrefOff := w.buf.Len()
	rows := bytes.Buffer{}
	entries := []struct {
		kind, off int
	}{
		{0, 0}, {2, 5}, {2, 5}, {2, 5}, {1, w.refs[4]}, {1, w.refs[5]}, {1, xrefOff},
	}
	for _, e := range entries {
		rows.WriteByte(byte(e.kind))
		rows.Write([]byte{byte(e.off >> 24), byte(e.off >> 16), byte(e.off >> 8), byte(e.off)})
	}
	w.stream(6, "/Type /XRef /Size 7 /W [1 4 0] /Root 1 0 R", rows.Bytes())
	w.buf.WriteString("startxref\n")
	fmt.Fprintf(&w.buf, "%d\n%%%%EOF\n", xrefOff)

	path := writeTemp(t, "objstm.pdf", w.buf.Bytes())
	doc, ok := Load(path)
	if !ok {
		t.Fatal("Load failed on an object-stream file")
	}
	if doc.PageCount() != 1 {
		t.Fatalf("page count = %d, want 1", doc.PageCount())
	}
	pw, ph := doc.PageSize(0)
	if pw != 300 || ph != 400 {
		t.Fatalf("page = %vx%v, want 300x400", pw, ph)
	}
}

func TestEncryptedPDFIsRefused(t *testing.T) {
	w := newPDFWriter()
	w.obj(1, "<< /Type /Catalog /Pages 2 0 R >>")
	w.obj(2, "<< /Type /Pages /Kids [] /Count 0 >>")
	data := w.finish(1)
	data = bytes.Replace(data, []byte("/Size"), []byte("/Encrypt << /Filter /Standard >> /Size"), 1)
	path := writeTemp(t, "encrypted.pdf", data)
	if _, ok := Load(path); ok {
		t.Fatal("an encrypted file was accepted")
	}
}

func TestDamagedFileIsRefused(t *testing.T) {
	path := writeTemp(t, "broken.pdf", []byte("%PDF-1.7\nthis is not a pdf"))
	if _, ok := Load(path); ok {
		t.Fatal("a damaged file was accepted")
	}
}
