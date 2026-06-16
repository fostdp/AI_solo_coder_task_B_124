package vessel_analyzer

import (
	"math"
	"sync"

	"lingqu-dou-gate/internal/config"
	"lingqu-dou-gate/internal/models"
	"lingqu-dou-gate/internal/modules/hydraulic_sim"
)

type VesselAnalyzer struct {
	mu           sync.RWMutex
	hydraulicSim *hydraulic_sim.HydraulicSimulator
	params       config.HydraulicJSONConfig
}

func NewVesselAnalyzer(hydro *hydraulic_sim.HydraulicSimulator) *VesselAnalyzer {
	return &VesselAnalyzer{
		hydraulicSim: hydro,
		params:       config.AppConfig.HydraulicJSON,
	}
}

func (va *VesselAnalyzer) ReloadConfig() {
	va.mu.Lock()
	defer va.mu.Unlock()
	va.params = config.AppConfig.HydraulicJSON
}

func (va *VesselAnalyzer) Analyze(
	gate models.DouGate, waterUp, waterDown, opening float64,
	shipTypes []models.ShipType,
) []*models.ShipTypeEfficiencyReport {
	allSpecs := models.GetShipTypeSpecs()
	filter := map[models.ShipType]bool{}
	if len(shipTypes) > 0 {
		for _, t := range shipTypes {
			filter[t] = true
		}
	}
	reports := make([]*models.ShipTypeEfficiencyReport, 0)
	for _, spec := range allSpecs {
		if len(shipTypes) > 0 && !filter[spec.ShipType] {
			continue
		}
		simReq := hydraulic_sim.SimulateRequest{
			Gate:           gate,
			WaterLevelUp:   waterUp,
			WaterLevelDown: waterDown,
			GateOpening:    opening,
			Direction:      "upstream",
			ShipType:       spec,
		}
		result := va.hydraulicSim.SimulateSync(simReq)
		totalCycle := result.TotalPassageS
		waitBase := 420.0 + float64(spec.BasePriority)*30
		avgWaitTime := waitBase + totalCycle*0.18
		waterPerTon := 0.0
		if spec.CapacityTon > 0 {
			waterPerTon = result.TotalWaterVolume / spec.CapacityTon
		}
		var throughputTon float64
		if totalCycle > 0 {
			throughputTon = spec.CapacityTon * (86400.0 / totalCycle) * 0.8
		}
		conflictRate := 0.0
		occupancy := spec.ChamberOccupancy
		if occupancy <= 0 {
			occupancy = spec.LengthMax * spec.WidthMax / (gate.ChamberLength * gate.ChamberWidth)
			if occupancy > 1 {
				occupancy = 1
			}
		}
		if occupancy > 0.55 {
			conflictRate = 0.25 + (occupancy-0.55)*1.2
		} else if occupancy > 0.30 {
			conflictRate = 0.08 + (occupancy-0.30)*0.68
		} else if occupancy > 0.10 {
			conflictRate = 0.02 + (occupancy-0.10)*0.3
		} else {
			conflictRate = occupancy * 0.2
		}
		if conflictRate > 0.6 {
			conflictRate = 0.6
		}
		resistancePenalty := 1.0 + spec.ResistanceCoeff*occupancy*10
		effIdx := 0.0
		if totalCycle > 0 {
			timeEff := 1 - math.Min(1, totalCycle/2400)
			waterEff := 1 - math.Min(1, waterPerTon/20)
			priorityBoost := float64(spec.BasePriority) / 6.0
			effIdx = (timeEff*40 + waterEff*35 + priorityBoost*25) / resistancePenalty
		}
		if effIdx > 100 {
			effIdx = 100
		}
		reports = append(reports, &models.ShipTypeEfficiencyReport{
			ShipType:         spec.ShipType,
			TypeName:         spec.TypeName,
			AvgWaitTimeS:     avgWaitTime,
			AvgPassageTimeS:  totalCycle,
			AvgWaterPerTon:   waterPerTon,
			AvgThroughputTon: throughputTon,
			ConflictRate:     conflictRate,
			SampleCount:      120,
			EfficiencyIndex:  math.Round(effIdx*100) / 100,
		})
	}
	return reports
}
