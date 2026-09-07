// Package analytics contains transparent, versioned derived measures.
package analytics

import (
	"fmt"
	"math"
)

// RiskVersion identifies the weighting and threshold contract used by ScoreRisk.
const RiskVersion = "risk/v1"

// RiskFactor identifies a normalized [0,100] input, where larger is worse.
type RiskFactor string

const (
	Complexity         RiskFactor = "complexity"
	Churn              RiskFactor = "churn"
	ChangeFrequency    RiskFactor = "change_frequency"
	DependencyExposure RiskFactor = "dependency_exposure"
	Ownership          RiskFactor = "ownership"
	CoverageGap        RiskFactor = "coverage_gap"
	ConfirmedBugs      RiskFactor = "confirmed_bugs"
)

// RiskTier classifies a sufficient aggregate risk score.
type RiskTier string

const (
	RiskInsufficient RiskTier = "insufficient"
	RiskLow          RiskTier = "low"
	RiskMedium       RiskTier = "medium"
	RiskHigh         RiskTier = "high"
)

// FactorBreakdown explains one factor's knownness, weight, and score contribution.
type FactorBreakdown struct {
	// Factor identifies the normalized input.
	Factor RiskFactor `json:"factor"`
	// Known reports whether Value was supplied.
	Known bool `json:"known"`
	// Value is nil for an unknown factor; known values are in [0,100].
	Value *float64 `json:"value"`
	// Weight is the factor's base weight from zero to one.
	Weight float64 `json:"weight"`
	// EffectiveWeight is base weight renormalized across known factors.
	EffectiveWeight float64 `json:"effectiveWeight"`
	// Contribution is Value multiplied by EffectiveWeight.
	Contribution float64 `json:"contribution"`
}

// RiskResult is a versioned aggregate whose score is absent when evidence is insufficient.
type RiskResult struct {
	// Version identifies the scoring contract.
	Version string `json:"version"`
	// Score is nil until enough weighted inputs are known.
	Score *float64 `json:"score"`
	// Tier is insufficient when Score is nil.
	Tier RiskTier `json:"tier"`
	// Sufficient reports whether at least 70% base weight is known.
	Sufficient bool `json:"sufficient"`
	// KnownWeight is the known base-weight fraction from zero to one.
	KnownWeight float64 `json:"knownWeight"`
	// Breakdown makes the score and missing inputs auditable.
	Breakdown []FactorBreakdown `json:"breakdown"`
}

// ScoreRisk accepts normalized risk values in [0,100], where larger is worse.
// Absent factors are unknown, never zero. Ownership means ownership risk, not
// owner count. Upstream producers must normalize their raw measures explicitly.
// Scores require >=70% known base weight; known weights are renormalized.
func ScoreRisk(values map[RiskFactor]float64) (RiskResult, error) {
	factors := []struct {
		factor RiskFactor
		weight int
	}{{Complexity, 20}, {Churn, 15}, {ChangeFrequency, 15}, {DependencyExposure, 15}, {Ownership, 10}, {CoverageGap, 20}, {ConfirmedBugs, 5}}
	knownFactors := map[RiskFactor]bool{}
	for _, f := range factors {
		knownFactors[f.factor] = true
	}
	for f, v := range values {
		if !knownFactors[f] {
			return RiskResult{}, fmt.Errorf("unknown risk factor %q", f)
		}
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 100 {
			return RiskResult{}, fmt.Errorf("risk factor %q must be finite in [0,100]", f)
		}
	}
	known := 0
	for _, f := range factors {
		if _, ok := values[f.factor]; ok {
			known += f.weight
		}
	}
	out := RiskResult{Version: RiskVersion, Tier: RiskInsufficient, KnownWeight: float64(known) / 100, Sufficient: known >= 70, Breakdown: []FactorBreakdown{}}
	score := 0.0
	for _, f := range factors {
		b := FactorBreakdown{Factor: f.factor, Weight: float64(f.weight) / 100}
		if v, ok := values[f.factor]; ok {
			b.Known = true
			b.Value = &v
			b.EffectiveWeight = float64(f.weight) / float64(known)
			b.Contribution = v * b.EffectiveWeight
			score += b.Contribution
		}
		out.Breakdown = append(out.Breakdown, b)
	}
	if out.Sufficient {
		out.Score = &score
		switch {
		case score < 34:
			out.Tier = RiskLow
		case score < 67:
			out.Tier = RiskMedium
		default:
			out.Tier = RiskHigh
		}
	}
	return out, nil
}
