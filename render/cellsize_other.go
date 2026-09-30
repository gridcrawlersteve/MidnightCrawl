//go:build windows || js

package render

// DetectCellSize returns default cell dimensions on Windows.
// Windows terminals generally don't support sixel, so this is a fallback.
func DetectCellSize() (int, int) {
	return CellWidth, CellHeight
}
