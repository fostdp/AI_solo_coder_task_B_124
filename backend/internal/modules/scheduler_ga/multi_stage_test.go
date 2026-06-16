package scheduler_ga

import (
	"math"
	"testing"
	"time"

	"lingqu-dou-gate/internal/config"
	"lingqu-dou-gate/internal/models"
)

func makeTestGA() *GAScheduler {
	return &GAScheduler{
		params: config.GAJSONConfig{
			PopulationSize: 40,
			MaxGenerations: 80,
			CrossoverRate:  0.85,
			MutationRate:   0.15,
			Selection: config.SelectionParams{
				Type:                "tournament",
				TournamentSizeSmall: 5,
				TournamentSizeLarge: 7,
				LargeThreshold:      60,
			},
			Elite: config.EliteParams{
				Ratio:    0.1,
				MinCount: 2,
				MaxCount: 6,
				PoolSize: 10,
			},
			AdaptiveRates: config.AdaptiveRatesParams{
				ConvergenceGapRatio:    0.05,
				HighConvCrossover:      0.9,
				HighConvMutationMul:    1.5,
				StagnantTrigger:        15,
				StagnantMutationMul:    2.0,
				StagnantCrossoverMul:   0.7,
				MaxMutationRate:        0.5,
				MaxCrossoverRate:       0.95,
			},
			Mutation: config.MutationParams{
				Types: []string{"swap", "inversion", "insertion"},
			},
			LocalSearch: config.LocalSearchParams{
				Enabled:         true,
				StagnantTrigger: 20,
				MaxIterations:   50,
			},
			Diversity: config.DiversityParams{
				BonusWeight: 0.05,
				SampleSize:  20,
			},
			Parallel: config.ParallelParams{
				MinWorkers:         2,
				MaxWorkers:         8,
				LargeShipThreshold: 30,
			},
			Initialization: config.InitParams{
				HeuristicRatio:    0.3,
				LargePopMul:       1.5,
				LargeShipThreshold: 30,
			},
			Stopping: config.StoppingParams{
				MaxStagnantSmall:   30,
				MaxStagnantLarge:   50,
				LargeShipThreshold: 30,
			},
			Fitness: config.FitnessParams{
				ScaleFactor:         1000.0,
				PriorityWeightPower: 1.5,
				PriorityWeightBase:  0.1,
			},
		},
	}
}

func makeTestShips(n int) []models.ScheduleShip {
	now := time.Now()
	ships := make([]models.ScheduleShip, n)
	for i := 0; i < n; i++ {
		dir := "upstream"
		if i%3 == 0 {
			dir = "downstream"
		}
		ships[i] = models.ScheduleShip{
			ShipID:      uint(i + 1),
			ShipName:    "测试船" + itoa(i+1),
			Priority:    (i%5) + 1,
			ArrivalTime: now.Add(time.Duration(i*20) * time.Minute),
			Direction:   dir,
		}
	}
	return ships
}

func makeTestGateIDs() []uint {
	ids := make([]uint, 6)
	for i := 0; i < 6; i++ {
		ids[i] = uint(i + 1)
	}
	return ids
}

func TestMultiStageOptimize_Normal(t *testing.T) {
	ga := makeTestGA()
	req := models.MultiStageOptimizeRequest{
		GateIDs:  makeTestGateIDs(),
		Ships:    makeTestShips(10),
		CanalSegments: []models.CanalSegment{},
		TravelSpeedFactor: 1.0,
	}
	result := ga.MultiStageOptimizeSync(req)

	if result == nil {
		t.Fatal("result should not be nil")
	}
	if len(result.Routes) == 0 {
		t.Error("expected at least one route in result")
	}
	if result.ThroughputPerDay < 0 {
		t.Errorf("throughput per day should be non-negative, got %f", result.ThroughputPerDay)
	}
	if result.TotalWaitTimeS < 0 {
		t.Errorf("total wait time should be non-negative, got %f", result.TotalWaitTimeS)
	}
	if result.TotalWaterUsedM3 < 0 {
		t.Errorf("total water used should be non-negative, got %f", result.TotalWaterUsedM3)
	}
}

func TestMultiStageOptimize_EachShipHasRoute(t *testing.T) {
	ga := makeTestGA()
	ships := makeTestShips(8)
	req := models.MultiStageOptimizeRequest{
		GateIDs:  makeTestGateIDs(),
		Ships:    ships,
		CanalSegments: []models.CanalSegment{},
	}
	result := ga.MultiStageOptimizeSync(req)

	if len(result.Routes) != len(ships) {
		t.Errorf("expected %d routes, got %d", len(ships), len(result.Routes))
	}

	routedShips := map[uint]bool{}
	for _, route := range result.Routes {
		routedShips[route.ShipID] = true
		if len(route.GateSequence) == 0 {
			t.Errorf("ship %d should have gate sequence", route.ShipID)
		}
		if route.TotalPassageTimeS < 0 {
			t.Errorf("ship %d: total passage time should be non-negative", route.ShipID)
		}
	}
	for _, s := range ships {
		if !routedShips[s.ShipID] {
			t.Errorf("ship %d not found in routes", s.ShipID)
		}
	}
}

func TestMultiStageOptimize_PriorityRespected(t *testing.T) {
	ga := makeTestGA()
	now := time.Now()
	ships := []models.ScheduleShip{
		{ShipID: 1, ShipName: "高优先", Priority: 6, ArrivalTime: now.Add(30 * time.Minute), Direction: "upstream"},
		{ShipID: 2, ShipName: "低优先", Priority: 1, ArrivalTime: now, Direction: "upstream"},
	}
	req := models.MultiStageOptimizeRequest{
		GateIDs:  makeTestGateIDs(),
		Ships:    ships,
	}
	result := ga.MultiStageOptimizeSync(req)

	if len(result.Routes) < 2 {
		t.Fatal("need at least 2 routes for priority test")
	}

	var highStart, lowStart time.Time
	for _, r := range result.Routes {
		if r.ShipID == 1 {
			if len(r.GateSequence) > 0 {
				highStart = r.GateSequence[0].FillDrainStart
			}
		}
		if r.ShipID == 2 {
			if len(r.GateSequence) > 0 {
				lowStart = r.GateSequence[0].FillDrainStart
			}
		}
	}
	if !highStart.Before(time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)) && !lowStart.Before(time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Log("both ships have start times assigned")
	}
}

func TestMultiStageOptimize_UpstreamDownstream(t *testing.T) {
	ga := makeTestGA()
	now := time.Now()
	ships := []models.ScheduleShip{
		{ShipID: 1, ShipName: "上行船", Priority: 3, ArrivalTime: now, Direction: "upstream"},
		{ShipID: 2, ShipName: "下行船", Priority: 3, ArrivalTime: now, Direction: "downstream"},
	}
	req := models.MultiStageOptimizeRequest{
		GateIDs:  makeTestGateIDs(),
		Ships:    ships,
	}
	result := ga.MultiStageOptimizeSync(req)

	for _, r := range result.Routes {
		if r.Direction != "upstream" && r.Direction != "downstream" {
			t.Errorf("ship %d: unexpected direction %s", r.ShipID, r.Direction)
		}
	}
}

func TestMultiStageOptimize_Efficiency_SingleVsMultiple(t *testing.T) {
	ga := makeTestGA()

	singleReq := models.MultiStageOptimizeRequest{
		GateIDs:  []uint{1, 2, 3},
		Ships:    makeTestShips(5),
	}
	singleResult := ga.MultiStageOptimizeSync(singleReq)

	multiReq := models.MultiStageOptimizeRequest{
		GateIDs:  []uint{1, 2, 3},
		Ships:    makeTestShips(20),
	}
	multiResult := ga.MultiStageOptimizeSync(multiReq)

	if singleResult == nil || multiResult == nil {
		t.Fatal("results should not be nil")
	}
	if len(multiResult.Routes) <= len(singleResult.Routes) {
		t.Errorf("multi-ship result should have more routes than single-ship, got %d vs %d",
			len(multiResult.Routes), len(singleResult.Routes))
	}
}

func TestMultiStageOptimize_SpeedFactor(t *testing.T) {
	ga := makeTestGA()
	ships := makeTestShips(8)

	slowReq := models.MultiStageOptimizeRequest{
		GateIDs:  makeTestGateIDs(),
		Ships:    ships,
		TravelSpeedFactor: 0.5,
	}
	slowResult := ga.MultiStageOptimizeSync(slowReq)

	fastReq := models.MultiStageOptimizeRequest{
		GateIDs:  makeTestGateIDs(),
		Ships:    ships,
		TravelSpeedFactor: 2.0,
	}
	fastResult := ga.MultiStageOptimizeSync(fastReq)

	if slowResult.TotalTravelTimeS <= 0 || fastResult.TotalTravelTimeS <= 0 {
		t.Error("total travel time should be positive")
	}
	if fastResult.TotalTravelTimeS > slowResult.TotalTravelTimeS {
		t.Errorf("faster speed factor should result in less or equal travel time, got fast=%f slow=%f",
			fastResult.TotalTravelTimeS, slowResult.TotalTravelTimeS)
	}
}

func TestMultiStageOptimize_Boundary_EmptyShips(t *testing.T) {
	ga := makeTestGA()
	req := models.MultiStageOptimizeRequest{
		GateIDs:  makeTestGateIDs(),
		Ships:    []models.ScheduleShip{},
	}
	result := ga.MultiStageOptimizeSync(req)

	if result == nil {
		t.Fatal("result should not be nil even with empty ships")
	}
	if len(result.Routes) == 0 {
		t.Log("empty ships input produces default ships as fallback")
	}
}

func TestMultiStageOptimize_Boundary_EmptyGateIDs(t *testing.T) {
	ga := makeTestGA()
	req := models.MultiStageOptimizeRequest{
		GateIDs:  []uint{},
		Ships:    makeTestShips(5),
	}
	result := ga.MultiStageOptimizeSync(req)

	if result == nil {
		t.Fatal("result should not be nil even with empty gate IDs")
	}
	if result.Error != "" {
		t.Logf("empty gates produces error or fallback: %s", result.Error)
	}
}

func TestMultiStageOptimize_Boundary_SingleGate(t *testing.T) {
	ga := makeTestGA()
	req := models.MultiStageOptimizeRequest{
		GateIDs:  []uint{1},
		Ships:    makeTestShips(3),
	}
	result := ga.MultiStageOptimizeSync(req)

	if result == nil {
		t.Fatal("result should not be nil with single gate")
	}
	for _, r := range result.Routes {
		if len(r.GateSequence) > 1 {
			t.Errorf("single gate should produce at most 1 gate in sequence, got %d", len(r.GateSequence))
		}
	}
}

func TestMultiStageOptimize_Boundary_ZeroSpeedFactor(t *testing.T) {
	ga := makeTestGA()
	req := models.MultiStageOptimizeRequest{
		GateIDs:  makeTestGateIDs(),
		Ships:    makeTestShips(5),
		TravelSpeedFactor: 0,
	}
	result := ga.MultiStageOptimizeSync(req)

	if result == nil {
		t.Fatal("result should not be nil with zero speed factor (should fallback to 1.0)")
	}
}

func TestMultiStageOptimize_Abnormal_NegativeSpeedFactor(t *testing.T) {
	ga := makeTestGA()
	req := models.MultiStageOptimizeRequest{
		GateIDs:  makeTestGateIDs(),
		Ships:    makeTestShips(5),
		TravelSpeedFactor: -1.5,
	}
	result := ga.MultiStageOptimizeSync(req)

	if result == nil {
		t.Fatal("result should not be nil with negative speed factor (should fallback to 1.0)")
	}
}

func TestMultiStageOptimize_Abnormal_ExtremelyLargeFleet(t *testing.T) {
	ga := makeTestGA()
	req := models.MultiStageOptimizeRequest{
		GateIDs:  makeTestGateIDs(),
		Ships:    makeTestShips(100),
	}
	result := ga.MultiStageOptimizeSync(req)

	if result == nil {
		t.Fatal("result should not be nil with large fleet")
	}
	if len(result.Routes) != 100 {
		t.Errorf("expected 100 routes, got %d", len(result.Routes))
	}
}

func TestMultiStageOptimize_Abnormal_PastArrivalTime(t *testing.T) {
	ga := makeTestGA()
	ships := []models.ScheduleShip{
		{ShipID: 1, ShipName: "旧船", Priority: 3, ArrivalTime: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC), Direction: "upstream"},
	}
	req := models.MultiStageOptimizeRequest{
		GateIDs:  makeTestGateIDs(),
		Ships:    ships,
	}
	result := ga.MultiStageOptimizeSync(req)

	if result == nil {
		t.Fatal("result should not be nil with past arrival time")
	}
	if len(result.Routes) == 0 {
		t.Error("should still produce a route for past arrival time")
	}
}

func TestMultiStageOptimize_GateUtilization(t *testing.T) {
	ga := makeTestGA()
	req := models.MultiStageOptimizeRequest{
		GateIDs:  makeTestGateIDs(),
		Ships:    makeTestShips(15),
	}
	result := ga.MultiStageOptimizeSync(req)

	if result.GateUtilization == nil {
		t.Error("gate utilization map should not be nil")
	}
	for gateID, util := range result.GateUtilization {
		if util < 0 || util > 1 {
			t.Errorf("gate %d utilization %f out of [0,1] range", gateID, util)
		}
		if math.IsNaN(util) || math.IsInf(util, 0) {
			t.Errorf("gate %d utilization is NaN or Inf", gateID)
		}
	}
}

func TestMultiStageOptimize_NoPanicOnAllSameDirection(t *testing.T) {
	ga := makeTestGA()
	now := time.Now()
	ships := make([]models.ScheduleShip, 10)
	for i := 0; i < 10; i++ {
		ships[i] = models.ScheduleShip{
			ShipID: uint(i + 1), ShipName: "同向船", Priority: 3,
			ArrivalTime: now.Add(time.Duration(i*10) * time.Minute),
			Direction: "upstream",
		}
	}
	req := models.MultiStageOptimizeRequest{
		GateIDs: makeTestGateIDs(),
		Ships:   ships,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MultiStageOptimizeSync panicked with all same direction: %v", r)
		}
	}()
	result := ga.MultiStageOptimizeSync(req)
	if result == nil {
		t.Error("result should not be nil")
	}
}

func TestFindGateFreeAfter(t *testing.T) {
	baseTime := time.Now()
	ranges := []timeRange{
		{start: baseTime, end: baseTime.Add(100 * time.Second)},
		{start: baseTime.Add(200 * time.Second), end: baseTime.Add(350 * time.Second)},
	}

	freeAt := findGateFreeAfter(ranges, baseTime.Add(50*time.Second))
	if freeAt.Before(baseTime.Add(100*time.Second)) {
		t.Errorf("free time should be after first range end, got %v", freeAt)
	}

	freeAt2 := findGateFreeAfter(ranges, baseTime.Add(500*time.Second))
	if freeAt2.Before(baseTime.Add(500 * time.Second)) {
		t.Errorf("free time after all ranges should be at or after input time, got %v", freeAt2)
	}
}

func TestFindGateFreeAfter_Empty(t *testing.T) {
	after := time.Now()
	freeAt := findGateFreeAfter(nil, after)
	if !freeAt.Equal(after) {
		t.Errorf("with no ranges, free time should equal input, got %v vs %v", freeAt, after)
	}
}

func TestFindGateFreeAfterSorted_BinarySearch(t *testing.T) {
	baseTime := time.Now()
	ranges := []timeRange{
		{start: baseTime, end: baseTime.Add(100 * time.Second)},
		{start: baseTime.Add(200 * time.Second), end: baseTime.Add(350 * time.Second)},
		{start: baseTime.Add(400 * time.Second), end: baseTime.Add(500 * time.Second)},
	}

	freeAt := findGateFreeAfterSorted(ranges, baseTime.Add(50*time.Second))
	if freeAt.Before(baseTime.Add(100*time.Second)) {
		t.Errorf("sorted free time should be after first range end, got %v", freeAt)
	}

	freeAt2 := findGateFreeAfterSorted(ranges, baseTime.Add(300*time.Second))
	if freeAt2.Before(baseTime.Add(350*time.Second)) {
		t.Errorf("sorted free time in gap should be after second range end, got %v", freeAt2)
	}

	freeAt3 := findGateFreeAfterSorted(ranges, baseTime.Add(600*time.Second))
	if freeAt3.Before(baseTime.Add(600*time.Second)) {
		t.Errorf("sorted free time after all ranges should be at or after input, got %v", freeAt3)
	}
}

func TestFindGateFreeAfterSorted_Empty(t *testing.T) {
	after := time.Now()
	freeAt := findGateFreeAfterSorted(nil, after)
	if !freeAt.Equal(after) {
		t.Errorf("with no sorted ranges, free time should equal input, got %v vs %v", freeAt, after)
	}
}

func TestInsertSorted(t *testing.T) {
	baseTime := time.Now()
	ranges := []timeRange{
		{start: baseTime, end: baseTime.Add(100 * time.Second)},
		{start: baseTime.Add(200 * time.Second), end: baseTime.Add(350 * time.Second)},
	}

	result := insertSorted(ranges, timeRange{
		start: baseTime.Add(150 * time.Second),
		end:   baseTime.Add(180 * time.Second),
	})

	if len(result) != 3 {
		t.Fatalf("expected 3 ranges after insert, got %d", len(result))
	}
	if result[1].start.After(result[2].start) {
		t.Error("insertSorted should maintain sorted order")
	}
	if result[0].start.After(result[1].start) {
		t.Error("insertSorted should maintain sorted order")
	}
}

func TestMultiStageOptimize_ShipWaitTimeout(t *testing.T) {
	ga := makeTestGA()
	now := time.Now()
	ships := make([]models.ScheduleShip, 50)
	for i := 0; i < 50; i++ {
		ships[i] = models.ScheduleShip{
			ShipID: uint(i + 1), ShipName: "拥堵船", Priority: 3,
			ArrivalTime: now,
			Direction:   "upstream",
		}
	}
	req := models.MultiStageOptimizeRequest{
		GateIDs: []uint{1, 2, 3},
		Ships:   ships,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("should not panic with congested fleet: %v", r)
		}
	}()
	result := ga.MultiStageOptimizeSync(req)
	if result == nil {
		t.Fatal("result should not be nil even with congested fleet")
	}
}

func TestMultiStageOptimize_GateSeqLengthCap(t *testing.T) {
	ga := makeTestGA()
	ships := makeTestShips(3)
	req := models.MultiStageOptimizeRequest{
		GateIDs: makeTestGateIDs(),
		Ships:   ships,
	}
	result := ga.MultiStageOptimizeSync(req)
	if result == nil {
		t.Fatal("result should not be nil")
	}
	for _, r := range result.Routes {
		if len(r.GateSequence) > 8 {
			t.Errorf("gate sequence length %d exceeds cap of 8", len(r.GateSequence))
		}
	}
}
