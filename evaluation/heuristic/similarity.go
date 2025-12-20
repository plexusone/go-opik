package heuristic

import (
	"context"
	"math"
	"strings"
	"unicode"

	"github.com/grokify/go-comet-ml-opik/evaluation"
)

// LevenshteinSimilarity calculates similarity based on Levenshtein distance.
type LevenshteinSimilarity struct {
	evaluation.BaseMetric
	caseSensitive bool
}

// NewLevenshteinSimilarity creates a new LevenshteinSimilarity metric.
func NewLevenshteinSimilarity(caseSensitive bool) *LevenshteinSimilarity {
	return &LevenshteinSimilarity{
		BaseMetric:    evaluation.NewBaseMetric("levenshtein_similarity"),
		caseSensitive: caseSensitive,
	}
}

// Score calculates the Levenshtein similarity between output and expected.
func (m *LevenshteinSimilarity) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	s1, s2 := input.Output, input.Expected
	if !m.caseSensitive {
		s1 = strings.ToLower(s1)
		s2 = strings.ToLower(s2)
	}

	distance := levenshteinDistance(s1, s2)
	maxLen := max(len(s1), len(s2))
	if maxLen == 0 {
		return evaluation.NewScoreResult(m.Name(), 1.0) // Both empty
	}

	similarity := 1.0 - float64(distance)/float64(maxLen)
	return evaluation.NewScoreResult(m.Name(), similarity)
}

func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	r1, r2 := []rune(s1), []rune(s2)
	m, n := len(r1), len(r2)

	// Create matrix
	d := make([][]int, m+1)
	for i := range d {
		d[i] = make([]int, n+1)
	}

	// Initialize first row and column
	for i := 0; i <= m; i++ {
		d[i][0] = i
	}
	for j := 0; j <= n; j++ {
		d[0][j] = j
	}

	// Fill in the rest
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}
			d[i][j] = min(
				d[i-1][j]+1,      // deletion
				d[i][j-1]+1,      // insertion
				d[i-1][j-1]+cost, // substitution
			)
		}
	}

	return d[m][n]
}

// JaccardSimilarity calculates the Jaccard similarity coefficient.
type JaccardSimilarity struct {
	evaluation.BaseMetric
	caseSensitive bool
	useWords      bool // true for word-level, false for character-level
}

// NewJaccardSimilarity creates a new JaccardSimilarity metric.
func NewJaccardSimilarity(caseSensitive, useWords bool) *JaccardSimilarity {
	return &JaccardSimilarity{
		BaseMetric:    evaluation.NewBaseMetric("jaccard_similarity"),
		caseSensitive: caseSensitive,
		useWords:      useWords,
	}
}

// Score calculates the Jaccard similarity between output and expected.
func (m *JaccardSimilarity) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	s1, s2 := input.Output, input.Expected
	if !m.caseSensitive {
		s1 = strings.ToLower(s1)
		s2 = strings.ToLower(s2)
	}

	var set1, set2 map[string]bool
	if m.useWords {
		set1, set2 = wordSet(s1), wordSet(s2)
	} else {
		set1, set2 = charSet(s1), charSet(s2)
	}

	if len(set1) == 0 && len(set2) == 0 {
		return evaluation.NewScoreResult(m.Name(), 1.0)
	}

	intersection := 0
	for k := range set1 {
		if set2[k] {
			intersection++
		}
	}

	union := len(set1) + len(set2) - intersection
	similarity := float64(intersection) / float64(union)

	return evaluation.NewScoreResult(m.Name(), similarity)
}

func wordSet(s string) map[string]bool {
	words := strings.Fields(s)
	set := make(map[string]bool, len(words))
	for _, w := range words {
		set[w] = true
	}
	return set
}

func charSet(s string) map[string]bool {
	set := make(map[string]bool, len(s))
	for _, r := range s {
		set[string(r)] = true
	}
	return set
}

// CosineSimilarity calculates word-based cosine similarity.
type CosineSimilarity struct {
	evaluation.BaseMetric
	caseSensitive bool
}

// NewCosineSimilarity creates a new CosineSimilarity metric.
func NewCosineSimilarity(caseSensitive bool) *CosineSimilarity {
	return &CosineSimilarity{
		BaseMetric:    evaluation.NewBaseMetric("cosine_similarity"),
		caseSensitive: caseSensitive,
	}
}

// Score calculates the cosine similarity between output and expected.
func (m *CosineSimilarity) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	s1, s2 := input.Output, input.Expected
	if !m.caseSensitive {
		s1 = strings.ToLower(s1)
		s2 = strings.ToLower(s2)
	}

	vec1 := wordFrequency(s1)
	vec2 := wordFrequency(s2)

	if len(vec1) == 0 || len(vec2) == 0 {
		if len(vec1) == 0 && len(vec2) == 0 {
			return evaluation.NewScoreResult(m.Name(), 1.0)
		}
		return evaluation.NewScoreResult(m.Name(), 0.0)
	}

	// Calculate dot product
	var dotProduct float64
	for word, count1 := range vec1 {
		if count2, ok := vec2[word]; ok {
			dotProduct += float64(count1 * count2)
		}
	}

	// Calculate magnitudes
	var mag1, mag2 float64
	for _, count := range vec1 {
		mag1 += float64(count * count)
	}
	for _, count := range vec2 {
		mag2 += float64(count * count)
	}

	mag1 = math.Sqrt(mag1)
	mag2 = math.Sqrt(mag2)

	if mag1 == 0 || mag2 == 0 {
		return evaluation.NewScoreResult(m.Name(), 0.0)
	}

	similarity := dotProduct / (mag1 * mag2)
	return evaluation.NewScoreResult(m.Name(), similarity)
}

func wordFrequency(s string) map[string]int {
	words := strings.Fields(s)
	freq := make(map[string]int, len(words))
	for _, w := range words {
		// Remove punctuation
		w = strings.TrimFunc(w, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})
		if w != "" {
			freq[w]++
		}
	}
	return freq
}

// BLEU calculates a simplified BLEU (Bilingual Evaluation Understudy) score.
// This is a simplified implementation focusing on n-gram precision.
type BLEU struct {
	evaluation.BaseMetric
	maxN int // maximum n-gram size (typically 4)
}

// NewBLEU creates a new BLEU metric.
func NewBLEU(maxN int) *BLEU {
	if maxN <= 0 {
		maxN = 4
	}
	return &BLEU{
		BaseMetric: evaluation.NewBaseMetric("bleu"),
		maxN:       maxN,
	}
}

// Score calculates the BLEU score between output and expected.
func (m *BLEU) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	candidate := strings.ToLower(input.Output)
	reference := strings.ToLower(input.Expected)

	candWords := strings.Fields(candidate)
	refWords := strings.Fields(reference)

	if len(candWords) == 0 {
		return evaluation.NewScoreResult(m.Name(), 0.0)
	}

	// Calculate brevity penalty
	bp := brevityPenalty(len(candWords), len(refWords))

	// Calculate n-gram precisions
	var logPrecSum float64
	for n := 1; n <= m.maxN; n++ {
		prec := ngramPrecision(candWords, refWords, n)
		if prec > 0 {
			logPrecSum += math.Log(prec)
		} else {
			// Smoothing: use a small value instead of 0
			logPrecSum += math.Log(0.01)
		}
	}

	// Geometric mean of precisions
	avgLogPrec := logPrecSum / float64(m.maxN)
	score := bp * math.Exp(avgLogPrec)

	return evaluation.NewScoreResult(m.Name(), score)
}

func brevityPenalty(candLen, refLen int) float64 {
	if candLen > refLen {
		return 1.0
	}
	return math.Exp(1.0 - float64(refLen)/float64(candLen))
}

func ngramPrecision(candidate, reference []string, n int) float64 {
	if len(candidate) < n || len(reference) < n {
		return 0.0
	}

	candNgrams := getNgrams(candidate, n)
	refNgrams := getNgrams(reference, n)

	// Count matches with clipping
	matches := 0
	for ngram, count := range candNgrams {
		if refCount, ok := refNgrams[ngram]; ok {
			matches += min(count, refCount)
		}
	}

	total := len(candidate) - n + 1
	if total <= 0 {
		return 0.0
	}

	return float64(matches) / float64(total)
}

func getNgrams(words []string, n int) map[string]int {
	ngrams := make(map[string]int)
	for i := 0; i <= len(words)-n; i++ {
		ngram := strings.Join(words[i:i+n], " ")
		ngrams[ngram]++
	}
	return ngrams
}

// ROUGE calculates a simplified ROUGE-L score based on longest common subsequence.
type ROUGE struct {
	evaluation.BaseMetric
	beta float64 // weight for recall vs precision (typically 1.0)
}

// NewROUGE creates a new ROUGE metric.
func NewROUGE(beta float64) *ROUGE {
	if beta <= 0 {
		beta = 1.0
	}
	return &ROUGE{
		BaseMetric: evaluation.NewBaseMetric("rouge_l"),
		beta:       beta,
	}
}

// Score calculates the ROUGE-L score between output and expected.
func (m *ROUGE) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	candidate := strings.ToLower(input.Output)
	reference := strings.ToLower(input.Expected)

	candWords := strings.Fields(candidate)
	refWords := strings.Fields(reference)

	if len(candWords) == 0 || len(refWords) == 0 {
		if len(candWords) == 0 && len(refWords) == 0 {
			return evaluation.NewScoreResult(m.Name(), 1.0)
		}
		return evaluation.NewScoreResult(m.Name(), 0.0)
	}

	lcsLen := lcsLength(candWords, refWords)

	// Calculate precision and recall
	precision := float64(lcsLen) / float64(len(candWords))
	recall := float64(lcsLen) / float64(len(refWords))

	if precision+recall == 0 {
		return evaluation.NewScoreResult(m.Name(), 0.0)
	}

	// F-score with beta weighting
	betaSq := m.beta * m.beta
	fScore := ((1 + betaSq) * precision * recall) / (betaSq*precision + recall)

	return evaluation.NewScoreResult(m.Name(), fScore)
}

func lcsLength(a, b []string) int {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	return dp[m][n]
}

// FuzzyMatch calculates a fuzzy matching score using multiple similarity metrics.
type FuzzyMatch struct {
	evaluation.BaseMetric
	threshold     float64
	caseSensitive bool
}

// NewFuzzyMatch creates a new FuzzyMatch metric.
func NewFuzzyMatch(threshold float64, caseSensitive bool) *FuzzyMatch {
	return &FuzzyMatch{
		BaseMetric:    evaluation.NewBaseMetric("fuzzy_match"),
		threshold:     threshold,
		caseSensitive: caseSensitive,
	}
}

// Score calculates a combined fuzzy match score.
func (m *FuzzyMatch) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	s1, s2 := input.Output, input.Expected
	if !m.caseSensitive {
		s1 = strings.ToLower(s1)
		s2 = strings.ToLower(s2)
	}

	// Combine multiple similarity measures
	levSim := NewLevenshteinSimilarity(true).Score(ctx, evaluation.MetricInput{Output: s1, Expected: s2})
	jacSim := NewJaccardSimilarity(true, true).Score(ctx, evaluation.MetricInput{Output: s1, Expected: s2})

	avgScore := (levSim.Value + jacSim.Value) / 2

	if avgScore >= m.threshold {
		return evaluation.NewScoreResultWithReason(m.Name(), avgScore, "fuzzy match above threshold")
	}
	return evaluation.NewScoreResultWithReason(m.Name(), avgScore, "fuzzy match below threshold")
}

// SemanticSimilarity is a placeholder for embedding-based semantic similarity.
// In a full implementation, this would use an embedding model.
type SemanticSimilarity struct {
	evaluation.BaseMetric
}

// NewSemanticSimilarity creates a new SemanticSimilarity metric.
// Note: This requires an embedding provider to be useful.
func NewSemanticSimilarity() *SemanticSimilarity {
	return &SemanticSimilarity{
		BaseMetric: evaluation.NewBaseMetric("semantic_similarity"),
	}
}

// Score falls back to cosine similarity for now.
// Override this with embedding-based similarity when available.
func (m *SemanticSimilarity) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	// Fallback to word-based cosine similarity
	cosSim := NewCosineSimilarity(false)
	result := cosSim.Score(ctx, input)
	result.Reason = "using word-based approximation (no embedding provider)"
	return &evaluation.ScoreResult{
		Name:   m.Name(),
		Value:  result.Value,
		Reason: result.Reason,
	}
}
