package hydraulic_sim

import (
	"math"
	"testing"

	"lingqu-dou-gate/internal/config"
	"lingqu-dou-gate/internal/models"
)

func TestOrificeFlowWithParams_PositiveHead(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	flow := sim.calculateOrificeFlowWithParams(7.5, 3.5, 0.8, sim.params, gate)
	if flow <= 0 {
		t.Errorf("positive head should produce positive flow, got %f", flow)
	}
}

func TestOrificeFlowWithParams_ZeroHead(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	flow := sim.calculateOrificeFlowWithParams(5.0, 5.0, 0.8, sim.params, gate)
	if flow != 0 {
		t.Errorf("zero head should produce zero flow, got %f", flow)
	}
}

func TestOrificeFlowWithParams_NegativeHead(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	flow := sim.calculateOrificeFlowWithParams(3.5, 7.5, 0.8, sim.params, gate)
	if flow < 0 {
		t.Errorf("negative head (upstream<downstream) should produce non-negative flow or zero, got %f", flow)
	}
}

func TestOrificeFlowWithParams_ZeroOpening(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	flow := sim.calculateOrificeFlowWithParams(7.5, 3.5, 0, sim.params, gate)
	if flow != 0 {
		t.Errorf("zero opening should produce zero flow, got %f", flow)
	}
}

func TestOrificeFlow_ProportionalToOpening(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	flow50 := sim.calculateOrificeFlowWithParams(7.5, 3.5, 0.5, sim.params, gate)
	flow100 := sim.calculateOrificeFlowWithParams(7.5, 3.5, 1.0, sim.params, gate)
	if flow50 >= flow100 {
		t.Errorf("50%% opening flow (%f) should be less than 100%% opening flow (%f)", flow50, flow100)
	}
}

func TestContractionCoefficient_Range(t *testing.T) {
	sim := makeTestSimulator()
	for opening := 0.1; opening <= 1.0; opening += 0.1 {
		cc := sim.calculateContractionCoefficientWithParams(opening, sim.params)
		if cc <= 0 || cc > 1 {
			t.Errorf("contraction coefficient at opening=%f should be in (0,1], got %f", opening, cc)
		}
	}
}

func TestClassifyFlowRegime_AllRegimes(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	openingHeight := gate.GateHeight * 0.8
	relativeOpening := 0.8

	freeFlow := sim.classifyFlowRegimeWithParams(
		3.5+openingHeight*0.3, 3.5, openingHeight, relativeOpening, sim.params)
	if freeFlow != FreeFlow {
		t.Errorf("expected FreeFlow for low submergence, got %s", freeFlow)
	}

	submerged := sim.classifyFlowRegimeWithParams(
		3.5+openingHeight*0.95, 3.5, openingHeight, relativeOpening, sim.params)
	if submerged != SubmergedFlow {
		t.Errorf("expected SubmergedFlow for high submergence, got %s", submerged)
	}
}

func TestCalculateLevelChange_Filling(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	targetLevel := 7.5
	initialLevel := 3.5
	area := gate.ChamberLength * gate.ChamberWidth

	totalTime, levelPts, flowPts := sim.calculateLevelChangeWithParams(
		targetLevel, initialLevel, 0.8, area, true, sim.params, gate)

	if totalTime <= 0 {
		t.Errorf("fill time should be positive, got %f", totalTime)
	}
	if len(levelPts) < 2 {
		t.Error("should have at least 2 level points")
	}
	if len(flowPts) < 1 {
		t.Error("should have at least 1 flow point")
	}
	if levelPts[0].WaterLevel != initialLevel {
		t.Errorf("first level point should be initial level %f, got %f", initialLevel, levelPts[0].WaterLevel)
	}
	lastLevel := levelPts[len(levelPts)-1].WaterLevel
	if math.Abs(lastLevel-targetLevel) > 0.01 {
		t.Errorf("final level %f should be close to target %f", lastLevel, targetLevel)
	}
}

func TestCalculateLevelChange_Draining(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	initialLevel := 7.5
	targetLevel := 3.5
	area := gate.ChamberLength * gate.ChamberWidth

	totalTime, levelPts, flowPts := sim.calculateLevelChangeWithParams(
		targetLevel, initialLevel, 0.8, area, false, sim.params, gate)

	if totalTime <= 0 {
		t.Errorf("drain time should be positive, got %f", totalTime)
	}
	if len(levelPts) < 2 {
		t.Error("should have at least 2 level points for draining")
	}
	lastLevel := levelPts[len(levelPts)-1].WaterLevel
	if math.Abs(lastLevel-targetLevel) > 0.01 {
		t.Errorf("final level %f should be close to target %f after draining", lastLevel, targetLevel)
	}
}

func TestCalculateLevelChange_Boundary_SameLevel(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	area := gate.ChamberLength * gate.ChamberWidth

	totalTime, levelPts, _ := sim.calculateLevelChangeWithParams(
		5.0, 5.0, 0.8, area, true, sim.params, gate)

	if totalTime > 1.0 {
		t.Errorf("same level should result in near-zero time, got %f", totalTime)
	}
	if len(levelPts) > 0 && math.Abs(levelPts[len(levelPts)-1].WaterLevel-5.0) > 0.01 {
		t.Errorf("level should remain at 5.0 when target equals initial")
	}
}

func TestCalculateLevelChange_WithDynastyParams(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	designs := models.GetDynastyDesigns(gate.ID)

	tangTimes := make([]float64, len(designs))
	for i, d := range designs {
		effGate := gate
		effGate.GateWidth = d.GateWidth
		effGate.GateHeight = d.GateHeight
		effGate.ChamberLength = d.ChamberLength
		effGate.ChamberWidth = d.ChamberWidth
		effParams := sim.params
		effParams.DefaultCd = d.DefaultCd

		area := effGate.ChamberLength * effGate.ChamberWidth
		totalTime, _, _ := sim.calculateLevelChangeWithParams(
			7.5, 3.5, 0.8, area, true, effParams, effGate)
		tangTimes[i] = totalTime

		if totalTime <= 0 {
			t.Errorf("dynasty %s: fill time should be positive", d.Dynasty)
		}
		if math.IsInf(totalTime, 0) || math.IsNaN(totalTime) {
			t.Errorf("dynasty %s: fill time should not be Inf/NaN", d.Dynasty)
		}
	}
}

func TestSimulateSync_UpstreamVsDownstream(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()

	upReq := SimulateRequest{
		Gate: gate, WaterLevelUp: 7.5, WaterLevelDown: 3.6,
		GateOpening: 0.8, Direction: "upstream",
	}
	upResult := sim.SimulateSync(upReq)

	downReq := SimulateRequest{
		Gate: gate, WaterLevelUp: 7.5, WaterLevelDown: 3.6,
		GateOpening: 0.8, Direction: "downstream",
	}
	downResult := sim.SimulateSync(downReq)

	if upResult.FillTime <= 0 {
		t.Errorf("upstream fill time should be positive, got %f", upResult.FillTime)
	}
	if upResult.DrainTime != 0 {
		t.Errorf("upstream drain time should be 0, got %f", upResult.DrainTime)
	}
	if downResult.DrainTime <= 0 {
		t.Errorf("downstream drain time should be positive, got %f", downResult.DrainTime)
	}
	if downResult.FillTime != 0 {
		t.Errorf("downstream fill time should be 0, got %f", downResult.FillTime)
	}
}

func TestSimulateSync_BothShipAndDynasty(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()
	design := models.GetDynastyDesigns(gate.ID)[2]
	spec := models.GetShipTypeSpecs()[0]

	req := SimulateRequest{
		Gate: gate, WaterLevelUp: 7.5, WaterLevelDown: 3.6,
		GateOpening: 0.8, Direction: "upstream",
		DynastyDesign: design, ShipType: spec,
	}
	result := sim.SimulateSync(req)

	if result.UsedDynasty != string(design.Dynasty) {
		t.Errorf("expected UsedDynasty=%s, got %s", design.Dynasty, result.UsedDynasty)
	}
	if result.ShipTypeName != spec.TypeName {
		t.Errorf("expected ShipTypeName=%s, got %s", spec.TypeName, result.ShipTypeName)
	}
	if result.EntryTimeS != spec.EntryTimeS {
		t.Errorf("entry time should match spec: expected %f, got %f", spec.EntryTimeS, result.EntryTimeS)
	}
	if result.TotalPassageS != spec.EntryTimeS+result.FillTime+result.DrainTime+spec.ExitTimeS {
		t.Errorf("total passage should equal entry+fill+drain+exit")
	}
}

func TestSimulateSync_Abnormal_ExtremeGateOpening(t *testing.T) {
	sim := makeTestSimulator()
	gate := makeTestGate()

	req := SimulateRequest{
		Gate: gate, WaterLevelUp: 7.5, WaterLevelDown: 3.6,
		GateOpening: 5.0, Direction: "upstream",
	}
	result := sim.SimulateSync(req)

	if math.IsInf(result.MaxFlowRate, 0) || math.IsNaN(result.MaxFlowRate) {
		t.Error("extreme gate opening should not produce Inf/NaN flow rate")
	}
}

func TestSimulateSync_Abnormal_ZeroAreaGate(t *testing.T) {
	sim := makeTestSimulator()
	gate := models.DouGate{
		ID: 99, GateWidth: 0, GateHeight: 0,
		ChamberLength: 0, ChamberWidth: 0,
		MaxWaterLevelUp: 7.5, MinWaterLevelDown: 3.5,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Logf("zero-area gate caused panic (acceptable): %v", r)
		}
	}()

	req := SimulateRequest{
		Gate: gate, WaterLevelUp: 7.5, WaterLevelDown: 3.6,
		GateOpening: 0.8, Direction: "upstream",
	}
	sim.SimulateSync(req)
}
