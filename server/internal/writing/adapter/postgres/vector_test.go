package postgres

import "testing"

func TestVectorLiteral(t *testing.T) {
	got, err := vectorLiteral([]float32{0.1, -0.25, 3}, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "[0.1,-0.25,3]" {
		t.Fatalf("expected vector literal [0.1,-0.25,3], got %s", got)
	}
}

func TestVectorLiteralRejectsDimensionMismatch(t *testing.T) {
	_, err := vectorLiteral([]float32{0.1, 0.2}, 3)
	if err == nil {
		t.Fatal("expected dimension mismatch error")
	}
}

func TestVectorLiteralRejectsEmptyEmbedding(t *testing.T) {
	_, err := vectorLiteral(nil, 0)
	if err == nil {
		t.Fatal("expected empty embedding error")
	}
}
