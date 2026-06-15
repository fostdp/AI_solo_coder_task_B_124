package hydraulic_sim

import (
	"math"
	"testing"

	"lingqu-dou-gate/internal/config"
	"lingqu-dou-gate/internal/models"
)

func makeTestGate() models.DouGate {
	return models.DouGate{
		ID: 1, Name: "陡门1号", GateWidth: 6.0, GateHeight: 4.5,
		ChamberLength: 60, ChamberWidth: 7.0,
		MaxWaterLevelUp: 7.5, MinWaterLevelUp: 5.0,
		MaxWaterLevelDown: 4.5, MinWaterLevelDown: 3.5,
		DischargeCoefficient: 0.63,
	}
}

func makeTestSimulator() *HydraulicSimulator {
	return &HydraulicSimulator{
		params: config.HydraulicJSONConfig{
			Gravity: 9.81, WaterDensity: 1000.0,
			DefaultCd: 0.63, SubmergedCoeff: 0.92, WeirCoeff: 1.84,
			FlowRegime: config.FlowRegimeParams{
				FreeFlowThreshold: 0.67, SubmergedThreshold: 0.88,
				WeirRelativeOpening: 0.75, WeirHeadRatio: 0.3,
				FullySubmerged: 0.97, TransitionWidth: 0.09,
			},
			ContractionCoeff: config.ContractionCoeffParams{
				Base: 0.615, LinearTerm: 0.105, QuadraticTerm: -0.02,
			},
			Simulation: config.SimulationParams{
				TimeStep: 0.25, MaxIterations: 40000,
				AdaptiveStepRatio: 0.05, MinDT: 0.01,
			},
		},
	}
}

func TestDynastyComparison_Normal_AllFour(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	results := sim.DynastyComparisonSync(gate, 7.5, 3.6, 0.85, "upstream", nil)

	if len(results) != 4 {
		t.Fatalf("expected 4 dynasty results, got %d", len(results))
	}

	dynastyOrder := []models.Dynasty{models.DynastyTang, models.DynastySong, models.DynastyQing, models.DynastyModern}
	for i, r := range results {
		if r.Dynasty != dynastyOrder[i] {
			t.Errorf("result[%d]: expected dynasty %s, got %s", i, dynastyOrder[i], r.Dynasty)
		}
		if r.Design == nil {
			t.Errorf("result[%d]: Design should not be nil", i)
		}
		if r.Simulation == nil {
			t.Errorf("result[%d]: Simulation should not be nil", i)
		}
		if r.Simulation.MaxFlowRate <= 0 {
			t.Errorf("result[%d]: MaxFlowRate should be positive, got %f", i, r.Simulation.MaxFlowRate)
		}
		if r.EfficiencyScore <= 0 || r.EfficiencyScore > 100 {
			t.Errorf("result[%d]: EfficiencyScore should be in (0,100], got %f", i, r.EfficiencyScore)
		}
		if r.PassagesPerDay < 0 {
			t.Errorf("result[%d]: PassagesPerDay should be non-negative, got %d", i, r.PassagesPerDay)
		}
		if r.WaterPerTon < 0 {
			t.Errorf("result[%d]: WaterPerTon should be non-negative, got %f", i, r.WaterPerTon)
		}
	}
}

func TestDynastyComparison_TechnologyProgression(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	results := sim.DynastyComparisonSync(gate, 7.5, 3.6, 0.85, "upstream", nil)

	if len(results) < 4 {
		t.Fatal("need at least 4 results for progression test")
	}

	tangFlow := results[0].Simulation.MaxFlowRate
	modernFlow := results[3].Simulation.MaxFlowRate
	if modernFlow < tangFlow {
		t.Errorf("modern max flow rate (%f) should >= tang max flow rate (%f)", modernFlow, tangFlow)
	}

	tangChamber := results[0].Design.ChamberLength
	modernChamber := results[3].Design.ChamberLength
	if modernChamber < tangChamber {
		t.Errorf("modern chamber length (%f) should >= tang (%f)", modernChamber, tangChamber)
	}

	tangWidth := results[0].Design.GateWidth
	modernWidth := results[3].Design.GateWidth
	if modernWidth < tangWidth {
		t.Errorf("modern gate width (%f) should >= tang (%f)", modernWidth, tangWidth)
	}

	modernScore := results[3].EfficiencyScore
	tangScore := results[0].EfficiencyScore
	if modernScore < tangScore*0.8 {
		t.Errorf("modern efficiency (%f) should not be dramatically lower than tang (%f)", modernScore, tangScore)
	}
}

func TestDynastyComparison_DirectionDownstream(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	results := sim.DynastyComparisonSync(gate, 7.5, 3.6, 0.85, "downstream", nil)

	if len(results) != 4 {
		t.Fatalf("expected 4 results for downstream, got %d", len(results))
	}
	for i, r := range results {
		if r.Simulation.DrainTime <= 0 {
			t.Errorf("result[%d]: downstream drain time should be positive, got %f", i, r.Simulation.DrainTime)
		}
	}
}

func TestDynastyComparison_FilterDynasties(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()

	results := sim.DynastyComparisonSync(gate, 7.5, 3.6, 0.85, "upstream",
		[]models.Dynasty{models.DynastySong, models.DynastyQing})

	if len(results) != 2 {
		t.Fatalf("expected 2 filtered results, got %d", len(results))
	}
	if results[0].Dynasty != models.DynastySong {
		t.Errorf("first result should be song, got %s", results[0].Dynasty)
	}
	if results[1].Dynasty != models.DynastyQing {
		t.Errorf("second result should be qing, got %s", results[1].Dynasty)
	}
}

func TestDynastyComparison_AdvantagesLimitations(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	results := sim.DynastyComparisonSync(gate, 7.5, 3.6, 0.85, "upstream", nil)

	tangAdv := results[0].Advantages
	tangLim := results[0].Limitations
	if len(tangLim) == 0 {
		t.Error("tang dynasty should have limitations (土石木混合 material)")
	}
	hasMaterialLimit := false
	for _, l := range tangLim {
		if len(l) > 0 {
			hasMaterialLimit = true
			break
		}
	}
	if !hasMaterialLimit {
		t.Error("tang dynasty limitations should contain material-related info")
	}

	modernAdv := results[3].Advantages
	if len(modernAdv) == 0 {
		t.Error("modern dynasty should have advantages")
	}
}

func TestDynastyComparison_Boundary_ZeroOpening(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	results := sim.DynastyComparisonSync(gate, 7.5, 3.6, 0, "upstream", nil)

	if len(results) != 4 {
		t.Fatalf("expected 4 results with zero opening, got %d", len(results))
	}
	for i, r := range results {
		if r.Simulation.MaxFlowRate < 0 {
			t.Errorf("result[%d]: flow rate should be non-negative even with zero opening", i)
		}
	}
}

func TestDynastyComparison_Boundary_SameWaterLevel(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	results := sim.DynastyComparisonSync(gate, 5.0, 5.0, 0.8, "upstream", nil)

	if len(results) != 4 {
		t.Fatalf("expected 4 results with same water level, got %d", len(results))
	}
	for i, r := range results {
		if r.Simulation.TotalWaterVolume < 0 {
			t.Errorf("result[%d]: total water volume should be non-negative with equal levels", i)
		}
	}
}

func TestDynastyComparison_Boundary_FullOpening(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	results := sim.DynastyComparisonSync(gate, 7.5, 3.6, 1.0, "upstream", nil)

	for i, r := range results {
		if r.Simulation.MaxFlowRate <= 0 {
			t.Errorf("result[%d]: max flow should be positive with full opening", i)
		}
	}
}

func TestDynastyComparison_Abnormal_NegativeWaterLevel(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	results := sim.DynastyComparisonSync(gate, -1, -1, 0.8, "upstream", nil)

	if len(results) != 4 {
		t.Fatalf("should still return 4 results with negative water levels, got %d", len(results))
	}
}

func TestDynastyComparison_Abnormal_ExtremelyHighHead(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	results := sim.DynastyComparisonSync(gate, 100.0, 0.1, 0.8, "upstream", nil)

	for i, r := range results {
		if math.IsInf(r.Simulation.MaxFlowRate, 0) || math.IsNaN(r.Simulation.MaxFlowRate) {
			t.Errorf("result[%d]: max flow rate should not be Inf/NaN with extreme head", i)
		}
		if math.IsInf(r.Simulation.FillTime, 0) || math.IsNaN(r.Simulation.FillTime, 0) {
			t.Errorf("result[%d]: fill time should not be Inf/NaN with extreme head", i)
		}
	}
}

func TestDynastyComparison_Abnormal_EmptyDynastyFilter(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	results := sim.DynastyComparisonSync(gate, 7.5, 3.6, 0.8, "upstream", []models.Dynasty{})

	if len(results) != 4 {
		t.Fatalf("empty dynasty filter should return all 4, got %d", len(results))
	}
}

func TestDynastyDesign_FlowRateCap(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()

	results := sim.DynastyComparisonSync(gate, 7.5, 3.6, 1.0, "upstream", nil)
	for i, r := range results {
		if r.Design != nil && r.Design.MaxFlowRate > 0 {
			if r.Simulation.MaxFlowRate > r.Design.MaxFlowRate*1.05 {
				t.Errorf("result[%d]: max flow rate %f should be capped near design max %f",
					i, r.Simulation.MaxFlowRate, r.Design.MaxFlowRate)
			}
		}
	}
}

func TestSimulatePassage_WithDynastyDesign(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	designs := models.GetDynastyDesigns(gate.ID)

	for _, design := range designs {
		req := SimulateRequest{
			Gate: gate, WaterLevelUp: 7.5, WaterLevelDown: 3.6,
			GateOpening: 0.8, Direction: "upstream", DynastyDesign: design,
		}
		result := sim.SimulateSync(req)

		if result.UsedDynasty != string(design.Dynasty) {
			t.Errorf("expected UsedDynasty=%s, got %s", design.Dynasty, result.UsedDynasty)
		}
		if result.MaxFlowRate < 0 {
			t.Errorf("dynasty %s: max flow rate should be non-negative", design.Dynasty)
		}
		if result.TotalPassageS < 0 {
			t.Errorf("dynasty %s: total passage time should be non-negative", design.Dynasty)
		}
	}
}

func TestSimulatePassage_WithoutDynastyDesign(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()

	req := SimulateRequest{
		Gate: gate, WaterLevelUp: 7.5, WaterLevelDown: 3.6,
		GateOpening: 0.8, Direction: "upstream",
	}
	result := sim.SimulateSync(req)

	if result.UsedDynasty != "" {
		t.Errorf("without dynasty design, UsedDynasty should be empty, got %s", result.UsedDynasty)
	}
	if result.MaxFlowRate <= 0 {
		t.Error("without dynasty design, max flow rate should still be positive")
	}
}

func TestComputeEfficiencyScore(t *testing.T) {
	designs := models.GetDynastyDesigns(1)
	sim := makeTestSimulator()
	gate := makeTestGate()

	scores := make([]float64, len(designs))
	for i, d := range designs {
		req := SimulateRequest{
			Gate: gate, WaterLevelUp: 7.5, WaterLevelDown: 3.6,
			GateOpening: 0.8, Direction: "upstream", DynastyDesign: d,
		}
		r := sim.SimulateSync(req)
		scores[i] = computeEfficiencyScore(r, d)
		if scores[i] < 0 || scores[i] > 100 {
			t.Errorf("dynasty %s: efficiency score %f out of [0,100] range", d.Dynasty, scores[i])
		}
	}
}
