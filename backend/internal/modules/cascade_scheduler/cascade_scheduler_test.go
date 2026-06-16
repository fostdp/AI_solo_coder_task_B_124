package cascade_scheduler

import (
	"testing"
	"time"

	"lingqu-dou-gate/internal/config"
	"lingqu-dou-gate/internal/models"
)

func makeTestScheduler() *CascadeScheduler {
	config.AppConfig.GAJSON = config.DefaultGAJSONConfig()
	return NewCascadeScheduler()
}

func makeTestGateIDs() []uint {
	return []uint{1, 2, 3, 4, 5}
}

func makeTestShips(n int) []models.ScheduleShip {
	ships := make([]models.ScheduleShip, n)
	now := time.Now()
	for i := 0; i < n; i++ {
		dir := "upstream"
		if i%2 == 1 {
			dir = "downstream"
		}
		ships[i] = models.ScheduleShip{
			ShipID:      uint(i + 1),
			ShipName:    "测试船" + intToStrI(i+1),
			Priority:    (i % 5) + 1,
			ArrivalTime: now.Add(time.Duration(i*15) * time.Minute),
			Direction:   dir,
		}
	}
	return ships
}

func TestNewCascadeScheduler(t *testing.T) {
	cs := makeTestScheduler()
	if cs == nil {
		t.Fatal("CascadeScheduler should not be nil")
	}
}

func TestOptimize_Basic(t *testing.T) {
	cs := makeTestScheduler()
	req := models.MultiStageOptimizeRequest{
		GateIDs: makeTestGateIDs(),
		Ships:   makeTestShips(8),
	}
	result := cs.Optimize(req)

	if result == nil {
		t.Fatal("result should not be nil")
	}
	if len(result.Routes) != 8 {
		t.Errorf("expected 8 routes, got %d", len(result.Routes))
	}
	if result.TotalWaitTimeS < 0 {
		t.Errorf("total wait time should be non-negative, got %f", result.TotalWaitTimeS)
	}
	if result.TotalTravelTimeS < 0 {
		t.Errorf("total travel time should be non-negative, got %f", result.TotalTravelTimeS)
	}
	if result.TotalWaterUsedM3 <= 0 {
		t.Errorf("total water should be positive, got %f", result.TotalWaterUsedM3)
	}
	if result.ThroughputShips != 8 {
		t.Errorf("throughput ships should be 8, got %d", result.ThroughputShips)
	}
}

func TestOptimize_EachShipHasRoute(t *testing.T) {
	cs := makeTestScheduler()
	ships := makeTestShips(5)
	req := models.MultiStageOptimizeRequest{
		GateIDs: makeTestGateIDs(),
		Ships:   ships,
	}
	result := cs.Optimize(req)

	shipIDSet := map[uint]bool{}
	for _, r := range result.Routes {
		shipIDSet[r.ShipID] = true
	}
	for _, s := range ships {
		if !shipIDSet[s.ShipID] {
			t.Errorf("ship %d missing from routes", s.ShipID)
		}
	}
}

func TestOptimize_PriorityRespected(t *testing.T) {
	cs := makeTestScheduler()
	ships := makeTestShips(10)
	req := models.MultiStageOptimizeRequest{
		GateIDs: makeTestGateIDs(),
		Ships:   ships,
	}
	result := cs.Optimize(req)

	if len(result.Routes) < 2 {
		return
	}
	if result.Routes[0].Priority < result.Routes[len(result.Routes)-1].Priority {
		t.Error("routes should be sorted by priority descending")
	}
}

func TestOptimize_Bidirectional(t *testing.T) {
	cs := makeTestScheduler()
	ships := makeTestShips(10)
	req := models.MultiStageOptimizeRequest{
		GateIDs: makeTestGateIDs(),
		Ships:   ships,
	}
	result := cs.Optimize(req)

	hasUp := false
	hasDown := false
	for _, r := range result.Routes {
		if r.Direction == "upstream" {
			hasUp = true
		}
		if r.Direction == "downstream" {
			hasDown = true
		}
	}
	if !hasUp || !hasDown {
		t.Error("should have both upstream and downstream routes")
	}
}

func TestOptimize_SpeedFactorImprovesTime(t *testing.T) {
	cs := makeTestScheduler()
	ships := makeTestShips(5)

	reqSlow := models.MultiStageOptimizeRequest{
		GateIDs:          makeTestGateIDs(),
		Ships:            ships,
		TravelSpeedFactor: 0.5,
	}
	resultSlow := cs.Optimize(reqSlow)

	reqFast := models.MultiStageOptimizeRequest{
		GateIDs:          makeTestGateIDs(),
		Ships:            ships,
		TravelSpeedFactor: 2.0,
	}
	resultFast := cs.Optimize(reqFast)

	if resultSlow.TotalTravelTimeS <= resultFast.TotalTravelTimeS*1.1 {
		t.Errorf("slow speed factor should increase travel time: slow=%.0f, fast=%.0f",
			resultSlow.TotalTravelTimeS, resultFast.TotalTravelTimeS)
	}
}

func TestOptimize_BoundaryEmptyFleet(t *testing.T) {
	cs := makeTestScheduler()
	req := models.MultiStageOptimizeRequest{
		GateIDs: makeTestGateIDs(),
		Ships:   []models.ScheduleShip{},
	}
	result := cs.Optimize(req)

	if result == nil {
		t.Fatal("result should not be nil for empty fleet")
	}
	if len(result.Routes) == 0 {
		t.Error("empty fleet should generate default routes")
	}
}

func TestOptimize_BoundaryEmptyGateIDs(t *testing.T) {
	cs := makeTestScheduler()
	ships := makeTestShips(3)
	req := models.MultiStageOptimizeRequest{
		GateIDs: []uint{},
		Ships:   ships,
	}
	result := cs.Optimize(req)

	if result == nil {
		t.Fatal("result should not be nil for empty gate IDs")
	}
}

func TestOptimize_BoundarySingleGate(t *testing.T) {
	cs := makeTestScheduler()
	ships := makeTestShips(3)
	req := models.MultiStageOptimizeRequest{
		GateIDs: []uint{1},
		Ships:   ships,
	}
	result := cs.Optimize(req)

	if result == nil {
		t.Fatal("result should not be nil for single gate")
	}
}

func TestOptimize_BoundaryZeroSpeedFactor(t *testing.T) {
	cs := makeTestScheduler()
	ships := makeTestShips(3)
	req := models.MultiStageOptimizeRequest{
		GateIDs:          makeTestGateIDs(),
		Ships:            ships,
		TravelSpeedFactor: 0,
	}
	result := cs.Optimize(req)

	if result == nil {
		t.Fatal("result should not be nil with zero speed factor")
	}
}

func TestOptimize_AbnormalNegativeSpeedFactor(t *testing.T) {
	cs := makeTestScheduler()
	ships := makeTestShips(3)
	req := models.MultiStageOptimizeRequest{
		GateIDs:          makeTestGateIDs(),
		Ships:            ships,
		TravelSpeedFactor: -1.0,
	}
	result := cs.Optimize(req)

	if result == nil {
		t.Fatal("result should not be nil with negative speed factor")
	}
	if result.TotalTravelTimeS < 0 {
		t.Error("travel time should not be negative with negative speed factor")
	}
}

func TestOptimize_AbnormalLargeFleet(t *testing.T) {
	cs := makeTestScheduler()
	ships := makeTestShips(100)
	req := models.MultiStageOptimizeRequest{
		GateIDs: makeTestGateIDs(),
		Ships:   ships,
	}
	result := cs.Optimize(req)

	if result == nil {
		t.Fatal("result should not be nil with large fleet")
	}
	if len(result.Routes) != 100 {
		t.Errorf("expected 100 routes, got %d", len(result.Routes))
	}
}

func TestOptimize_GateUtilization(t *testing.T) {
	cs := makeTestScheduler()
	ships := makeTestShips(8)
	req := models.MultiStageOptimizeRequest{
		GateIDs: makeTestGateIDs(),
		Ships:   ships,
	}
	result := cs.Optimize(req)

	if len(result.GateUtilization) == 0 {
		t.Error("gate utilization should not be empty")
	}
	for gid, util := range result.GateUtilization {
		if util < 0 || util > 1 {
			t.Errorf("gate %d utilization %f out of [0,1]", gid, util)
		}
	}
}

func TestFindGateFreeAfterSorted_Empty(t *testing.T) {
	after := time.Now()
	freeAt := findGateFreeAfterSorted(nil, after)
	if !freeAt.Equal(after) {
		t.Errorf("with no ranges, free time should equal input, got %v vs %v", freeAt, after)
	}
}

func TestFindGateFreeAfterSorted_BinarySearch(t *testing.T) {
	baseTime := time.Now()
	ranges := []timeRange{
		{Start: baseTime, End: baseTime.Add(100 * time.Second)},
		{Start: baseTime.Add(200 * time.Second), End: baseTime.Add(350 * time.Second)},
		{Start: baseTime.Add(400 * time.Second), End: baseTime.Add(500 * time.Second)},
	}

	freeAt := findGateFreeAfterSorted(ranges, baseTime.Add(50*time.Second))
	if freeAt.Before(baseTime.Add(100 * time.Second)) {
		t.Errorf("sorted free time should be after first range end, got %v", freeAt)
	}

	freeAt2 := findGateFreeAfterSorted(ranges, baseTime.Add(300*time.Second))
	if freeAt2.Before(baseTime.Add(350 * time.Second)) {
		t.Errorf("sorted free time in gap should be after second range end, got %v", freeAt2)
	}

	freeAt3 := findGateFreeAfterSorted(ranges, baseTime.Add(600*time.Second))
	if freeAt3.Before(baseTime.Add(600 * time.Second)) {
		t.Errorf("sorted free time after all ranges should be at or after input, got %v", freeAt3)
	}
}

func TestInsertSorted(t *testing.T) {
	baseTime := time.Now()
	ranges := []timeRange{
		{Start: baseTime, End: baseTime.Add(100 * time.Second)},
		{Start: baseTime.Add(200 * time.Second), End: baseTime.Add(350 * time.Second)},
	}

	result := insertSorted(ranges, timeRange{
		Start: baseTime.Add(150 * time.Second),
		End:   baseTime.Add(180 * time.Second),
	})

	if len(result) != 3 {
		t.Fatalf("expected 3 ranges after insert, got %d", len(result))
	}
	if result[1].Start.After(result[2].Start) {
		t.Error("insertSorted should maintain sorted order")
	}
	if result[0].Start.After(result[1].Start) {
		t.Error("insertSorted should maintain sorted order")
	}
}
