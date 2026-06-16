package vr_lock_experience

import (
	"sync"
	"time"

	"lingqu-dou-gate/internal/config"
	"lingqu-dou-gate/internal/models"
	"lingqu-dou-gate/internal/modules/hydraulic_sim"
)

type VRScenario struct {
	ScenarioID    string          `json:"scenario_id"`
	ScenarioName  string          `json:"scenario_name"`
	Dynasty       models.Dynasty  `json:"dynasty"`
	ShipType      models.ShipType `json:"ship_type"`
	Direction     string          `json:"direction"`
	WaterLevelUp  float64         `json:"water_level_up"`
	WaterLevelDown float64        `json:"water_level_down"`
	Description   string          `json:"description"`
}

type VRSession struct {
	SessionID   string    `json:"session_id"`
	UserID      string    `json:"user_id,omitempty"`
	Scenario    VRScenario `json:"scenario"`
	StartTime   time.Time `json:"start_time"`
	State       string    `json:"state"`
	ScoreTime   int       `json:"score_time"`
	ScoreSafety int       `json:"score_safety"`
	ScoreWater  int       `json:"score_water"`
	ScoreTotal  int       `json:"score_total"`
	Errors      []string  `json:"errors"`
}

type VRLockExperience struct {
	mu           sync.RWMutex
	hydraulicSim *hydraulic_sim.HydraulicSimulator
	params       config.HydraulicJSONConfig
	scenarios    []VRScenario
	sessions     map[string]*VRSession
}

func NewVRLockExperience(hydro *hydraulic_sim.HydraulicSimulator) *VRLockExperience {
	ve := &VRLockExperience{
		hydraulicSim: hydro,
		params:       config.AppConfig.HydraulicJSON,
		scenarios: []VRScenario{
			{
				ScenarioID: "tang_upstream",
				ScenarioName: "唐代陡门上行体验",
				Dynasty:      models.DynastyTang,
				ShipType:     models.ShipTypeGrain,
				Direction:    "upstream",
				WaterLevelUp: 7.0,
				WaterLevelDown: 5.5,
				Description:   "唐代土石木叠梁陡门，体验古代漕船跨越高差2米过闸的操作流程",
			},
			{
				ScenarioID: "song_downstream",
				ScenarioName: "宋代条石陡门下行体验",
				Dynasty:      models.DynastySong,
				ShipType:     models.ShipTypeCargo,
				Direction:    "downstream",
				WaterLevelUp: 7.5,
				WaterLevelDown: 4.8,
				Description:   "宋代包铁条石陡门，下行货船快速通航体验",
			},
			{
				ScenarioID: "qing_royal",
				ScenarioName: "清代御舟通航体验",
				Dynasty:      models.DynastyQing,
				ShipType:     models.ShipTypeRoyal,
				Direction:    "upstream",
				WaterLevelUp: 7.8,
				WaterLevelDown: 4.2,
				Description:   "清代糯米灰浆砌青石陡门，御舟过闸高规格体验",
			},
			{
				ScenarioID: "modern_fishing",
				ScenarioName: "现代渔船快速通航",
				Dynasty:      models.DynastyModern,
				ShipType:     models.ShipTypeFishing,
				Direction:    "upstream",
				WaterLevelUp: 7.5,
				WaterLevelDown: 3.6,
				Description:   "现代液压卷扬机陡门，渔船高效通过体验",
			},
		},
		sessions: map[string]*VRSession{},
	}
	return ve
}

func (ve *VRLockExperience) ReloadConfig() {
	ve.mu.Lock()
	defer ve.mu.Unlock()
	ve.params = config.AppConfig.HydraulicJSON
}

func (ve *VRLockExperience) ListScenarios() []VRScenario {
	ve.mu.RLock()
	defer ve.mu.RUnlock()
	return ve.scenarios
}

func (ve *VRLockExperience) GetScenario(id string) (*VRScenario, bool) {
	ve.mu.RLock()
	defer ve.mu.RUnlock()
	for _, s := range ve.scenarios {
		if s.ScenarioID == id {
			copy := s
			return &copy, true
		}
	}
	return nil, false
}

func (ve *VRLockExperience) NewSession(scenarioID, userID string) (*VRSession, error) {
	scenario, ok := ve.GetScenario(scenarioID)
	if !ok {
		scenario = &ve.scenarios[0]
	}
	sess := &VRSession{
		SessionID:   "vr_" + time.Now().Format("20060102150405") + "_" + randStr(6),
		UserID:      userID,
		Scenario:    *scenario,
		StartTime:   time.Now(),
		State:       "idle",
		ScoreTime:   30,
		ScoreSafety: 40,
		ScoreWater:  30,
		ScoreTotal:  100,
		Errors:      []string{},
	}
	ve.mu.Lock()
	ve.sessions[sess.SessionID] = sess
	ve.mu.Unlock()
	return sess, nil
}

func (ve *VRLockExperience) GetSession(sessionID string) (*VRSession, bool) {
	ve.mu.RLock()
	defer ve.mu.RUnlock()
	s, ok := ve.sessions[sessionID]
	return s, ok
}

func (ve *VRLockExperience) SimulateScenario(scenarioID string) (*hydraulic_sim.SimulateResult, error) {
	scenario, ok := ve.GetScenario(scenarioID)
	if !ok {
		return &hydraulic_sim.SimulateResult{}, nil
	}
	dynastyDesigns := models.GetDynastyDesigns(1)
	var design *models.DynastyDesign
	for _, d := range dynastyDesigns {
		if d.Dynasty == scenario.Dynasty {
			design = d
			break
		}
	}
	var shipSpec *models.ShipTypeSpec
	for _, s := range models.GetShipTypeSpecs() {
		if s.ShipType == scenario.ShipType {
			shipSpec = s
			break
		}
	}
	gate := models.DouGate{
		ID: 1, Name: "虚拟陡门",
		GateWidth: 6, GateHeight: 4.5, ChamberLength: 60, ChamberWidth: 7,
		MaxWaterLevelUp: scenario.WaterLevelUp, MinWaterLevelDown: scenario.WaterLevelDown,
	}
	req := hydraulic_sim.SimulateRequest{
		Gate:           gate,
		WaterLevelUp:   scenario.WaterLevelUp,
		WaterLevelDown: scenario.WaterLevelDown,
		GateOpening:    0.85,
		Direction:      scenario.Direction,
		DynastyDesign:  design,
		ShipType:       shipSpec,
	}
	return ve.hydraulicSim.SimulateSync(req), nil
}

func randStr(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[int(time.Now().UnixNano()>>uint(i*3))%len(letters)]
	}
	return string(b)
}
