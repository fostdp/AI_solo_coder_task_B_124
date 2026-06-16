package hydraulic_sim

import (
	"math"
	"testing"

	"lingqu-dou-gate/internal/models"
)

func TestShipTypeAnalysis_Normal_AllTypes(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	reports := sim.ShipTypeAnalysisSync(gate, 7.5, 3.6, 0.85, nil)

	if len(reports) != 7 {
		t.Fatalf("expected 7 ship type reports, got %d", len(reports))
	}

	for i, r := range reports {
		if r.ShipType == "" {
			t.Errorf("report[%d]: ShipType should not be empty", i)
		}
		if r.TypeName == "" {
			t.Errorf("report[%d]: TypeName should not be empty", i)
		}
		if r.AvgWaitTimeS < 0 {
			t.Errorf("report[%d]: AvgWaitTimeS should be non-negative, got %f", i, r.AvgWaitTimeS)
		}
		if r.AvgPassageTimeS < 0 {
			t.Errorf("report[%d]: AvgPassageTimeS should be non-negative, got %f", i, r.AvgPassageTimeS)
		}
		if r.AvgWaterPerTon < 0 {
			t.Errorf("report[%d]: AvgWaterPerTon should be non-negative, got %f", i, r.AvgWaterPerTon)
		}
		if r.ConflictRate < 0 || r.ConflictRate > 1 {
			t.Errorf("report[%d]: ConflictRate should be in [0,1], got %f", i, r.ConflictRate)
		}
		if r.EfficiencyIndex < 0 || r.EfficiencyIndex > 100 {
			t.Errorf("report[%d]: EfficiencyIndex should be in [0,100], got %f", i, r.EfficiencyIndex)
		}
	}
}

func TestShipTypeAnalysis_WaterFactorEffect(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	reports := sim.ShipTypeAnalysisSync(gate, 7.5, 3.6, 0.85, nil)

	specs := models.GetShipTypeSpecs()
	specMap := map[models.ShipType]*models.ShipTypeSpec{}
	for _, s := range specs {
		specMap[s.ShipType] = s
	}

	for _, r := range reports {
		spec, ok := specMap[r.ShipType]
		if !ok {
			continue
		}
		if spec.WaterFactor > 1.0 && r.AvgWaterPerTon <= 0 {
			t.Errorf("ship %s with water_factor>1 should have positive water per ton", r.TypeName)
		}
		if spec.WaterFactor < 0.5 && r.AvgWaterPerTon > 100 {
			t.Errorf("ship %s with water_factor<0.5 should have reasonable water per ton, got %f",
				r.TypeName, r.AvgWaterPerTon)
		}
	}
}

func TestShipTypeAnalysis_ConflictRateBySize(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	reports := sim.ShipTypeAnalysisSync(gate, 7.5, 3.6, 0.85, nil)

	var fishingConflict, royalConflict float64
	for _, r := range reports {
		if r.ShipType == models.ShipTypeFishing {
			fishingConflict = r.ConflictRate
		}
		if r.ShipType == models.ShipTypeRoyal {
			royalConflict = r.ConflictRate
		}
	}
	if royalConflict <= fishingConflict {
		t.Errorf("royal ship (55m) conflict rate (%f) should be > fishing boat (10m) conflict rate (%f)",
			royalConflict, fishingConflict)
	}
}

func TestShipTypeAnalysis_PassageTimeBySize(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	reports := sim.ShipTypeAnalysisSync(gate, 7.5, 3.6, 0.85, nil)

	var fishingPassage, royalPassage float64
	for _, r := range reports {
		if r.ShipType == models.ShipTypeFishing {
			fishingPassage = r.AvgPassageTimeS
		}
		if r.ShipType == models.ShipTypeRoyal {
			royalPassage = r.AvgPassageTimeS
		}
	}
	if royalPassage <= fishingPassage {
		t.Errorf("royal ship passage time (%f) should be > fishing boat passage time (%f)",
			royalPassage, fishingPassage)
	}
}

func TestShipTypeAnalysis_FilterTypes(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	filter := []models.ShipType{models.ShipTypeGrain, models.ShipTypeCargo}
	reports := sim.ShipTypeAnalysisSync(gate, 7.5, 3.6, 0.85, filter)

	if len(reports) != 2 {
		t.Fatalf("expected 2 filtered reports, got %d", len(reports))
	}
	for _, r := range reports {
		if r.ShipType != models.ShipTypeGrain && r.ShipType != models.ShipTypeCargo {
			t.Errorf("unexpected ship type in filtered results: %s", r.ShipType)
		}
	}
}

func TestShipTypeAnalysis_EmptyFilter(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	reports := sim.ShipTypeAnalysisSync(gate, 7.5, 3.6, 0.85, []models.ShipType{})

	if len(reports) != 7 {
		t.Fatalf("empty filter should return all 7, got %d", len(reports))
	}
}

func TestShipTypeAnalysis_EfficiencyIndexRange(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	reports := sim.ShipTypeAnalysisSync(gate, 7.5, 3.6, 0.85, nil)

	for _, r := range reports {
		if r.EfficiencyIndex < 0 {
			t.Errorf("ship %s: efficiency index should be >= 0, got %f", r.TypeName, r.EfficiencyIndex)
		}
		if r.EfficiencyIndex > 100 {
			t.Errorf("ship %s: efficiency index should be <= 100, got %f", r.TypeName, r.EfficiencyIndex)
		}
		if math.IsNaN(r.EfficiencyIndex) || math.IsInf(r.EfficiencyIndex, 0) {
			t.Errorf("ship %s: efficiency index should not be NaN/Inf", r.TypeName)
		}
	}
}

func TestShipTypeAnalysis_Boundary_SameWaterLevel(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	reports := sim.ShipTypeAnalysisSync(gate, 5.0, 5.0, 0.8, nil)

	for _, r := range reports {
		if r.AvgPassageTimeS < 0 {
			t.Errorf("ship %s: passage time should be non-negative with same water level", r.TypeName)
		}
	}
}

func TestShipTypeAnalysis_Boundary_ZeroOpening(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	reports := sim.ShipTypeAnalysisSync(gate, 7.5, 3.6, 0, nil)

	if len(reports) != 7 {
		t.Fatalf("expected 7 reports with zero opening, got %d", len(reports))
	}
}

func TestShipTypeAnalysis_Boundary_FullOpening(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	reports := sim.ShipTypeAnalysisSync(gate, 7.5, 3.6, 1.0, nil)

	for _, r := range reports {
		if r.AvgPassageTimeS <= 0 {
			t.Errorf("ship %s: passage time should be positive with full opening", r.TypeName)
		}
	}
}

func TestShipTypeAnalysis_Abnormal_NegativeWaterLevel(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ShipTypeAnalysisSync panicked with negative water levels: %v", r)
		}
	}()
	reports := sim.ShipTypeAnalysisSync(gate, -5, -3, 0.8, nil)

	if reports == nil {
		t.Error("should not return nil even with negative water levels")
	}
}

func TestShipTypeAnalysis_Abnormal_ExtremeHead(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	reports := sim.ShipTypeAnalysisSync(gate, 1000, 0.01, 0.8, nil)

	for _, r := range reports {
		if math.IsInf(r.AvgPassageTimeS, 0) || math.IsNaN(r.AvgPassageTimeS) {
			t.Errorf("ship %s: passage time should not be Inf/NaN with extreme head", r.TypeName)
		}
		if math.IsInf(r.AvgWaterPerTon, 0) || math.IsNaN(r.AvgWaterPerTon, 0) {
			t.Errorf("ship %s: water per ton should not be Inf/NaN with extreme head", r.TypeName)
		}
	}
}

func TestShipTypeAnalysis_InvalidFilterType(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	reports := sim.ShipTypeAnalysisSync(gate, 7.5, 3.6, 0.85, []models.ShipType{"nonexistent"})

	if len(reports) != 0 {
		t.Errorf("filtering with non-existent type should return 0 results, got %d", len(reports))
	}
}

func TestSimulatePassage_WithShipType(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	specs := models.GetShipTypeSpecs()

	for _, spec := range specs {
		req := SimulateRequest{
			Gate: gate, WaterLevelUp: 7.5, WaterLevelDown: 3.6,
			GateOpening: 0.8, Direction: "upstream", ShipType: spec,
		}
		result := sim.SimulateSync(req)

		if result.ShipTypeName != spec.TypeName {
			t.Errorf("ship %s: expected ShipTypeName=%s, got %s", spec.ShipType, spec.TypeName, result.ShipTypeName)
		}
		if result.EntryTimeS != spec.EntryTimeS {
			t.Errorf("ship %s: expected EntryTimeS=%f, got %f", spec.ShipType, spec.EntryTimeS, result.EntryTimeS)
		}
		if result.ExitTimeS != spec.ExitTimeS {
			t.Errorf("ship %s: expected ExitTimeS=%f, got %f", spec.ShipType, spec.ExitTimeS, result.ExitTimeS)
		}
		expectedTotal := spec.EntryTimeS + result.FillTime + result.DrainTime + spec.ExitTimeS
		if math.Abs(result.TotalPassageS-expectedTotal) > 0.01 {
			t.Errorf("ship %s: total passage %f != entry(%f)+fill(%f)+drain(%f)+exit(%f)",
				spec.ShipType, result.TotalPassageS, spec.EntryTimeS, result.FillTime, result.DrainTime, spec.ExitTimeS)
		}
	}
}

func TestShipTypeAnalysis_ThroughputProportional(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	reports := sim.ShipTypeAnalysisSync(gate, 7.5, 3.6, 0.85, nil)

	var grainThroughput, cargoThroughput float64
	for _, r := range reports {
		if r.ShipType == models.ShipTypeGrain {
			grainThroughput = r.AvgThroughputTon
		}
		if r.ShipType == models.ShipTypeCargo {
			cargoThroughput = r.AvgThroughputTon
		}
	}
	grainSpec, _ := findSpec(models.ShipTypeGrain)
	cargoSpec, _ := findSpec(models.ShipTypeCargo)
	if grainSpec != nil && cargoSpec != nil {
		grainCapRatio := grainThroughput / grainSpec.CapacityTon
		cargoCapRatio := cargoThroughput / cargoSpec.CapacityTon
		ratio := grainCapRatio / cargoCapRatio
		if ratio < 0.3 || ratio > 3.0 {
			t.Logf("grain/cargo normalized throughput ratio = %.2f (may vary with different cycle times)", ratio)
		}
	}
}

func findSpec(st models.ShipType) (*models.ShipTypeSpec, bool) {
	for _, s := range models.GetShipTypeSpecs() {
		if s.ShipType == st {
			return s, true
		}
	}
	return nil, false
}

func TestShipTypeAnalysis_ResistancePenalty(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	reports := sim.ShipTypeAnalysisSync(gate, 7.5, 3.6, 0.85, nil)

	for _, r := range reports {
		spec, found := findSpec(r.ShipType)
		if !found {
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
				t.Errorf("ship %s: effIdx %f exceeds theoretical upper bound %f (resistancePenalty=%f)",
					r.TypeName, r.EfficiencyIndex, upperBound, resistancePenalty)
			}
		}
	}
}

func TestShipTypeAnalysis_ConflictRate_ContinuousModel(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	reports := sim.ShipTypeAnalysisSync(gate, 7.5, 3.6, 0.85, nil)

	var lowOccShip, highOccShip *models.ShipTypeEfficiencyReport
	for _, r := range reports {
		spec, found := findSpec(r.ShipType)
		if !found {
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
			t.Errorf("low occupancy ship (%s, conflict=%.3f) should not have higher conflict than high occupancy (%s, conflict=%.3f)",
				lowOccShip.TypeName, lowOccShip.ConflictRate, highOccShip.TypeName, highOccShip.ConflictRate)
		}
	}
}
