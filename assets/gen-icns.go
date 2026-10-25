//go:build ignore

// Packs assets/icon-*.png into resources/icon.icns, the macOS app icon.
//
// ICNS is a flat container: a magic header, then one entry per size, each an
// OSType and a big-endian length followed by PNG data. macOS reads PNG entries
// directly, so the images the icon generator already wrote go in unchanged.
//
// Run with `go run assets/gen-icns.go`, after `go run assets/gen-icon.go`.
package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// The entries macOS asks for. A Retina icon is two files: the logical size and
// the doubled one, which is why 16 appears twice with different OSTypes.
var entries = []struct {
	kind string // the OSType
	size int    // the pixel size of the PNG
}{
	{"icp4", 16},
	{"icp5", 32},
	{"icp6", 64},
	{"ic07", 128},
	{"ic08", 256},
	{"ic09", 512},
	{"ic10", 1024},
	{"ic11", 32},  // 16 at 2x
	{"ic12", 64},  // 32 at 2x
	{"ic13", 256}, // 128 at 2x
	{"ic14", 512}, // 256 at 2x
}

func main() {
	var body []byte
	for _, e := range entries {
		path := filepath.Join("assets", fmt.Sprintf("icon-%d.png", e.size))
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gen-icns: %v\n", err)
			os.Exit(1)
		}
		// Each entry: the OSType, then the length including this 8-byte header.
		entry := make([]byte, 8+len(data))
		copy(entry[0:4], e.kind)
		binary.BigEndian.PutUint32(entry[4:8], uint32(len(entry)))
		copy(entry[8:], data)
		body = append(body, entry...)
	}

	// The file: 'icns', the length of everything, then the entries.
	out := make([]byte, 8+len(body))
	copy(out[0:4], "icns")
	binary.BigEndian.PutUint32(out[4:8], uint32(len(out)))
	copy(out[8:], body)

	path := filepath.Join("resources", "icon.icns")
	if err := os.WriteFile(path, out, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "gen-icns: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s (%d bytes, %d entries)\n", path, len(out), len(entries))
}
