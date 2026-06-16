package design_comparator

import (
	"math"
	"sync"

	"lingqu-dou-gate/internal/config"
	"lingqu-dou-gate/internal/models"
	"lingqu-dou-gate/internal/modules/hydraulic_sim"
)

type DesignComparator struct {
	mu            sync.RWMutex
	hydraulicSim  *hydraulic_sim.HydraulicSimulator
	params        config.HydraulicJSONConfig
}

func NewDesignComparator(hydro *hydraulic_sim.HydraulicSimulator) *DesignComparator {
	return &DesignComparator{
		hydraulicSim: hydro,
		params:       config.AppConfig.HydraulicJSON,
	}
}

func (dc *DesignComparator) ReloadConfig() {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.params = config.AppConfig.HydraulicJSON
}

func (dc *DesignComparator) Compare(
	gate models.DouGate, waterUp, waterDown, opening float64, direction string,
	dynasties []models.Dynasty,
) []*models.DynastyComparisonResult {
	allDesigns := models.GetDynastyDesigns(gate.ID)
	filter := map[models.Dynasty]bool{}
	if len(dynasties) > 0 {
		for _, d := range dynasties {
			filter[d] = true
		}
	}
	results := make([]*models.DynastyComparisonResult, 0)
	for _, design := range allDesigns {
		if len(dynasties) > 0 && !filter[design.Dynasty] {
			continue
		}
		simReq := hydraulic_sim.SimulateRequest{
			Gate:           gate,
			WaterLevelUp:   waterUp,
			WaterLevelDown: waterDown,
			GateOpening:    opening,
			Direction:      direction,
			DynastyDesign:  design,
		}
		result := dc.hydraulicSim.SimulateSync(simReq)
		capTon := 60.0
		if result.TotalWaterVolume > 0 {
			capTon = result.TotalWaterVolume / 10
		}
		waterPerTon := 0.0
		if capTon > 0 {
			waterPerTon = result.TotalWaterVolume / capTon
		}
		cycSec := result.FillTime + result.DrainTime + 600
		var passagesPerDay int
		if cycSec > 0 {
			passagesPerDay = int(86400.0 / cycSec)
		}
		tangDesign := filterDesign(allDesigns, models.DynastyTang)
		advantages, limitations := analyzeDynastyTradeoffs(design, tangDesign, result)
		effScore := computeEfficiencyScore(result, design)
		results = append(results, &models.DynastyComparisonResult{
			Dynasty:     design.Dynasty,
			DynastyName: design.DynastyName,
			Design:      design,
			Simulation: &models.SimulationResultLite{
				FillTime:         result.FillTime,
				DrainTime:        result.DrainTime,
				MaxFlowRate:      result.MaxFlowRate,
				AvgFlowRate:      result.AvgFlowRate,
				TotalWaterVolume: result.TotalWaterVolume,
				Regime:           result.Regime.String(),
			},
			EfficiencyScore: effScore,
			WaterPerTon:     waterPerTon,
			PassagesPerDay:  passagesPerDay,
			Advantages:      advantages,
			Limitations:     limitations,
		})
	}
	return results
}

func filterDesign(designs []*models.DynastyDesign, target models.Dynasty) *models.DynastyDesign {
	for _, d := range designs {
		if d.Dynasty == target {
			return d
		}
	}
	return nil
}

func analyzeDynastyTradeoffs(cur, baseline *models.DynastyDesign, r *hydraulic_sim.SimulateResult) ([]string, []string) {
	advantages := []string{}
	limitations := []string{}
	if baseline != nil {
		if cur.ChamberLength >= baseline.ChamberLength*1.3 {
			advantages = append(advantages, "闸室加长"+ratioPct(cur.ChamberLength, baseline.ChamberLength)+"，可容纳更长船队")
		}
		if cur.GateWidth >= baseline.GateWidth*1.2 {
			advantages = append(advantages, "闸孔宽"+ratioPct(cur.GateWidth, baseline.GateWidth)+"，大船通行更便利")
		}
		if cur.DefaultCd > baseline.DefaultCd*1.03 {
			advantages = append(advantages, "流量系数提升"+ratioPct(cur.DefaultCd, baseline.DefaultCd)+"，充放水更快")
		}
		if cur.WaterLift > baseline.WaterLift*1.2 {
			advantages = append(advantages, "提升水位增加"+ratioPct(cur.WaterLift, baseline.WaterLift)+"，梯级效率更高")
		}
		if cur.MaxFlowRate < baseline.MaxFlowRate*0.8 {
			limitations = append(limitations, "设计流量受限，超大型洪水易溢顶")
		}
	}
	if cur.Material == "土石木混合" {
		limitations = append(limitations, "土石木材质耐久度低，每年需大修")
		advantages = append(advantages, "取材便捷，建造成本低")
	}
	if cur.Structure == "对开立式门+双门闸室" {
		advantages = append(advantages, "双闸室门闭合精度高，泄漏量减少40%")
	}
	if cur.Structure == "重门式+泄水副槽" {
		advantages = append(advantages, "泄水副槽可分流，有效降低涡蚀破坏")
	}
	if cur.HoistType == "液压卷扬机" {
		advantages = append(advantages, "液压启闭8分钟完成，效率较人力提升3倍")
	}
	if cur.HoistType == "人力绞车" {
		limitations = append(limitations, "人力启闭需25-40人协同，单次操作耗时>25分钟")
	}
	return advantages, limitations
}

func ratioPct(a, b float64) string {
	if b <= 0 {
		return ""
	}
	pct := (a/b - 1) * 100
	return "+" + intToStr(int(pct)) + "%"
}

func intToStr(v int) string {
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

func computeEfficiencyScore(r *hydraulic_sim.SimulateResult, d *models.DynastyDesign) float64 {
	timeScore := 0.0
	if r.FillTime+r.DrainTime > 60 {
		timeScore = 30 * 600.0 / (r.FillTime + r.DrainTime)
	}
	volumeScore := 0.0
	if r.TotalWaterVolume > 0 {
		volumeScore = 35 * 1000.0 / r.TotalWaterVolume
	}
	liftScore := 25 * d.WaterLift / 3.5
	materialScore := 10.0
	switch d.Material {
	case "钢筋混凝土+古貌贴面":
		materialScore = 10
	case "糯米灰浆砌青石":
		materialScore = 9
	case "包铁条石":
		materialScore = 7
	case "土石木混合":
		materialScore = 4
	}
	s := timeScore + volumeScore + liftScore + materialScore
	if s > 100 {
		s = 100
	}
	return math.Round(s*100) / 100
}
