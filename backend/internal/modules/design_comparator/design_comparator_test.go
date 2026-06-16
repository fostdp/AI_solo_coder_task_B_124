package design_comparator

import (
	"testing"

	"lingqu-dou-gate/internal/config"
	"lingqu-dou-gate/internal/models"
	"lingqu-dou-gate/internal/modules/hydraulic_sim"
)

func makeTestGate() models.DouGate {
	return models.DouGate{
		ID:               1,
		Name:             "测试陡门",
		GateWidth:        6.0,
		GateHeight:       4.5,
		ChamberLength:    60.0,
		ChamberWidth:     7.0,
		MaxWaterLevelUp:  7.5,
		MinWaterLevelDown: 3.5,
		WeirCoeff:        1.62,
	}
}

func makeTestComparator() *DesignComparator {
	config.AppConfig.HydraulicJSON = config.DefaultHydraulicJSONConfig()
	hydro := hydraulic_sim.NewHydraulicSimulator(1)
	dc := NewDesignComparator(hydro)
	return dc
}

func TestNewDesignComparator(t *testing.T) {
	dc := makeTestComparator()
	if dc == nil {
		t.Fatal("DesignComparator should not be nil")
	}
}

func TestCompare_ReturnsAllDynasties(t *testing.T) {
	dc := makeTestComparator()
	gate := makeTestGate()
	results := dc.Compare(gate, 7.5, 3.5, 0.85, "upstream", nil)

	if len(results) != 4 {
		t.Errorf("expected 4 dynasty results, got %d", len(results))
	}
	dynasties := map[models.Dynasty]bool{}
	for _, r := range results {
		dynasties[r.Dynasty] = true
		if r.Design == nil {
			t.Errorf("dynasty %s: design should not be nil", r.Dynasty)
		}
		if r.Simulation == nil {
			t.Errorf("dynasty %s: simulation should not be nil", r.Dynasty)
		}
		if r.EfficiencyScore <= 0 {
			t.Errorf("dynasty %s: efficiency score should be positive, got %f", r.Dynasty, r.EfficiencyScore)
		}
		if r.PassagesPerDay <= 0 {
			t.Errorf("dynasty %s: passages per day should be positive", r.Dynasty)
		}
	}
	if !dynasties[models.DynastyTang] || !dynasties[models.DynastySong] ||
		!dynasties[models.DynastyQing] || !dynasties[models.DynastyModern] {
		t.Error("missing expected dynasties")
	}
}

func TestCompare_WithFilter(t *testing.T) {
	dc := makeTestComparator()
	gate := makeTestGate()
	filter := []models.Dynasty{models.DynastyTang, models.DynastySong}
	results := dc.Compare(gate, 7.5, 3.5, 0.85, "upstream", filter)

	if len(results) != 2 {
		t.Errorf("expected 2 filtered results, got %d", len(results))
	}
}

func TestCompare_EmptyFilter(t *testing.T) {
	dc := makeTestComparator()
	gate := makeTestGate()
	results := dc.Compare(gate, 7.5, 3.5, 0.85, "upstream", []models.Dynasty{})

	if len(results) != 4 {
		t.Errorf("empty filter should return all 4 dynasties, got %d", len(results))
	}
}

func TestCompare_Downstream(t *testing.T) {
	dc := makeTestComparator()
	gate := makeTestGate()
	results := dc.Compare(gate, 7.5, 3.5, 0.85, "downstream", nil)

	if len(results) != 4 {
		t.Errorf("downstream should return 4 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Simulation.FillTime <= 0 {
			t.Errorf("downstream fill time should be positive")
		}
	}
}

func TestCompare_AdvantagesAndLimitations(t *testing.T) {
	dc := makeTestComparator()
	gate := makeTestGate()
	results := dc.Compare(gate, 7.5, 3.5, 0.85, "upstream", nil)

	for _, r := range results {
		if len(r.Advantages) == 0 {
			t.Errorf("dynasty %s: should have at least one advantage", r.Dynasty)
		}
	}
}

func TestCompare_ModernEfficiencyHighest(t *testing.T) {
	dc := makeTestComparator()
	gate := makeTestGate()
	results := dc.Compare(gate, 7.5, 3.5, 0.85, "upstream", nil)

	var modernScore, tangScore float64
	for _, r := range results {
		if r.Dynasty == models.DynastyModern {
			modernScore = r.EfficiencyScore
		}
		if r.Dynasty == models.DynastyTang {
			tangScore = r.EfficiencyScore
		}
	}
	if modernScore <= tangScore {
		t.Errorf("modern efficiency %f should be > tang %f", modernScore, tangScore)
	}
}

func TestCompare_BoundarySameWaterLevel(t *testing.T) {
	dc := makeTestComparator()
	gate := makeTestGate()
	results := dc.Compare(gate, 5.0, 5.0, 0.85, "upstream", nil)

	for _, r := range results {
		if r.Simulation.FillTime > 1 {
			t.Errorf("same water level: fill time %f should be ~0", r.Simulation.FillTime)
		}
	}
}

func TestCompare_BoundaryZeroOpening(t *testing.T) {
	dc := makeTestComparator()
	gate := makeTestGate()
	results := dc.Compare(gate, 7.5, 3.5, 0.0, "upstream", nil)

	for _, r := range results {
		if r.Simulation.MaxFlowRate < 0 {
			t.Errorf("zero opening: max flow rate should not be negative")
		}
	}
}

func TestCompare_EfficiencyScoreRange(t *testing.T) {
	dc := makeTestComparator()
	gate := makeTestGate()
	results := dc.Compare(gate, 7.5, 3.5, 0.85, "upstream", nil)

	for _, r := range results {
		if r.EfficiencyScore < 0 || r.EfficiencyScore > 100 {
			t.Errorf("dynasty %s: efficiency score %f out of [0,100]", r.Dynasty, r.EfficiencyScore)
		}
	}
}
