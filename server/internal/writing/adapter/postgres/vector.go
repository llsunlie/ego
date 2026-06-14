package postgres

import (
	"fmt"
	"strconv"
	"strings"
)

func vectorLiteral(values []float32, expectedDim int) (string, error) {
	if expectedDim > 0 && len(values) != expectedDim {
		return "", fmt.Errorf("embedding dimension mismatch: got %d, want %d", len(values), expectedDim)
	}
	if len(values) == 0 {
		return "", fmt.Errorf("embedding is empty")
	}

	var b strings.Builder
	b.WriteByte('[')
	for i, v := range values {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(float64(v), 'g', -1, 32))
	}
	b.WriteByte(']')
	return b.String(), nil
}
