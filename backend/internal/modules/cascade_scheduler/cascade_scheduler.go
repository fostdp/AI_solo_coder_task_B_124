package cascade_scheduler

import (
	"sort"
	"sync"
	"time"

	"lingqu-dou-gate/internal/config"
	"lingqu-dou-gate/internal/models"
)

type CascadeScheduler struct {
	mu     sync.RWMutex
	params config.GAJSONConfig
}

func NewCascadeScheduler() *CascadeScheduler {
	return &CascadeScheduler{
		params: config.AppConfig.GAJSON,
	}
}

func (cs *CascadeScheduler) ReloadConfig() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.params = config.AppConfig.GAJSON
}

type timeRange struct {
	Start time.Time
	End   time.Time
}

func findGateFreeAfterSorted(ranges []timeRange, after time.Time) time.Time {
	if len(ranges) == 0 {
		return after
	}
	lo, hi := 0, len(ranges)-1
	for lo <= hi {
		mid := lo + (hi-lo)/2
		if ranges[mid].Start.After(after) {
			hi = mid - 1
		} else {
			lo = mid + 1
		}
	}
	startIdx := hi
	t := after
	for i := startIdx; i >= 0 && i < len(ranges); i++ {
		r := ranges[i]
		if !t.Before(r.Start.Add(-1*time.Second)) && t.Before(r.End) {
			t = r.End
		}
		if r.Start.After(t) {
			break
		}
	}
	for i := startIdx + 1; i < len(ranges); i++ {
		r := ranges[i]
		if !t.Before(r.Start.Add(-1*time.Second)) && t.Before(r.End) {
			t = r.End
		}
		if r.Start.After(t) {
			break
		}
	}
	return t
}

func insertSorted(ranges []timeRange, tr timeRange) []timeRange {
	n := len(ranges)
	if n == 0 {
		return []timeRange{tr}
	}
	lo, hi := 0, n-1
	for lo <= hi {
		mid := lo + (hi-lo)/2
		if ranges[mid].Start.Before(tr.Start) {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	result := make([]timeRange, n+1)
	copy(result, ranges[:lo])
	result[lo] = tr
	copy(result[lo+1:], ranges[lo:])
	return result
}

func intToStrI(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	digits := []byte{}
	for v > 0 {
		digits = append([]byte{byte('0' + v%10)}, digits...)
		v /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

func (cs *CascadeScheduler) Optimize(req models.MultiStageOptimizeRequest) *models.MultiStageOptimizeResult {
	gatesMap := map[uint]*models.DouGate{}
	for _, id := range req.GateIDs {
		gatesMap[id] = &models.DouGate{ID: id, GateWidth: 6.0, GateHeight: 4.5,
			ChamberLength: 60, ChamberWidth: 7.0, MaxWaterLevelUp: 7.5, MinWaterLevelDown: 3.5}
	}
	if len(gatesMap) == 0 {
		for i := uint(1); i <= 10; i++ {
			gatesMap[i] = &models.DouGate{ID: i, GateWidth: 6.0, GateHeight: 4.5,
				ChamberLength: 60, ChamberWidth: 7.0, MaxWaterLevelUp: 7.5, MinWaterLevelDown: 3.5}
			req.GateIDs = append(req.GateIDs, i)
		}
	}

	ships := req.Ships
	if len(ships) == 0 {
		now := time.Now()
		for i := 1; i <= 15; i++ {
			dir := "upstream"
			if i%3 == 0 {
				dir = "downstream"
			}
			ships = append(ships, models.ScheduleShip{
				ShipID:      uint(i),
				ShipName:    "联调船" + intToStrI(i),
				Priority:    (i % 5) + 1,
				ArrivalTime: now.Add(time.Duration(i*25) * time.Minute),
				Direction:   dir,
			})
		}
	}

	segments := req.CanalSegments
	if len(segments) == 0 {
		for _, s := range models.GetDefaultCanalSegments() {
			segments = append(segments, *s)
		}
	}

	segByFrom := map[uint]models.CanalSegment{}
	segByTo := map[uint]models.CanalSegment{}
	for _, seg := range segments {
		segByFrom[seg.FromGateID] = seg
		segByTo[seg.ToGateID] = seg
	}

	speedFactor := req.TravelSpeedFactor
	if speedFactor <= 0 {
		speedFactor = 1.0
	}

	routeMap := map[uint]*models.MultiStageShipRoute{}
	gateUsed := map[uint][]timeRange{}
	gateTotalTime := map[uint]float64{}
	segmentShips := map[string]int{}
	conflictCount := 0
	totalWait := 0.0
	totalTravel := 0.0
	totalWater := 0.0

	maxGateSeqLen := len(req.GateIDs)
	if maxGateSeqLen > 8 {
		maxGateSeqLen = 8
	}

	sortedShips := make([]models.ScheduleShip, len(ships))
	copy(sortedShips, ships)
	sort.Slice(sortedShips, func(a, b int) bool {
		pa := -float64(sortedShips[a].Priority)*1e6 + float64(sortedShips[a].ArrivalTime.Unix())
		pb := -float64(sortedShips[b].Priority)*1e6 + float64(sortedShips[b].ArrivalTime.Unix())
		return pa < pb
	})

	maxWaitPerShip := 7200.0

	for _, ship := range sortedShips {
		direction := ship.Direction
		if direction == "" {
			direction = "upstream"
		}
		var gateSeq []uint
		if direction == "upstream" {
			for i := len(req.GateIDs) - 1; i >= 0; i-- {
				gateSeq = append(gateSeq, req.GateIDs[i])
			}
		} else {
			gateSeq = append(gateSeq, req.GateIDs...)
		}
		if len(gateSeq) > maxGateSeqLen {
			gateSeq = gateSeq[:maxGateSeqLen]
		}

		shipArrival := ship.ArrivalTime
		if shipArrival.Before(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)) {
			shipArrival = time.Now()
		}

		route := &models.MultiStageShipRoute{
			ShipID:       ship.ShipID,
			ShipName:     ship.ShipName,
			Direction:    direction,
			Priority:     ship.Priority,
			OriginGateID: gateSeq[0],
			DestGateID:   gateSeq[len(gateSeq)-1],
		}

		currentTime := shipArrival
		shipWait := 0.0
		shipTruncated := false

		for i, gid := range gateSeq {
			gate := gatesMap[gid]
			if gate == nil {
				continue
			}
			if shipWait > maxWaitPerShip {
				shipTruncated = true
				break
			}

			baseChamberArea := gate.ChamberLength * gate.ChamberWidth
			if baseChamberArea <= 0 {
				baseChamberArea = 420.0
			}
			headDiff := gate.MaxWaterLevelUp - gate.MinWaterLevelDown
			if headDiff <= 0 {
				headDiff = 3.8
			}
			baseFlow := 32.0
			fillTimeSec := (baseChamberArea * headDiff) / baseFlow * 1.1
			drainTimeSec := fillTimeSec * 0.95
			entryTimeS := 150.0 + float64(ship.Priority)*10
			exitTimeS := 120.0 + float64(ship.Priority)*8

			var travelToNext float64
			if i < len(gateSeq)-1 {
				nextGid := gateSeq[i+1]
				var segFound *models.CanalSegment
				if direction == "downstream" {
					if s, ok := segByFrom[gid]; ok && s.ToGateID == nextGid {
						segFound = &s
					}
				} else {
					if s, ok := segByTo[gid]; ok && s.FromGateID == nextGid {
						segFound = &s
					}
				}
				if segFound != nil {
					travelToNext = segFound.TravelTimeS / speedFactor
				}
				if travelToNext <= 0 {
					travelToNext = 400 / speedFactor
				}
			}

			segKey := direction + "_" + intToStrI(int(gid))
			if segmentShips[segKey] > 2 && travelToNext < 300 {
				conflictCount++
				currentTime = currentTime.Add(120 * time.Second)
			}
			segmentShips[segKey]++

			arrivalAtGate := currentTime
			sortedRanges := gateUsed[gid]
			gateFree := findGateFreeAfterSorted(sortedRanges, arrivalAtGate)
			waitS := gateFree.Sub(arrivalAtGate).Seconds()
			if waitS < 0 {
				waitS = 0
			}
			totalWait += waitS
			shipWait += waitS

			fillDrainStart := arrivalAtGate.Add(time.Duration(waitS * float64(time.Second)))
			fillDrainEnd := fillDrainStart.Add(time.Duration((fillTimeSec + drainTimeSec) * float64(time.Second)))
			entryTime := fillDrainEnd
			exitTime := entryTime.Add(time.Duration((entryTimeS + exitTimeS) * float64(time.Second)))

			travelTimeS := travelToNext
			if i < len(gateSeq)-1 {
				totalTravel += travelTimeS
			}

			departureTime := exitTime.Add(time.Duration(travelTimeS * float64(time.Second)))

			waterUsed := baseChamberArea * headDiff
			totalWater += waterUsed

			gateSched := models.MultiStageGateSchedule{
				GateID:         gid,
				GateName:       "陡门" + intToStrI(int(gid)),
				ArrivalTime:    arrivalAtGate,
				FillDrainStart: fillDrainStart,
				FillDrainEnd:   fillDrainEnd,
				EntryTime:      entryTime,
				ExitTime:       exitTime,
				DepartureTime:  departureTime,
				FillTimeS:      fillTimeSec,
				DrainTimeS:     drainTimeSec,
				WaitTimeS:      waitS,
				WaterUsedM3:    waterUsed,
				Regime:         "transitional",
			}
			route.GateSequence = append(route.GateSequence, gateSched)

			tr := timeRange{Start: fillDrainStart, End: exitTime}
			gateUsed[gid] = insertSorted(gateUsed[gid], tr)
			gateTotalTime[gid] += fillTimeSec + drainTimeSec + entryTimeS + exitTimeS

			route.TotalWaitTimeS += waitS
			route.TotalTravelTimeS += travelTimeS
			route.TotalPassageTimeS += fillTimeSec + drainTimeSec + entryTimeS + exitTimeS
			route.TotalWaterUsedM3 += waterUsed

			currentTime = departureTime
		}
		if shipTruncated {
			conflictCount++
		}
		routeMap[ship.ShipID] = route
	}

	routes := make([]models.MultiStageShipRoute, 0, len(routeMap))
	for _, r := range routeMap {
		routes = append(routes, *r)
	}
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Priority > routes[j].Priority ||
			(routes[i].Priority == routes[j].Priority &&
				len(routes[i].GateSequence) > 0 && len(routes[j].GateSequence) > 0 &&
				routes[i].GateSequence[0].ArrivalTime.Before(routes[j].GateSequence[0].ArrivalTime))
	})

	windowStart := time.Now()
	windowEnd := windowStart.Add(24 * time.Hour)
	gateUtilization := map[uint]float64{}
	for gid, t := range gateTotalTime {
		total := windowEnd.Sub(windowStart).Seconds()
		if total > 0 {
			gateUtilization[gid] = t / total
			if gateUtilization[gid] > 1 {
				gateUtilization[gid] = 1
			}
		}
	}

	totalTimeSpan := totalWait + totalTravel
	throughputPerDay := 0.0
	if totalTimeSpan > 0 {
		throughputPerDay = float64(len(routes)) * 86400.0 / totalTimeSpan
	}
	fitness := 0.0
	if totalWait > 0 {
		fitness = 1e6 / (totalWait + 1)
	}

	return &models.MultiStageOptimizeResult{
		Routes:           routes,
		TotalWaitTimeS:   totalWait,
		TotalTravelTimeS: totalTravel,
		TotalWaterUsedM3: totalWater,
		ThroughputShips:  len(routes),
		ThroughputPerDay: throughputPerDay,
		GateUtilization:  gateUtilization,
		ConflictCount:    conflictCount,
		Fitness:          fitness,
		Generations:      1,
	}
}
