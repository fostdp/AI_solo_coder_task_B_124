package vessel_analyzer

import (
	"testing"

	"lingqu-dou-gate/internal/config"
	"lingqu-dou-gate/internal/models"
	"lingqu-dou-gate/internal/modules/hydraulic_sim"
)

func makeTestGate() models.DouGate {
	return models.DouGate{
		ID:                1,
		Name:              "测试陡门",
		GateWidth:         6.0,
		GateHeight:        4.5,
		ChamberLength:     60.0,
		ChamberWidth:      7.0,
		MaxWaterLevelUp:   7.5,
		MinWaterLevelDown: 3.5,
		WeirCoeff:         1.62,
	}
}

func makeTestAnalyzer() *VesselAnalyzer {
	config.AppConfig.HydraulicJSON = config.DefaultHydraulicJSONConfig()
	hydro := hydraulic_sim.NewHydraulicSimulator(1)
	va := NewVesselAnalyzer(hydro)
	return va
}

func findSpec(st models.ShipType) *models.ShipTypeSpec {
	for _, s := range models.GetShipTypeSpecs() {
		if s.ShipType == st {
			return s
		}
	}
	return nil
}

func TestNewVesselAnalyzer(t *testing.T) {
	va := makeTestAnalyzer()
	if va == nil {
		t.Fatal("VesselAnalyzer should not be nil")
	}
}

func TestAnalyze_AllSevenShipTypes(t *testing.T) {
	va := makeTestAnalyzer()
	gate := makeTestGate()
	reports := va.Analyze(gate, 7.5, 3.5, 0.85, nil)

	if len(reports) != 7 {
		t.Errorf("expected 7 ship type reports, got %d", len(reports))
	}
	typeSet := map[models.ShipType]bool{}
	for _, r := range reports {
		typeSet[r.ShipType] = true
		if r.TypeName == "" {
			t.Errorf("ship type %s: type name should not be empty", r.ShipType)
		}
		if r.EfficiencyIndex < 0 || r.EfficiencyIndex > 100 {
			t.Errorf("ship %s: efficiency index %f out of [0,100]", r.ShipType, r.EfficiencyIndex)
		}
		if r.ConflictRate < 0 || r.ConflictRate > 1 {
			t.Errorf("ship %s: conflict rate %f out of [0,1]", r.ShipType, r.ConflictRate)
		}
	}
	expected := []models.ShipType{
		models.ShipTypeGrain, models.ShipTypeCargo, models.ShipTypePassenger,
		models.ShipTypeMilitary, models.ShipTypeFishing, models.ShipTypeTribute, models.ShipTypeRoyal,
	}
	for _, e := range expected {
		if !typeSet[e] {
			t.Errorf("missing expected ship type: %s", e)
		}
	}
}

func TestAnalyze_WithFilter(t *testing.T) {
	va := makeTestAnalyzer()
	gate := makeTestGate()
	filter := []models.ShipType{models.ShipTypeFishing, models.ShipTypeRoyal}
	reports := va.Analyze(gate, 7.5, 3.5, 0.85, filter)

	if len(reports) != 2 {
		t.Errorf("expected 2 filtered results, got %d", len(reports))
	}
}

func TestAnalyze_EmptyFilter(t *testing.T) {
	va := makeTestAnalyzer()
	gate := makeTestGate()
	reports := va.Analyze(gate, 7.5, 3.5, 0.85, []models.ShipType{})

	if len(reports) != 7 {
		t.Errorf("empty filter should return all 7 ship types, got %d", len(reports))
	}
}

func TestAnalyze_WaterFactorEffect(t *testing.T) {
	va := makeTestAnalyzer()
	gate := makeTestGate()
	reports := va.Analyze(gate, 7.5, 3.5, 0.85, nil)

	fishingIdx := 0.0
	royalIdx := 0.0
	for _, r := range reports {
		if r.ShipType == models.ShipTypeFishing {
			fishingIdx = r.EfficiencyIndex
		}
		if r.ShipType == models.ShipTypeRoyal {
			royalIdx = r.EfficiencyIndex
		}
	}
	if fishingIdx < royalIdx {
		t.Errorf("fishing efficiency %f should be >= royal %f (small ship more efficient)", fishingIdx, royalIdx)
	}
}

func TestAnalyze_ConflictRateBySize(t *testing.T) {
	va := makeTestAnalyzer()
	gate := makeTestGate()
	reports := va.Analyze(gate, 7.5, 3.5, 0.85, nil)

	var fishingRate, royalRate float64
	for _, r := range reports {
		if r.ShipType == models.ShipTypeFishing {
			fishingRate = r.ConflictRate
		}
		if r.ShipType == models.ShipTypeRoyal {
			royalRate = r.ConflictRate
		}
	}
	if fishingRate > royalRate {
		t.Errorf("fishing conflict rate %f should be < royal %f", fishingRate, royalRate)
	}
}

func TestAnalyze_PassageTimeBySize(t *testing.T) {
	va := makeTestAnalyzer()
	gate := makeTestGate()
	reports := va.Analyze(gate, 7.5, 3.5, 0.85, nil)

	var fishTime, royalTime float64
	for _, r := range reports {
		if r.ShipType == models.ShipTypeFishing {
			fishTime = r.AvgPassageTimeS
		}
		if r.ShipType == models.ShipTypeRoyal {
			royalTime = r.AvgPassageTimeS
		}
	}
	if fishTime > royalTime {
		t.Errorf("fishing passage time %.0f should be < royal %.0f", fishTime, royalTime)
	}
}

func TestAnalyze_BoundarySameWaterLevel(t *testing.T) {
	va := makeTestAnalyzer()
	gate := makeTestGate()
	reports := va.Analyze(gate, 5.0, 5.0, 0.85, nil)

	for _, r := range reports {
		if r.AvgWaterPerTon > 0.1 {
			t.Errorf("ship %s: same water level should use minimal water, got %f", r.ShipType, r.AvgWaterPerTon)
		}
	}
}

func TestAnalyze_BoundaryZeroOpening(t *testing.T) {
	va := makeTestAnalyzer()
	gate := makeTestGate()
	reports := va.Analyze(gate, 7.5, 3.5, 0.0, nil)

	for _, r := range reports {
		if r.AvgThroughputTon < 0 {
			t.Errorf("ship %s: throughput should not be negative at zero opening", r.ShipType)
		}
	}
}

func TestAnalyze_BoundaryFullOpening(t *testing.T) {
	va := makeTestAnalyzer()
	gate := makeTestGate()
	reports := va.Analyze(gate, 7.5, 3.5, 1.0, nil)

	for _, r := range reports {
		if r.AvgPassageTimeS <= 0 {
			t.Errorf("ship %s: passage time should be positive at full opening", r.ShipType)
		}
	}
}

func TestAnalyze_AbnormalNegativeWater(t *testing.T) {
	va := makeTestAnalyzer()
	gate := makeTestGate()
	reports := va.Analyze(gate, -1.0, -2.0, 0.85, nil)

	if len(reports) != 7 {
		t.Errorf("should still return 7 results with abnormal water levels, got %d", len(reports))
	}
}

func TestAnalyze_AbnormalExtremeHead(t *testing.T) {
	va := makeTestAnalyzer()
	gate := makeTestGate()
	reports := va.Analyze(gate, 100.0, 0.0, 0.85, nil)

	for _, r := range reports {
		if r.EfficiencyIndex > 100 {
			t.Errorf("ship %s: efficiency index %f should not exceed 100", r.ShipType, r.EfficiencyIndex)
		}
	}
}

func TestAnalyze_AbnormalInvalidFilter(t *testing.T) {
	va := makeTestAnalyzer()
	gate := makeTestGate()
	invalidType := []models.ShipType{models.ShipType("nonexistent")}
	reports := va.Analyze(gate, 7.5, 3.5, 0.85, invalidType)

	if len(reports) != 0 {
		t.Errorf("invalid filter should return empty results, got %d", len(reports))
	}
}

func TestAnalyze_ResistancePenaltyApplied(t *testing.T) {
	va := makeTestAnalyzer()
	gate := makeTestGate()
	reports := va.Analyze(gate, 7.5, 3.5, 0.85, nil)

	for _, r := range reports {
		spec := findSpec(r.ShipType)
		if spec == nil {
			continue
		}
		occupancy := spec.ChamberOccupancy
		if occupancy <= 0 {
			occupancy = spec.LengthMax * spec.WidthMax / (gate.ChamberLength * gate.ChamberWidth)
		}
		resistancePenalty := 1.0 + spec.ResistanceCoeff*occupancy*10
		if r.EfficiencyIndex > 0 {
			upperBound := 100.0 / resistancePenalty
			if r.EfficiencyIndex > upperBound+5 {
				t.Errorf("ship %s: effIdx %f exceeds bound %f (penalty=%f)",
					r.TypeName, r.EfficiencyIndex, upperBound, resistancePenalty)
			}
		}
	}
}

func TestAnalyze_ConflictRateContinuousModel(t *testing.T) {
	va := makeTestAnalyzer()
	gate := makeTestGate()
	reports := va.Analyze(gate, 7.5, 3.5, 0.85, nil)

	var lowOccShip, highOccShip *models.ShipTypeEfficiencyReport
	for _, r := range reports {
		spec := findSpec(r.ShipType)
		if spec == nil {
			continue
		}
		if spec.ChamberOccupancy < 0.1 && lowOccShip == nil {
			lowOccShip = r
		}
		if spec.ChamberOccupancy > 0.4 && highOccShip == nil {
			highOccShip = r
		}
	}
	if lowOccShip != nil && highOccShip != nil {
		if lowOccShip.ConflictRate > highOccShip.ConflictRate+0.05 {
			t.Errorf("low occupancy (%s, %.3f) should not exceed high occupancy (%s, %.3f)",
				lowOccShip.TypeName, lowOccShip.ConflictRate, highOccShip.TypeName, highOccShip.ConflictRate)
		}
	}
}
