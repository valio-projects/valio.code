package analytics

import (
	"math"
	"testing"
)

func TestRiskMissingAndRenormalized(t *testing.T) {
	empty, err := ScoreRisk(nil)
	if err != nil {
		t.Fatal(err)
	}
	if empty.Score != nil || empty.Sufficient || empty.Tier != RiskInsufficient {
		t.Fatal("missing risk treated as zero")
	}
	partial, err := ScoreRisk(map[RiskFactor]float64{Complexity: 100, CoverageGap: 50, Churn: 0, ChangeFrequency: 0})
	if err != nil {
		t.Fatal(err)
	}
	if !partial.Sufficient || partial.KnownWeight != 0.7 {
		t.Fatalf("threshold failed %+v", partial)
	}
	want := 30.0 / 0.7
	if math.Abs(*partial.Score-want) > 1e-10 {
		t.Fatalf("got %v want %v", *partial.Score, want)
	}
	if partial.Tier != RiskMedium {
		t.Fatal("wrong tier")
	}
	missing, err := ScoreRisk(map[RiskFactor]float64{Complexity: 100, CoverageGap: 100, Churn: 100, Ownership: 100})
	if err != nil {
		t.Fatal(err)
	}
	if missing.Score != nil || missing.KnownWeight != 0.65 {
		t.Fatal("insufficient inputs should not yield score")
	}
}
func TestRiskTierBoundariesAndValidation(t *testing.T) {
	for _, tt := range []struct {
		value float64
		tier  RiskTier
	}{{0, RiskLow}, {33.99, RiskLow}, {34, RiskMedium}, {66.99, RiskMedium}, {67, RiskHigh}, {100, RiskHigh}} {
		values := map[RiskFactor]float64{}
		for _, f := range []RiskFactor{Complexity, Churn, ChangeFrequency, DependencyExposure, Ownership, CoverageGap, ConfirmedBugs} {
			values[f] = tt.value
		}
		got, err := ScoreRisk(values)
		if err != nil {
			t.Fatal(err)
		}
		if got.Tier != tt.tier {
			t.Errorf("%v got %s", tt.value, got.Tier)
		}
	}
	for _, v := range []float64{-1, 101, math.NaN(), math.Inf(1)} {
		if _, err := ScoreRisk(map[RiskFactor]float64{Complexity: v}); err == nil {
			t.Fatal("accepted invalid risk")
		}
	}
	if _, err := ScoreRisk(map[RiskFactor]float64{"mystery": 50}); err == nil {
		t.Fatal("accepted unknown factor")
	}
}
