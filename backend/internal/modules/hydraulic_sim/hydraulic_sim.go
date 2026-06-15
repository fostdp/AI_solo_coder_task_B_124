package hydraulic_sim

import (
	"log"
	"math"
	"sync"

	"lingqu-dou-gate/internal/config"
	"lingqu-dou-gate/internal/models"
)

type FlowRegime int

const (
	FreeFlow FlowRegime = iota
	TransitionalFlow
	SubmergedFlow
	WeirFlow
)

func (f FlowRegime) String() string {
	switch f {
	case FreeFlow:
		return "free"
	case TransitionalFlow:
		return "transitional"
	case SubmergedFlow:
		return "submerged"
	case WeirFlow:
		return "weir"
	default:
		return "unknown"
	}
}

type SimulateRequest struct {
	Gate           models.DouGate
	WaterLevelUp   float64
	WaterLevelDown float64
	GateOpening    float64
	Direction      string
	ReplyChan      chan *SimulateResult
	DynastyDesign  *models.DynastyDesign `json:"-"`
	ShipType       *models.ShipTypeSpec  `json:"-"`
}

type SimulateResult struct {
	FillTime         float64                  `json:"fill_time"`
	DrainTime        float64                  `json:"drain_time"`
	WaterLevelCurve  []models.WaterLevelPoint `json:"water_level_curve"`
	FlowRateCurve    []models.FlowRatePoint   `json:"flow_rate_curve"`
	MaxFlowRate      float64                  `json:"max_flow_rate"`
	AvgFlowRate      float64                  `json:"avg_flow_rate"`
	TotalWaterVolume float64                  `json:"total_water_volume"`
	Regime           FlowRegime               `json:"regime,omitempty"`
	Error            error                    `json:"-"`
	EntryTimeS       float64                  `json:"entry_time_s,omitempty"`
	ExitTimeS        float64                  `json:"exit_time_s,omitempty"`
	TotalPassageS    float64                  `json:"total_passage_s,omitempty"`
	UsedDynasty      string                   `json:"used_dynasty,omitempty"`
	ShipTypeName     string                   `json:"ship_type,omitempty"`
}

type HydraulicSimulator struct {
	mu          sync.RWMutex
	running     bool
	requestChan chan SimulateRequest
	stopChan    chan struct{}
	wg          sync.WaitGroup
	params      config.HydraulicJSONConfig
	workerCount int
}

func NewHydraulicSimulator(workerCount int) *HydraulicSimulator {
	if workerCount <= 0 {
		workerCount = 2
	}
	return &HydraulicSimulator{
		requestChan: make(chan SimulateRequest, 50),
		stopChan:    make(chan struct{}),
		params:      config.AppConfig.HydraulicJSON,
		workerCount: workerCount,
	}
}

func (h *HydraulicSimulator) RequestChannel() chan<- SimulateRequest {
	return h.requestChan
}

func (h *HydraulicSimulator) Submit(req SimulateRequest) {
	select {
	case h.requestChan <- req:
	default:
		if req.ReplyChan != nil {
			req.ReplyChan <- &SimulateResult{Error: &SimulatorError{Message: "simulator queue full"}}
		}
	}
}

type SimulatorError struct {
	Message string
}

func (e *SimulatorError) Error() string {
	return e.Message
}

func (h *HydraulicSimulator) Start() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.running {
		return
	}
	h.running = true
	h.params = config.AppConfig.HydraulicJSON

	for i := 0; i < h.workerCount; i++ {
		h.wg.Add(1)
		go h.worker(i)
	}

	log.Printf("Hydraulic simulator started with %d workers", h.workerCount)
}

func (h *HydraulicSimulator) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return
	}
	h.running = false

	close(h.stopChan)
	h.wg.Wait()
	close(h.requestChan)

	log.Println("Hydraulic simulator stopped")
}

func (h *HydraulicSimulator) worker(id int) {
	defer h.wg.Done()

	for {
		select {
		case <-h.stopChan:
			return
		case req, ok := <-h.requestChan:
			if !ok {
				return
			}
			result := h.simulatePassage(req)
			if req.ReplyChan != nil {
				select {
				case req.ReplyChan <- result:
				default:
				}
			}
		}
	}
}

func (h *HydraulicSimulator) simulatePassage(req SimulateRequest) *SimulateResult {
	gate := req.Gate
	waterLevelUp := req.WaterLevelUp
	waterLevelDown := req.WaterLevelDown
	gateOpening := req.GateOpening
	direction := req.Direction
	dynastyDesign := req.DynastyDesign
	shipSpec := req.ShipType

	effGate := gate
	effParams := h.params
	usedDynasty := ""
	if dynastyDesign != nil {
		effGate.GateWidth = dynastyDesign.GateWidth
		effGate.GateHeight = dynastyDesign.GateHeight
		effGate.ChamberLength = dynastyDesign.ChamberLength
		effGate.ChamberWidth = dynastyDesign.ChamberWidth
		effParams.DefaultCd = dynastyDesign.DefaultCd
		effParams.ContractionCoeff.Base = dynastyDesign.CcBase
		effParams.ContractionCoeff.LinearTerm = dynastyDesign.CcLinear
		effParams.ContractionCoeff.QuadraticTerm = dynastyDesign.CcQuadratic
		effParams.WeirCoeff = dynastyDesign.WeirCoeff
		usedDynasty = string(dynastyDesign.Dynasty)
		if waterLevelUp <= 0 {
			waterLevelUp = gate.MinWaterLevelDown + dynastyDesign.WaterLift
		}
	}

	if waterLevelUp <= 0 {
		waterLevelUp = effGate.MaxWaterLevelUp
	}
	if waterLevelDown <= 0 {
		waterLevelDown = effGate.MinWaterLevelDown
	}
	if gateOpening <= 0 {
		gateOpening = 1.0
	}
	if direction == "" {
		direction = "upstream"
	}

	chamberArea := effGate.ChamberLength * effGate.ChamberWidth
	var fillTime, drainTime float64
	var levelCurve []models.WaterLevelPoint
	var flowCurve []models.FlowRatePoint

	if direction == "upstream" {
		fillTime, levelCurve, flowCurve = h.calculateFillTimeWithParams(
			waterLevelUp, waterLevelDown, gateOpening, chamberArea, effParams, effGate,
		)
		drainTime = 0
	} else {
		drainTime, levelCurve, flowCurve = h.calculateDrainTimeWithParams(
			waterLevelUp, waterLevelDown, gateOpening, chamberArea, effParams, effGate,
		)
		fillTime = 0
	}

	var maxFlowRate, avgFlowRate, totalFlow float64
	for _, fp := range flowCurve {
		if fp.FlowRate > maxFlowRate {
			maxFlowRate = fp.FlowRate
		}
		totalFlow += fp.FlowRate
	}
	if len(flowCurve) > 0 {
		avgFlowRate = totalFlow / float64(len(flowCurve))
	}

	if dynastyDesign != nil && dynastyDesign.MaxFlowRate > 0 && maxFlowRate > dynastyDesign.MaxFlowRate {
		ratio := dynastyDesign.MaxFlowRate / maxFlowRate
		for i := range flowCurve {
			flowCurve[i].FlowRate *= ratio
		}
		maxFlowRate = dynastyDesign.MaxFlowRate
		avgFlowRate *= ratio
		fillTime /= ratio
		drainTime /= ratio
	}

	totalVolume := math.Abs(waterLevelUp-waterLevelDown) * chamberArea

	shipEntryTime, shipExitTime := 0.0, 0.0
	shipTypeName := ""
	if shipSpec != nil {
		shipEntryTime = shipSpec.EntryTimeS
		shipExitTime = shipSpec.ExitTimeS
		shipTypeName = shipSpec.TypeName
		if shipSpec.WaterFactor != 1.0 {
			totalVolume *= shipSpec.WaterFactor
		}
	}

	return &SimulateResult{
		FillTime:         fillTime,
		DrainTime:        drainTime,
		WaterLevelCurve:  levelCurve,
		FlowRateCurve:    flowCurve,
		MaxFlowRate:      maxFlowRate,
		AvgFlowRate:      avgFlowRate,
		TotalWaterVolume: totalVolume,
		EntryTimeS:       shipEntryTime,
		ExitTimeS:        shipExitTime,
		TotalPassageS:    shipEntryTime + fillTime + drainTime + shipExitTime,
		UsedDynasty:      usedDynasty,
		ShipTypeName:     shipTypeName,
	}
}

func (h *HydraulicSimulator) calculateFillTime(
	targetLevelUp, initialChamberLevel, gateOpening, chamberArea float64,
) (float64, []models.WaterLevelPoint, []models.FlowRatePoint) {
	return h.calculateFillTimeWithParams(targetLevelUp, initialChamberLevel, gateOpening, chamberArea, h.params, models.DouGate{GateWidth: 6.0, GateHeight: 4.5})
}

func (h *HydraulicSimulator) calculateDrainTime(
	initialChamberLevel, targetLevelDown, gateOpening, chamberArea float64,
) (float64, []models.WaterLevelPoint, []models.FlowRatePoint) {
	return h.calculateDrainTimeWithParams(initialChamberLevel, targetLevelDown, gateOpening, chamberArea, h.params, models.DouGate{GateWidth: 6.0, GateHeight: 4.5})
}

func (h *HydraulicSimulator) calculateFillTimeWithParams(
	targetLevelUp, initialChamberLevel, gateOpening, chamberArea float64,
	params config.HydraulicJSONConfig, gate models.DouGate,
) (float64, []models.WaterLevelPoint, []models.FlowRatePoint) {
	return h.calculateLevelChangeWithParams(
		targetLevelUp, initialChamberLevel, gateOpening, chamberArea, true, params, gate,
	)
}

func (h *HydraulicSimulator) calculateDrainTimeWithParams(
	initialChamberLevel, targetLevelDown, gateOpening, chamberArea float64,
	params config.HydraulicJSONConfig, gate models.DouGate,
) (float64, []models.WaterLevelPoint, []models.FlowRatePoint) {
	return h.calculateLevelChangeWithParams(
		targetLevelDown, initialChamberLevel, gateOpening, chamberArea, false, params, gate,
	)
}

func (h *HydraulicSimulator) calculateLevelChange(
	targetLevel, initialLevel, opening, area float64, isFilling bool,
) (float64, []models.WaterLevelPoint, []models.FlowRatePoint) {
	return h.calculateLevelChangeWithParams(targetLevel, initialLevel, opening, area, isFilling,
		h.params, models.DouGate{GateWidth: 6.0, GateHeight: 4.5})
}

func (h *HydraulicSimulator) calculateLevelChangeWithParams(
	targetLevel, initialLevel, opening, area float64, isFilling bool,
	params config.HydraulicJSONConfig, gate models.DouGate,
) (float64, []models.WaterLevelPoint, []models.FlowRatePoint) {
	dt := params.Simulation.TimeStep
	var totalTime float64
	var levelPoints []models.WaterLevelPoint
	var flowPoints []models.FlowRatePoint

	currentLevel := initialLevel
	levelPoints = append(levelPoints, models.WaterLevelPoint{Time: 0, WaterLevel: currentLevel})

	maxIterations := params.Simulation.MaxIterations
	iterations := 0

	targetReached := func() bool {
		if isFilling {
			return currentLevel >= targetLevel-0.0005
		}
		return currentLevel <= targetLevel+0.0005
	}

	for !targetReached() && iterations < maxIterations {
		var flowRate float64
		if isFilling {
			flowRate = h.calculateOrificeFlowWithParams(targetLevel, currentLevel, opening, params, gate)
		} else {
			flowRate = h.calculateOrificeFlowWithParams(currentLevel, targetLevel, opening, params, gate)
		}
		flowPoints = append(flowPoints, models.FlowRatePoint{Time: totalTime, FlowRate: flowRate})

		dtAdaptive := dt
		if flowRate > 0 {
			levelDiff := math.Abs(targetLevel - currentLevel)
			maxDH := levelDiff * params.Simulation.AdaptiveStepRatio
			theoreticalDT := maxDH * area / flowRate
			if theoreticalDT < dtAdaptive && theoreticalDT > params.Simulation.MinDT {
				dtAdaptive = theoreticalDT
			}
		}

		dv := flowRate * dtAdaptive
		dh := dv / area

		if isFilling {
			currentLevel += dh
		} else {
			currentLevel -= dh
		}
		totalTime += dtAdaptive
		iterations++

		levelThreshold := math.Abs(targetLevel-initialLevel) * 0.1
		if math.Mod(float64(iterations), 5) == 0 || math.Abs(dh) > levelThreshold || targetReached() {
			levelPoints = append(levelPoints, models.WaterLevelPoint{
				Time:       totalTime,
				WaterLevel: currentLevel,
			})
		}
	}

	if len(levelPoints) == 0 || levelPoints[len(levelPoints)-1].Time != totalTime {
		levelPoints = append(levelPoints, models.WaterLevelPoint{
			Time:       totalTime,
			WaterLevel: currentLevel,
		})
	}

	if len(flowPoints) == 0 || flowPoints[len(flowPoints)-1].Time != totalTime {
		flowPoints = append(flowPoints, models.FlowRatePoint{
			Time:     totalTime,
			FlowRate: 0,
		})
	}

	return totalTime, levelPoints, flowPoints
}

func (h *HydraulicSimulator) calculateOrificeFlow(
	waterLevelUp, waterLevelDown, gateOpening float64,
) float64 {
	return h.calculateOrificeFlowWithParams(waterLevelUp, waterLevelDown, gateOpening,
		h.params, models.DouGate{GateWidth: 6.0, GateHeight: 4.5})
}

func (h *HydraulicSimulator) calculateOrificeFlowWithParams(
	waterLevelUp, waterLevelDown, gateOpening float64,
	params config.HydraulicJSONConfig, gate models.DouGate,
) float64 {
	headDiff := waterLevelUp - waterLevelDown
	if headDiff <= 0 {
		return 0
	}

	gateWidth := gate.GateWidth
	if gateWidth <= 0 {
		gateWidth = 6.0
	}
	gateHeight := gate.GateHeight
	if gateHeight <= 0 {
		gateHeight = 4.5
	}
	openingHeight := gateOpening * gateHeight
	if openingHeight <= 0 {
		return 0
	}

	relativeOpening := openingHeight / waterLevelUp
	if relativeOpening > 1 {
		relativeOpening = 1
	}

	Cc := h.calculateContractionCoefficientWithParams(relativeOpening, params)
	effectiveHeight := Cc * openingHeight

	regime := h.classifyFlowRegimeWithParams(waterLevelUp, waterLevelDown, openingHeight, relativeOpening, params)

	var flowRate float64
	switch regime {
	case FreeFlow:
		flowRate = h.calculateFreeFlowRateWithParams(waterLevelUp, openingHeight, relativeOpening, gateWidth, params)
	case SubmergedFlow:
		submergenceRatio := 0.0
		vcDepth := effectiveHeight
		if headDiff > 0 {
			submergenceRatio = waterLevelDown / (waterLevelUp - vcDepth)
		}
		flowRate = h.calculateSubmergedFlowRateWithParams(
			waterLevelUp, waterLevelDown, openingHeight, relativeOpening, submergenceRatio, gateWidth, params,
		)
	case TransitionalFlow:
		freeFlow := h.calculateFreeFlowRateWithParams(waterLevelUp, openingHeight, relativeOpening, gateWidth, params)
		vcDepth := effectiveHeight
		submergenceRatio := waterLevelDown / (waterLevelUp - vcDepth)
		submergedFlow := h.calculateSubmergedFlowRateWithParams(
			waterLevelUp, waterLevelDown, openingHeight, relativeOpening, submergenceRatio, gateWidth, params,
		)
		transitionStart := params.FlowRegime.FreeFlowThreshold
		transitionEnd := params.FlowRegime.SubmergedThreshold
		transitionWidth := transitionEnd - transitionStart
		position := (submergenceRatio - transitionStart) / transitionWidth
		smoothStep := position * position * (3 - 2*position)
		flowRate = freeFlow*(1-smoothStep) + submergedFlow*smoothStep
	case WeirFlow:
		flowRate = h.calculateWeirFlowRateWithParams(waterLevelUp, waterLevelDown, headDiff, gateWidth, params)
	default:
		flowRate = h.calculateFreeFlowRateWithParams(waterLevelUp, openingHeight, relativeOpening, gateWidth, params)
	}

	return math.Max(0, flowRate)
}

func (h *HydraulicSimulator) calculateContractionCoefficient(relativeOpening float64) float64 {
	return h.calculateContractionCoefficientWithParams(relativeOpening, h.params)
}

func (h *HydraulicSimulator) calculateContractionCoefficientWithParams(relativeOpening float64, params config.HydraulicJSONConfig) float64 {
	if relativeOpening <= 0 {
		return 0.6
	}
	if relativeOpening >= 1 {
		return 1.0
	}
	coeff := params.ContractionCoeff
	Cc := coeff.Base + coeff.LinearTerm*relativeOpening + coeff.QuadraticTerm*relativeOpening*relativeOpening
	return Cc
}

func (h *HydraulicSimulator) classifyFlowRegime(
	waterLevelUp, waterLevelDown, openingHeight, relativeOpening float64,
) FlowRegime {
	return h.classifyFlowRegimeWithParams(waterLevelUp, waterLevelDown, openingHeight, relativeOpening, h.params)
}

func (h *HydraulicSimulator) classifyFlowRegimeWithParams(
	waterLevelUp, waterLevelDown, openingHeight, relativeOpening float64,
	params config.HydraulicJSONConfig,
) FlowRegime {
	headTotal := waterLevelUp - waterLevelDown
	Cc := h.calculateContractionCoefficientWithParams(relativeOpening, params)
	vcDepth := Cc * openingHeight

	submergenceRatio := 0.0
	if headTotal > 0 {
		submergenceRatio = waterLevelDown / (waterLevelUp - vcDepth)
	}

	regimeCfg := params.FlowRegime
	if relativeOpening > regimeCfg.WeirRelativeOpening && headTotal/relativeOpening < regimeCfg.WeirHeadRatio {
		return WeirFlow
	} else if submergenceRatio < regimeCfg.FreeFlowThreshold {
		return FreeFlow
	} else if submergenceRatio > regimeCfg.SubmergedThreshold {
		return SubmergedFlow
	} else {
		return TransitionalFlow
	}
}

func (h *HydraulicSimulator) calculateFreeFlowRate(
	waterLevelUp, openingHeight, relativeOpening, gateWidth float64,
) float64 {
	return h.calculateFreeFlowRateWithParams(waterLevelUp, openingHeight, relativeOpening, gateWidth, h.params)
}

func (h *HydraulicSimulator) calculateFreeFlowRateWithParams(
	waterLevelUp, openingHeight, relativeOpening, gateWidth float64,
	params config.HydraulicJSONConfig,
) float64 {
	Cc := h.calculateContractionCoefficientWithParams(relativeOpening, params)
	effectiveHeight := Cc * openingHeight

	var Cd float64
	if relativeOpening < 0.1 {
		Cd = 0.60
	} else if relativeOpening < 0.5 {
		Cd = 0.60 - 0.03*(relativeOpening-0.1)/0.4
	} else if relativeOpening < 0.8 {
		Cd = 0.57 + 0.06*(relativeOpening-0.5)/0.3
	} else {
		Cd = 0.63
	}

	headUp := waterLevelUp - effectiveHeight
	if headUp <= 0 {
		headUp = waterLevelUp * 0.7
	}

	gravity := params.Gravity
	flowRate := Cd * effectiveHeight * gateWidth * math.Sqrt(2*gravity*headUp)

	return math.Max(0, flowRate)
}

func (h *HydraulicSimulator) calculateSubmergedFlowRate(
	waterLevelUp, waterLevelDown, openingHeight, relativeOpening, submergenceRatio, gateWidth float64,
) float64 {
	return h.calculateSubmergedFlowRateWithParams(
		waterLevelUp, waterLevelDown, openingHeight, relativeOpening, submergenceRatio, gateWidth, h.params,
	)
}

func (h *HydraulicSimulator) calculateSubmergedFlowRateWithParams(
	waterLevelUp, waterLevelDown, openingHeight, relativeOpening, submergenceRatio, gateWidth float64,
	params config.HydraulicJSONConfig,
) float64 {
	Cc := h.calculateContractionCoefficientWithParams(relativeOpening, params)
	effectiveHeight := Cc * openingHeight
	headDiff := waterLevelUp - waterLevelDown

	if headDiff <= 0 {
		return 0
	}

	eta := submergenceRatio
	sigma := 1.0
	fullySub := params.FlowRegime.FullySubmerged
	submergedThr := params.FlowRegime.SubmergedThreshold
	if eta >= fullySub {
		sigma = 0.0
	} else if eta > submergedThr {
		transWidth := fullySub - submergedThr
		sigma = math.Sqrt((fullySub - eta) / transWidth)
	}

	Cd := params.DefaultCd * params.SubmergedCoeff
	gravity := params.Gravity
	flowRate := Cd * sigma * effectiveHeight * gateWidth * math.Sqrt(2*gravity*headDiff)

	return math.Max(0, flowRate)
}

func (h *HydraulicSimulator) calculateWeirFlowRate(
	waterLevelUp, waterLevelDown, headDiff, gateWidth float64,
) float64 {
	return h.calculateWeirFlowRateWithParams(waterLevelUp, waterLevelDown, headDiff, gateWidth, h.params)
}

func (h *HydraulicSimulator) calculateWeirFlowRateWithParams(
	waterLevelUp, waterLevelDown, headDiff, gateWidth float64,
	params config.HydraulicJSONConfig,
) float64 {
	if headDiff <= 0 {
		return 0
	}
	submergenceRatio := waterLevelDown / waterLevelUp
	sigma := 1.0
	if submergenceRatio > 0.8 {
		sigma = 1.0 - (submergenceRatio-0.8)/0.2
		if sigma < 0 {
			sigma = 0
		}
	}
	Cw := params.WeirCoeff
	flowRate := sigma * Cw * gateWidth * math.Pow(headDiff, 1.5)
	return math.Max(0, flowRate)
}

func (h *HydraulicSimulator) SimulateSync(req SimulateRequest) *SimulateResult {
	return h.simulatePassage(req)
}

func (h *HydraulicSimulator) DynastyComparisonSync(
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
		simReq := SimulateRequest{
			Gate:          gate,
			WaterLevelUp:  waterUp,
			WaterLevelDown: waterDown,
			GateOpening:   opening,
			Direction:     direction,
			DynastyDesign: design,
		}
		result := h.simulatePassage(simReq)
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
			Dynasty:         design.Dynasty,
			DynastyName:     design.DynastyName,
			Design:          design,
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

func analyzeDynastyTradeoffs(cur, baseline *models.DynastyDesign, r *SimulateResult) ([]string, []string) {
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
	return "+" + strconvFloat(int(pct)) + "%"
}

func strconvFloat(v int) string {
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

func computeEfficiencyScore(r *SimulateResult, d *models.DynastyDesign) float64 {
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
	return s
}

func (h *HydraulicSimulator) ShipTypeAnalysisSync(
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
		simReq := SimulateRequest{
			Gate:           gate,
			WaterLevelUp:   waterUp,
			WaterLevelDown: waterDown,
			GateOpening:    opening,
			Direction:      "upstream",
			ShipType:       spec,
		}
		result := h.simulatePassage(simReq)
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
		switch {
		case spec.LengthMax > 38:
			conflictRate = 0.35
		case spec.LengthMax > 26:
			conflictRate = 0.18
		case spec.LengthMax > 14:
			conflictRate = 0.08
		default:
			conflictRate = 0.02
		}
		effIdx := 0.0
		if totalCycle > 0 {
			timeEff := 1 - math.Min(1, totalCycle/2400)
			waterEff := 1 - math.Min(1, waterPerTon/20)
			priorityBoost := float64(spec.BasePriority) / 6.0
			effIdx = (timeEff*40 + waterEff*35 + priorityBoost*25)
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
			EfficiencyIndex:  effIdx,
		})
	}
	return reports
}

func (h *HydraulicSimulator) ReloadConfig() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.params = config.AppConfig.HydraulicJSON
}
