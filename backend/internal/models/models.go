package models

import (
	"time"
)

type DouGate struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	Name               string    `json:"name"`
	Location           string    `json:"location"`
	GateWidth          float64   `json:"gate_width"`
	GateHeight         float64   `json:"gate_height"`
	MaxWaterLevelUp    float64   `json:"max_water_level_up"`
	MinWaterLevelUp    float64   `json:"min_water_level_up"`
	MaxWaterLevelDown  float64   `json:"max_water_level_down"`
	MinWaterLevelDown  float64   `json:"min_water_level_down"`
	ChamberLength      float64   `json:"chamber_length"`
	ChamberWidth       float64   `json:"chamber_width"`
	DischargeCoefficient float64 `json:"discharge_coefficient"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
}

type SensorData struct {
	Time          time.Time `gorm:"primaryKey" json:"time"`
	GateID        uint      `gorm:"primaryKey" json:"gate_id"`
	WaterLevelUp  float64   `json:"water_level_up"`
	WaterLevelDown float64  `json:"water_level_down"`
	GateOpening   float64   `json:"gate_opening"`
	FlowRate      float64   `json:"flow_rate"`
	PassageTime   float64   `json:"passage_time"`
	Status        string    `json:"status"`
}

func (SensorData) TableName() string {
	return "sensor_data"
}

type Ship struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `json:"name"`
	Priority    int       `json:"priority"`
	Length      float64   `json:"length"`
	Width       float64   `json:"width"`
	Draft       float64   `json:"draft"`
	ArrivalTime time.Time `json:"arrival_time"`
	Direction   string    `json:"direction"`
	Status      string    `json:"status"`
}

type PassageRecord struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ShipID     uint      `json:"ship_id"`
	GateID     uint      `json:"gate_id"`
	EntryTime  time.Time `json:"entry_time"`
	ExitTime   time.Time `json:"exit_time"`
	FillTime   float64   `json:"fill_time"`
	DrainTime  float64   `json:"drain_time"`
	TotalTime  float64   `json:"total_time"`
	WaitTime   float64   `json:"wait_time"`
	Status     string    `json:"status"`
}

type Alert struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Time       time.Time `json:"time"`
	GateID     uint      `json:"gate_id"`
	AlertType  string    `json:"alert_type"`
	Severity   string    `json:"severity"`
	Message    string    `json:"message"`
	Resolved   bool      `json:"resolved"`
	ResolvedAt time.Time `json:"resolved_at"`
}

type SchedulePlan struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	GateID        uint      `json:"gate_id"`
	Schedule      string    `gorm:"type:json" json:"schedule"`
	TotalWaitTime float64   `json:"total_wait_time"`
	Generation    int       `json:"generation"`
	Fitness       float64   `json:"fitness"`
}

type SimulationResult struct {
	FillTime           float64   `json:"fill_time"`
	DrainTime          float64   `json:"drain_time"`
	WaterLevelCurve    []WaterLevelPoint `json:"water_level_curve"`
	FlowRateCurve      []FlowRatePoint   `json:"flow_rate_curve"`
	MaxFlowRate        float64   `json:"max_flow_rate"`
	AvgFlowRate        float64   `json:"avg_flow_rate"`
	TotalWaterVolume   float64   `json:"total_water_volume"`
}

type WaterLevelPoint struct {
	Time       float64 `json:"time"`
	WaterLevel float64 `json:"water_level"`
}

type FlowRatePoint struct {
	Time     float64 `json:"time"`
	FlowRate float64 `json:"flow_rate"`
}

type ScheduleShip struct {
	ShipID     uint      `json:"ship_id"`
	ShipName   string    `json:"ship_name"`
	Priority   int       `json:"priority"`
	ArrivalTime time.Time `json:"arrival_time"`
	Direction  string    `json:"direction"`
}

type ScheduleItem struct {
	ShipID             uint      `json:"ship_id"`
	ShipName           string    `json:"ship_name"`
	StartTime          time.Time `json:"start_time"`
	EndTime            time.Time `json:"end_time"`
	WaitTime           float64   `json:"wait_time"`
	Priority           int       `json:"priority"`
	Direction          string    `json:"direction"`
	GateID             uint      `json:"gate_id,omitempty"`
	ShipType           string    `json:"ship_type,omitempty"`
	SegmentTravelTime  float64   `json:"segment_travel_time_s,omitempty"`
	FillTime           float64   `json:"fill_time_s,omitempty"`
	DrainTime          float64   `json:"drain_time_s,omitempty"`
}

type Dynasty string

const (
	DynastyTang   Dynasty = "tang"
	DynastySong   Dynasty = "song"
	DynastyQing   Dynasty = "qing"
	DynastyModern Dynasty = "modern"
)

type DynastyDesign struct {
	ID            uint    `gorm:"primaryKey" json:"id"`
	GateID        uint    `json:"gate_id"`
	Dynasty       Dynasty `json:"dynasty"`
	DynastyName   string  `json:"dynasty_name"`
	YearRange     string  `json:"year_range"`
	GateWidth     float64 `json:"gate_width"`
	GateHeight    float64 `json:"gate_height"`
	ChamberLength float64 `json:"chamber_length"`
	ChamberWidth  float64 `json:"chamber_width"`
	Material      string  `json:"material"`
	Structure     string  `json:"structure"`
	HoistType     string  `json:"hoist_type"`
	DefaultCd     float64 `json:"default_cd"`
	CcBase        float64 `json:"cc_base"`
	CcLinear      float64 `json:"cc_linear"`
	CcQuadratic   float64 `json:"cc_quadratic"`
	WeirCoeff     float64 `json:"weir_coeff"`
	MaxFlowRate   float64 `json:"design_max_flow_rate"`
	WaterLift     float64 `json:"design_water_lift"`
	HistoricalNote string `json:"historical_note"`
	Innovation     string `json:"innovation"`
	SourceRef      string `json:"source_ref"`
	ParamConfidence float64 `json:"param_confidence"`
}

type DynastyComparisonResult struct {
	Dynasty         Dynasty         `json:"dynasty"`
	DynastyName     string          `json:"dynasty_name"`
	Design          *DynastyDesign  `json:"design"`
	Simulation      *SimulationResultLite `json:"simulation"`
	EfficiencyScore float64         `json:"efficiency_score"`
	WaterPerTon     float64         `json:"water_per_ton_m3"`
	PassagesPerDay  int             `json:"passages_per_day"`
	Advantages      []string        `json:"advantages"`
	Limitations     []string        `json:"limitations"`
}

type SimulationResultLite struct {
	FillTime         float64 `json:"fill_time_s"`
	DrainTime        float64 `json:"drain_time_s"`
	MaxFlowRate      float64 `json:"max_flow_rate_m3s"`
	AvgFlowRate      float64 `json:"avg_flow_rate_m3s"`
	TotalWaterVolume float64 `json:"total_water_volume_m3"`
	Regime           string  `json:"flow_regime"`
}

type CanalSegment struct {
	ID            uint    `gorm:"primaryKey" json:"id"`
	SegmentCode   string  `json:"segment_code"`
	FromGateID    uint    `json:"from_gate_id"`
	ToGateID      uint    `json:"to_gate_id"`
	DistanceM     float64 `json:"distance_m"`
	TravelTimeS   float64 `json:"travel_time_s"`
	AvgCurrentMs  float64 `json:"avg_current_ms"`
	HasTributary  bool    `json:"has_tributary"`
	MaxShips      int     `json:"max_ships_concurrent"`
	SegmentOrder  int     `json:"segment_order"`
}

type MultiStageOptimizeRequest struct {
	CanalSegments   []CanalSegment     `json:"segments"`
	GateIDs         []uint             `json:"gate_ids"`
	Ships           []ScheduleShip     `json:"ships"`
	TravelSpeedFactor float64          `json:"travel_speed_factor,omitempty"`
	SimulateHydro   bool               `json:"simulate_hydro,omitempty"`
}

type MultiStageGateSchedule struct {
	GateID           uint            `json:"gate_id"`
	GateName         string          `json:"gate_name"`
	ArrivalTime      time.Time       `json:"arrival_time"`
	FillDrainStart   time.Time       `json:"fill_drain_start"`
	FillDrainEnd     time.Time       `json:"fill_drain_end"`
	EntryTime        time.Time       `json:"entry_time"`
	ExitTime         time.Time       `json:"exit_time"`
	DepartureTime    time.Time       `json:"departure_time"`
	FillTimeS        float64         `json:"fill_time_s"`
	DrainTimeS       float64         `json:"drain_time_s"`
	WaitTimeS        float64         `json:"wait_time_s"`
	WaterUsedM3      float64         `json:"water_used_m3"`
	Regime           string          `json:"flow_regime"`
}

type MultiStageShipRoute struct {
	ShipID           uint                      `json:"ship_id"`
	ShipName         string                    `json:"ship_name"`
	Direction        string                    `json:"direction"`
	Priority         int                       `json:"priority"`
	ShipType         string                    `json:"ship_type"`
	OriginGateID     uint                      `json:"origin_gate_id"`
	DestGateID       uint                      `json:"dest_gate_id"`
	GateSequence     []MultiStageGateSchedule  `json:"gate_sequence"`
	TotalWaitTimeS   float64                   `json:"total_wait_time_s"`
	TotalTravelTimeS float64                   `json:"total_travel_time_s"`
	TotalPassageTimeS float64                  `json:"total_passage_time_s"`
	TotalWaterUsedM3 float64                   `json:"total_water_used_m3"`
}

type MultiStageOptimizeResult struct {
	Routes           []MultiStageShipRoute `json:"routes"`
	TotalWaitTimeS   float64               `json:"total_wait_time_s"`
	TotalTravelTimeS float64               `json:"total_travel_time_s"`
	TotalWaterUsedM3 float64               `json:"total_water_used_m3"`
	ThroughputShips  int                   `json:"throughput_ships"`
	ThroughputPerDay float64               `json:"throughput_per_day"`
	GateUtilization  map[uint]float64      `json:"gate_utilization"`
	ConflictCount    int                   `json:"conflict_count"`
	Fitness          float64               `json:"fitness"`
	Generations      int                   `json:"generations"`
	Error            string                `json:"error,omitempty"`
}

type ShipType string

const (
	ShipTypeGrain  ShipType = "grain"
	ShipTypeCargo  ShipType = "cargo"
	ShipTypePassenger ShipType = "passenger"
	ShipTypeMilitary ShipType = "military"
	ShipTypeFishing ShipType = "fishing"
	ShipTypeTribute ShipType = "tribute"
	ShipTypeRoyal   ShipType = "royal"
)

type ShipTypeSpec struct {
	ShipType        ShipType `json:"ship_type"`
	TypeName        string   `json:"type_name"`
	LengthMin       float64  `json:"length_min"`
	LengthMax       float64  `json:"length_max"`
	WidthMin        float64  `json:"width_min"`
	WidthMax        float64  `json:"width_max"`
	DraftMin        float64  `json:"draft_min"`
	DraftMax        float64  `json:"draft_max"`
	CapacityTon     float64  `json:"capacity_ton"`
	BasePriority    int      `json:"base_priority"`
	EntryTimeS      float64  `json:"entry_time_s"`
	ExitTimeS       float64  `json:"exit_time_s"`
	WaterFactor     float64  `json:"water_factor"`
	HistoricalUsage string   `json:"historical_usage"`
	ColorHex        string   `json:"color_hex"`
	FormFactor      float64  `json:"form_factor"`
	ResistanceCoeff float64  `json:"resistance_coeff"`
	ChamberOccupancy float64 `json:"chamber_occupancy"`
}

type ShipTypeEfficiencyReport struct {
	ShipType         ShipType  `json:"ship_type"`
	TypeName         string    `json:"type_name"`
	AvgWaitTimeS     float64   `json:"avg_wait_time_s"`
	AvgPassageTimeS  float64   `json:"avg_passage_time_s"`
	AvgWaterPerTon   float64   `json:"avg_water_per_ton_m3"`
	AvgThroughputTon float64   `json:"avg_throughput_ton_per_day"`
	ConflictRate     float64   `json:"conflict_rate"`
	SampleCount      int       `json:"sample_count"`
	EfficiencyIndex  float64   `json:"efficiency_index_0_100"`
}

type ShipTypeAnalysisRequest struct {
	GateID        uint         `json:"gate_id"`
	WaterLevelUp  float64      `json:"water_level_up"`
	WaterLevelDown float64     `json:"water_level_down"`
	GateOpening   float64      `json:"gate_opening"`
	ShipTypes     []ShipType   `json:"ship_types,omitempty"`
}

type DynastyComparisonRequest struct {
	GateID         uint    `json:"gate_id"`
	WaterLevelUp   float64 `json:"water_level_up"`
	WaterLevelDown float64 `json:"water_level_down"`
	GateOpening    float64 `json:"gate_opening,omitempty"`
	Direction      string  `json:"direction,omitempty"`
	Dynasties      []Dynasty `json:"dynasties,omitempty"`
}

func GetDynastyDesigns(gateID uint) []*DynastyDesign {
	return []*DynastyDesign{
		{
			GateID: gateID, Dynasty: DynastyTang, DynastyName: "唐代陡门", YearRange: "610-907",
			GateWidth: 4.0, GateHeight: 2.5, ChamberLength: 25, ChamberWidth: 4.8,
			Material: "土石木混合", Structure: "单门叠梁式", HoistType: "人力绞车",
			DefaultCd: 0.48, CcBase: 0.56, CcLinear: 0.08, CcQuadratic: -0.015,
			WeirCoeff: 1.35, MaxFlowRate: 12.8, WaterLift: 1.5,
			HistoricalNote: "唐宝历初年(825年)观察使李渤创设陡门，以石木叠梁阻水，人力绞关启闭，为灵渠通航之雏形。唐陡门窄小(约4m)，叠梁间隙大，泄漏严重。",
			Innovation:     "世界最早之分级闸道理念，单陡可提升水位1.2-1.8米",
			SourceRef:      "《新唐书·地理志》; 郑连第《灵渠工程史述略》(1990); 广西文物考古报告",
			ParamConfidence: 0.55,
		},
		{
			GateID: gateID, Dynasty: DynastySong, DynastyName: "宋代陡门", YearRange: "960-1279",
			GateWidth: 4.8, GateHeight: 3.2, ChamberLength: 38, ChamberWidth: 5.8,
			Material: "包铁条石", Structure: "对开立式门+双门闸室", HoistType: "畜力滑车",
			DefaultCd: 0.55, CcBase: 0.59, CcLinear: 0.095, CcQuadratic: -0.018,
			WeirCoeff: 1.55, MaxFlowRate: 24.2, WaterLift: 2.2,
			HistoricalNote: "宋嘉祐三年(1058年)提点刑狱李师中兴修三十六陡，改为条石包铁，形成完整梯级通航体系。宋陡门宽约4.8-5.4m，闸室长约38-45m。",
			Innovation:     "首创双门闭合闸室，畜力提升启闭效率3倍，日通行量达30艘",
			SourceRef:      "《宋史·河渠志》; 李师中《修灵渠记》; 灵渠考古实测数据(1986)",
			ParamConfidence: 0.65,
		},
		{
			GateID: gateID, Dynasty: DynastyQing, DynastyName: "清代陡门", YearRange: "1644-1911",
			GateWidth: 5.4, GateHeight: 3.8, ChamberLength: 48, ChamberWidth: 6.4,
			Material: "糯米灰浆砌青石", Structure: "重门式+泄水副槽", HoistType: "双绞盘人牛并用",
			DefaultCd: 0.59, CcBase: 0.61, CcLinear: 0.10, CcQuadratic: -0.02,
			WeirCoeff: 1.72, MaxFlowRate: 36.5, WaterLift: 2.8,
			HistoricalNote: "清康熙二十二年(1683年)大规模重修，采用糯米灰浆砌青石工艺，增设泄水副槽防冲蚀。清陡实测宽约5.4-6.0m，闸室长48-56m，提升2.5-3.0m。",
			Innovation:     "闸室扩容+副槽减涡设计，单级最大提升2.8米，泄水效率较宋提升50%",
			SourceRef:      "《兴安县志·乾隆版》; 陈宏谋《修灵渠碑记》; 1939年扬子江水利委员会实测图",
			ParamConfidence: 0.75,
		},
		{
			GateID: gateID, Dynasty: DynastyModern, DynastyName: "现代修复陡门", YearRange: "1949-至今",
			GateWidth: 6.0, GateHeight: 4.6, ChamberLength: 60, ChamberWidth: 7.0,
			Material: "钢筋混凝土+古貌贴面", Structure: "仿古重门+液压启闭", HoistType: "液压卷扬机",
			DefaultCd: 0.63, CcBase: 0.615, CcLinear: 0.105, CcQuadratic: -0.02,
			WeirCoeff: 1.84, MaxFlowRate: 48.0, WaterLift: 3.5,
			HistoricalNote: "1985-1990年全面修复，保留古陡外观，内部采用钢筋混凝土结构和现代液压启闭系统。参数基于实测数据。",
			Innovation:     "外观古法+结构现代，启闭时间由25分钟缩短至8分钟",
			SourceRef:      "《灵渠志》(2005)GB版; 桂林水利设计院实测报告(1989); 现场测绘数据",
			ParamConfidence: 0.92,
		},
	}
}

func GetShipTypeSpecs() []*ShipTypeSpec {
	return []*ShipTypeSpec{
		{
			ShipType: ShipTypeGrain, TypeName: "漕船(粮船)",
			LengthMin: 22, LengthMax: 30, WidthMin: 4.0, WidthMax: 5.2, DraftMin: 1.2, DraftMax: 1.8,
			CapacityTon: 80, BasePriority: 3, EntryTimeS: 180, ExitTimeS: 150, WaterFactor: 1.0,
			HistoricalUsage: "漕运官粮，每年春秋两季集中运输，占灵渠货运量60%", ColorHex: "#d4a574",
			FormFactor: 0.82, ResistanceCoeff: 0.012, ChamberOccupancy: 0.38,
		},
		{
			ShipType: ShipTypeCargo, TypeName: "货船(杂货)",
			LengthMin: 16, LengthMax: 24, WidthMin: 3.2, WidthMax: 4.2, DraftMin: 0.9, DraftMax: 1.4,
			CapacityTon: 40, BasePriority: 2, EntryTimeS: 120, ExitTimeS: 100, WaterFactor: 0.85,
			HistoricalUsage: "盐铁陶瓷布匹杂货，商民常用船型，常年通行", ColorHex: "#8b7355",
			FormFactor: 0.78, ResistanceCoeff: 0.010, ChamberOccupancy: 0.24,
		},
		{
			ShipType: ShipTypePassenger, TypeName: "客船(画舫)",
			LengthMin: 14, LengthMax: 20, WidthMin: 3.0, WidthMax: 3.8, DraftMin: 0.6, DraftMax: 1.0,
			CapacityTon: 15, BasePriority: 2, EntryTimeS: 90, ExitTimeS: 80, WaterFactor: 0.65,
			HistoricalUsage: "官绅商旅往来乘坐，装饰华丽，吃水较浅", ColorHex: "#c9a86c",
			FormFactor: 0.70, ResistanceCoeff: 0.008, ChamberOccupancy: 0.17,
		},
		{
			ShipType: ShipTypeTribute, TypeName: "贡船(贡品)",
			LengthMin: 26, LengthMax: 34, WidthMin: 4.8, WidthMax: 5.8, DraftMin: 1.4, DraftMax: 2.0,
			CapacityTon: 120, BasePriority: 5, EntryTimeS: 240, ExitTimeS: 200, WaterFactor: 1.15,
			HistoricalUsage: "运送岭南贡品北上入京(珊瑚/香料/药材/铜锭)，皇家特优先权", ColorHex: "#b8860b",
			FormFactor: 0.88, ResistanceCoeff: 0.015, ChamberOccupancy: 0.47,
		},
		{
			ShipType: ShipTypeMilitary, TypeName: "军船(漕运兼用)",
			LengthMin: 24, LengthMax: 32, WidthMin: 4.5, WidthMax: 5.5, DraftMin: 1.1, DraftMax: 1.7,
			CapacityTon: 70, BasePriority: 4, EntryTimeS: 180, ExitTimeS: 150, WaterFactor: 0.95,
			HistoricalUsage: "驻军换防/军粮北运，与漕船形制相近但速度较快", ColorHex: "#5a6e7f",
			FormFactor: 0.80, ResistanceCoeff: 0.011, ChamberOccupancy: 0.35,
		},
		{
			ShipType: ShipTypeFishing, TypeName: "渔船(小型)",
			LengthMin: 5, LengthMax: 10, WidthMin: 1.5, WidthMax: 2.4, DraftMin: 0.3, DraftMax: 0.6,
			CapacityTon: 3, BasePriority: 1, EntryTimeS: 40, ExitTimeS: 30, WaterFactor: 0.2,
			HistoricalUsage: "沿岸渔民打鱼谋生，船小灵活，随来随过不占闸室", ColorHex: "#6b8e6b",
			FormFactor: 0.60, ResistanceCoeff: 0.006, ChamberOccupancy: 0.05,
		},
		{
			ShipType: ShipTypeRoyal, TypeName: "御舟(皇家专用)",
			LengthMin: 42, LengthMax: 55, WidthMin: 7.0, WidthMax: 8.5, DraftMin: 1.6, DraftMax: 2.3,
			CapacityTon: 200, BasePriority: 6, EntryTimeS: 360, ExitTimeS: 300, WaterFactor: 1.4,
			HistoricalUsage: "皇帝/钦差南巡专用，规模宏大，需临时加固陡门", ColorHex: "#8b0000",
			FormFactor: 0.95, ResistanceCoeff: 0.018, ChamberOccupancy: 0.68,
		},
	}
}

func GetDefaultCanalSegments() []*CanalSegment {
	segs := []*CanalSegment{}
	order := 1
	for i := 1; i <= 35; i++ {
		var dist float64
		switch {
		case i <= 12:
			dist = 420 + float64(i%5)*80
		case i <= 24:
			dist = 380 + float64(i%4)*60
		default:
			dist = 350 + float64(i%6)*70
		}
		speed := 1.2
		if i >= 10 && i <= 18 {
			speed = 1.4
		}
		segs = append(segs, &CanalSegment{
			SegmentCode:  "L" + itoaSeg(order),
			FromGateID:   uint(i),
			ToGateID:     uint(i + 1),
			DistanceM:    dist,
			TravelTimeS:  dist / speed,
			AvgCurrentMs: 0.3,
			HasTributary: i == 7 || i == 15 || i == 27,
			MaxShips:     map[int]int{7: 2, 15: 3, 27: 2}[i],
			SegmentOrder: order,
		})
		if map[int]int{7: 2, 15: 3, 27: 2}[i] == 0 {
			segs[order-1].MaxShips = 4
		}
		order++
	}
	return segs
}

func itoaSeg(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}
