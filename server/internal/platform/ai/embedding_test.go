package ai_test

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"ego-server/internal/config"
	"ego-server/internal/platform/ai"
)

// =========================================================================
// Test data — see .harness/embedding_test_design.md for rationale
// =========================================================================

// Similar group 1: imposter syndrome about career transitions.
// Different surface wording, same underlying emotion/topic.
var similarGroup1 = []string{
	"下个月就要正式入职了。虽然拿到了满意的 offer，但心里一直悬着，总怕自己真实写代码的能力没有面试时表现出来的那么好。",
	"终于把最后的论文提交了。可是看着日历上越来越近的报到日，突然觉得很心虚。在学校里做的那些并行计算的研究，到了真正的工业界是不是根本派不上用场？",
	"最近总是下意识地逃避去看公司发来的新人文档。明明之前很期待去新团队做后台架构的，现在真到了眼前，反而有点不敢面对。",
}

// Similar group 2: photography and self-reflection.
// Aesthetic sensibilities (tones, editing) tied to emotional states.
var similarGroup2 = []string{
	"今天帮朋友拍了一组人像，后期修图的时候特意拉低了对比度，还加了一点点颗粒。感觉这种偏冷、偏暗的色调更能衬托出她平时那种安静的特质。",
	"试着用新写的代码跑了一下最近自己最满意的几张照片，发现它们都在高光部分带了一层极淡的蓝紫色。原来我潜意识里一直在追求这种清冷感。",
	"每次给别人按快门，其实也是在捕捉我自己当下的状态。画面里的人笑得再开心，如果那天我心里觉得很丧，选出来的照片底色也会透着一种孤单。",
}

// Dissimilar pair 1: same keywords ("信号传递", "头大"), completely different domains.
var dissimilarPair1 = [2]string{
	"Go 里面的 context 传递有时候真让人头大，特别是在嵌套了多层协程之后，一旦某个环节没有正确捕获 cancel 信号，就很容易导致 goroutine 泄露。",
	"人际关系里的信号传递有时候真让人头大，特别是在有了误会之后，一旦某个人没有正确捕捉到对方的言外之意，就很容易导致情绪崩溃。",
}

// Dissimilar pair 2: trivial daily log vs. deep introspection about control.
var dissimilarPair2 = [2]string{
	"今天出门又忘了带伞，跑到楼下便利店买了一把透明的。这好像是我今年买的第三把同样的伞了，每次用完就随手一丢，根本记不住放哪了。",
	"这次去马来西亚的行程安排得还是太满。我发现自己总是习惯性地把每一天都填满，好像一旦停下来无事可做，就会有一种失控的恐慌。下次出去玩，必须强制自己留出一整天只在酒店躺着。",
}

// =========================================================================
// Helpers
// =========================================================================

func loadEnvAndClient(t *testing.T) *ai.Client {
	t.Helper()

	_ = config.Load()

	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		t.Skip("AI_API_KEY not set in env or .env, skipping embedding test")
	}

	baseURL := envOrDefault("AI_BASE_URL", "https://api.siliconflow.cn/v1")
	return ai.NewClient(ai.Config{
		EmbeddingAPIKey:  apiKey,
		EmbeddingBaseURL: baseURL,
		EmbeddingModel:   envOrDefault("AI_EMBEDDING_MODEL", "Qwen/Qwen3-VL-Embedding-8B"),
		ChatAPIKey:       apiKey,
		ChatBaseURL:      baseURL,
		ChatModel:        envOrDefault("AI_CHAT_MODEL", "deepseek-ai/DeepSeek-V3"),
	})
}

// embedAll returns embeddings for a slice of texts.
func embedAll(t *testing.T, client *ai.Client, texts []string) [][]float32 {
	t.Helper()
	ctx := context.Background()
	results := make([][]float32, len(texts))
	for i, text := range texts {
		r, err := client.CreateEmbedding(ctx, text)
		if err != nil {
			t.Fatalf("embed text[%d]: %v", i, err)
		}
		results[i] = r.Embedding
	}
	return results
}

// sim is shorthand for CosineSimilarity.
func sim(a, b []float32) float64 {
	return ai.CosineSimilarity(a, b)
}

type pair struct {
	i, j int
	sim  float64
}

func formatPairs(pairs []pair) string {
	var b strings.Builder
	for _, p := range pairs {
		fmt.Fprintf(&b, "  [%d↔%d] %.4f\n", p.i, p.j, p.sim)
	}
	return b.String()
}

// =========================================================================
// Tests
// =========================================================================

func TestEmbedding_SimilarGroup1_ImposterSyndrome(t *testing.T) {
	client := loadEnvAndClient(t)
	emb := embedAll(t, client, similarGroup1)

	pairs := []pair{
		{0, 1, sim(emb[0], emb[1])},
		{0, 2, sim(emb[0], emb[2])},
		{1, 2, sim(emb[1], emb[2])},
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].sim > pairs[j].sim })

	t.Log("Imposter syndrome group — pairwise similarities:\n" + formatPairs(pairs))

	minSim := pairs[len(pairs)-1].sim
	if minSim < 0.55 {
		t.Errorf("expected all within-group similarities >= 0.55, got min=%.4f", minSim)
	}
}

func TestEmbedding_SimilarGroup2_Photography(t *testing.T) {
	client := loadEnvAndClient(t)
	emb := embedAll(t, client, similarGroup2)

	pairs := []pair{
		{0, 1, sim(emb[0], emb[1])},
		{0, 2, sim(emb[0], emb[2])},
		{1, 2, sim(emb[1], emb[2])},
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].sim > pairs[j].sim })

	t.Log("Photography group — pairwise similarities:\n" + formatPairs(pairs))

	minSim := pairs[len(pairs)-1].sim
	if minSim < 0.55 {
		t.Errorf("expected all within-group similarities >= 0.55, got min=%.4f", minSim)
	}
}

func TestEmbedding_DissimilarPair1_KeywordCollision(t *testing.T) {
	client := loadEnvAndClient(t)
	ctx := context.Background()

	r0, _ := client.CreateEmbedding(ctx, dissimilarPair1[0])
	r1, _ := client.CreateEmbedding(ctx, dissimilarPair1[1])
	s := sim(r0.Embedding, r1.Embedding)

	t.Logf("keyword collision pair similarity: %.4f (expect < 0.6)", s)
	t.Logf("  A (tech):   %s", truncate(dissimilarPair1[0], 60))
	t.Logf("  B (social): %s", truncate(dissimilarPair1[1], 60))

	if s >= 0.75 {
		t.Errorf("keyword-collision similarity too high (%.4f): model may be overfitting to surface words", s)
	}
}

func TestEmbedding_DissimilarPair2_TrivialVsDeep(t *testing.T) {
	client := loadEnvAndClient(t)
	ctx := context.Background()

	r0, _ := client.CreateEmbedding(ctx, dissimilarPair2[0])
	r1, _ := client.CreateEmbedding(ctx, dissimilarPair2[1])
	s := sim(r0.Embedding, r1.Embedding)

	t.Logf("trivial-vs-deep pair similarity: %.4f (expect < 0.6)", s)
	t.Logf("  C (trivial): %s", truncate(dissimilarPair2[0], 60))
	t.Logf("  D (deep):    %s", truncate(dissimilarPair2[1], 60))

	if s >= 0.75 {
		t.Errorf("trivial-vs-deep similarity too high (%.4f): model should separate surface diary from introspection", s)
	}
}

// TestEmbedding_CrossGroupSeparation verifies that texts from different
// similar groups are farther apart than texts within the same group.
func TestEmbedding_CrossGroupSeparation(t *testing.T) {
	client := loadEnvAndClient(t)

	allTexts := append([]string{}, similarGroup1...)
	allTexts = append(allTexts, similarGroup2...)
	emb := embedAll(t, client, allTexts)

	// Within-group similarities.
	var within []float64
	within = append(within, sim(emb[0], emb[1]), sim(emb[0], emb[2]), sim(emb[1], emb[2]))
	within = append(within, sim(emb[3], emb[4]), sim(emb[3], emb[5]), sim(emb[4], emb[5]))

	// Cross-group similarities.
	var cross []float64
	for i := range 3 {
		for j := range 3 {
			cross = append(cross, sim(emb[i], emb[j+3]))
		}
	}

	avgWithin := mean(within)
	avgCross := mean(cross)

	t.Logf("avg within-group: %.4f", avgWithin)
	t.Logf("avg cross-group:  %.4f", avgCross)
	t.Logf("separation gap:   %.4f", avgWithin-avgCross)

	if avgCross >= avgWithin {
		t.Errorf("cross-group similarity (%.4f) should be lower than within-group (%.4f)", avgCross, avgWithin)
	}
}

func TestEmbedding_FullMatrix(t *testing.T) {
	client := loadEnvAndClient(t)

	texts := []string{
		// Group 1: imposter syndrome [0..2]
		similarGroup1[0], similarGroup1[1], similarGroup1[2],
		// Group 2: photography [3..5]
		similarGroup2[0], similarGroup2[1], similarGroup2[2],
		// Dissimilar pair 1: keyword collision [6..7]
		dissimilarPair1[0], dissimilarPair1[1],
		// Dissimilar pair 2: trivial vs deep [8..9]
		dissimilarPair2[0], dissimilarPair2[1],
	}
	labels := []string{
		"入职焦虑A", "论文心虚B", "逃避新人文档C",
		"人像调色D", "照片清冷感E", "摄影自我映射F",
		"Go context(G)", "人际关系信号(H)",
		"买伞琐事(I)", "旅行掌控感(J)",
	}

	emb := embedAll(t, client, texts)

	// Print full 10x10 similarity matrix.
	t.Log("\nFull similarity matrix (10×10):\n" + matrixString(emb, labels))

	// Quick structural check: within-group max should exceed cross-group max.
	withinMax := max6(
		sim(emb[0], emb[1]), sim(emb[0], emb[2]), sim(emb[1], emb[2]),
		sim(emb[3], emb[4]), sim(emb[3], emb[5]), sim(emb[4], emb[5]),
	)
	crossMax := max9(
		sim(emb[0], emb[3]), sim(emb[0], emb[4]), sim(emb[0], emb[5]),
		sim(emb[1], emb[3]), sim(emb[1], emb[4]), sim(emb[1], emb[5]),
		sim(emb[2], emb[3]), sim(emb[2], emb[4]), sim(emb[2], emb[5]),
	)

	t.Logf("max within-group: %.4f", withinMax)
	t.Logf("max cross-group:  %.4f", crossMax)

	// Keyword-collision check.
	kwSim := sim(emb[6], emb[7])
	t.Logf("keyword-collision sim (G↔H): %.4f", kwSim)
	if kwSim > crossMax {
		t.Errorf("keyword-collision sim (%.4f) should not exceed cross-group max (%.4f)", kwSim, crossMax)
	}

	// Trivial-vs-deep check.
	tvSim := sim(emb[8], emb[9])
	t.Logf("trivial-vs-deep sim (I↔J):  %.4f", tvSim)
}

// =========================================================================
// Display helpers
// =========================================================================

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}

func mean(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var sum float64
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func max6(a, b, c, d, e, f float64) float64 {
	return max(max(max(a, b), max(c, d)), max(e, f))
}

func max9(a, b, c, d, e, f, g, h, i float64) float64 {
	return max(max6(a, b, c, d, e, f), max6(g, h, i, 0, 0, 0))
}

func matrixString(emb [][]float32, labels []string) string {
	n := len(emb)
	var b strings.Builder

	b.WriteString("     ")
	for i := range n {
		fmt.Fprintf(&b, " %6s", labels[i])
	}
	b.WriteByte('\n')
	for i := range n {
		fmt.Fprintf(&b, " %4s", labels[i])
		for j := range n {
			if i == j {
				b.WriteString("   ·   ")
			} else {
				fmt.Fprintf(&b, " %6.4f", sim(emb[i], emb[j]))
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}
