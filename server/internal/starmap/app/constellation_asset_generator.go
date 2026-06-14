package app

import (
	"context"

	"ego-server/internal/starmap/domain"
	writingdomain "ego-server/internal/writing/domain"

	"github.com/google/uuid"
)

// DefaultConstellationAssetGenerator is the MVP asset-generation policy.
// It produces a placeholder name, static insight, and starter topic prompts.
type DefaultConstellationAssetGenerator struct{}

func NewDefaultConstellationAssetGenerator() DefaultConstellationAssetGenerator {
	return DefaultConstellationAssetGenerator{}
}

func (DefaultConstellationAssetGenerator) Generate(_ context.Context, _ []writingdomain.Moment) (string, []float32, string, string, []string, error) {
	return "未命名的星座",
		nil,
		"星座" + uuid.New().String()[:8],
		"这些话语之间似乎有着某种共鸣。随着你写下更多，它们之间的联系会变得越来越清晰。",
		[]string{"关于这个主题，还有什么想说的吗？", "换个角度再看一看？"},
		nil
}

// Compile-time check.
var _ domain.ConstellationAssetGenerator = DefaultConstellationAssetGenerator{}
