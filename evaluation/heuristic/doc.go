// Package heuristic provides rule-based evaluation metrics that don't require LLM calls.
//
// # String Metrics
//
// Basic string comparison metrics:
//   - Equals: Exact string match
//   - Contains: Substring presence
//   - StartsWith, EndsWith: Prefix/suffix matching
//   - ContainsAny, ContainsAll: Multiple value matching
//   - NotEmpty: Non-empty output check
//   - LengthBetween, WordCount: Length constraints
//
// # Parsing Metrics
//
// Format validation metrics:
//   - IsJSON, IsJSONObject, IsJSONArray: JSON validation
//   - JSONHasKeys, JSONSchemaValid: JSON structure validation
//   - IsXML: XML validation
//   - IsNumber, IsBoolean: Type validation
//
// # Pattern Metrics
//
// Regular expression and format validation:
//   - RegexMatch, RegexNotMatch: Pattern matching
//   - EmailFormat, URLFormat: Common formats
//   - PhoneFormat, DateFormat, UUIDFormat: Specialized formats
//
// # Similarity Metrics
//
// Text similarity metrics:
//   - LevenshteinSimilarity: Edit distance based
//   - JaccardSimilarity: Set-based overlap
//   - CosineSimilarity: Word vector similarity
//   - BLEU: N-gram precision (machine translation style)
//   - ROUGE: Longest common subsequence
//   - FuzzyMatch: Combined similarity score
//
// # Usage Example
//
//	metrics := []evaluation.Metric{
//	    heuristic.NewEquals(false),
//	    heuristic.NewContains(false),
//	    heuristic.NewIsJSON(),
//	    heuristic.NewLevenshteinSimilarity(false),
//	}
//
//	engine := evaluation.NewEngine(metrics)
//	input := evaluation.NewMetricInput("prompt", "response").WithExpected("expected")
//	result := engine.EvaluateOne(ctx, input)
package heuristic
