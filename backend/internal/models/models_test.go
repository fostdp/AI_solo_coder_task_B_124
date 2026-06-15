package models

import (
	"math"
	"testing"
)

func TestGetDynastyDesigns_Count(t *testing.T) {
	designs := GetDynastyDesigns(1)
	if len(designs) != 4 {
		t.Errorf("expected 4 dynasty designs, got %d", len(designs))
	}
}

func TestGetDynastyDesigns_Dynasties(t *testing.T) {
	designs := GetDynastyDesigns(1)
	expected := []Dynasty{DynastyTang, DynastySong, DynastyQing, DynastyModern}
	for i, d := range designs {
		if d.Dynasty != expected[i] {
			t.Errorf("design[%d]: expected %s, got %s", i, expected[i], d.Dynasty)
		}
	}
}

func TestGetDynastyDesigns_GateIDPropagation(t *testing.T) {
	for gateID := uint(1); gateID <= 5; gateID++ {
		designs := GetDynastyDesigns(gateID)
		for _, d := range designs {
			if d.GateID != gateID {
				t.Errorf("gateID=%d: design.GateID=%d mismatch", gateID, d.GateID)
			}
		}
	}
}

func TestGetDynastyDesigns_PhysicalConsistency(t *testing.T) {
	designs := GetDynastyDesigns(1)
	for _, d := range designs {
		if d.GateWidth <= 0 {
			t.Errorf("dynasty %s: gate width should be positive", d.Dynasty)
		}
		if d.GateHeight <= 0 {
			t.Errorf("dynasty %s: gate height should be positive", d.Dynasty)
		}
		if d.ChamberLength <= 0 {
			t.Errorf("dynasty %s: chamber length should be positive", d.Dynasty)
		}
		if d.DefaultCd <= 0 || d.DefaultCd > 1 {
			t.Errorf("dynasty %s: discharge coefficient %f should be in (0,1]", d.Dynasty, d.DefaultCd)
		}
		if d.WaterLift <= 0 {
			t.Errorf("dynasty %s: water lift should be positive", d.Dynasty)
		}
		if d.MaxFlowRate <= 0 {
			t.Errorf("dynasty %s: max flow rate should be positive", d.Dynasty)
		}
		if d.WeirCoeff <= 0 {
			t.Errorf("dynasty %s: weir coefficient should be positive", d.Dynasty)
		}
	}
}

func TestGetDynastyDesigns_TechnologicalProgression(t *testing.T) {
	designs := GetDynastyDesigns(1)
	for i := 1; i < len(designs); i++ {
		prev := designs[i-1]
		curr := designs[i]
		if curr.ChamberLength < prev.ChamberLength*0.9 {
			t.Errorf("chamber length should generally increase: %s(%f) -> %s(%f)",
				prev.Dynasty, prev.ChamberLength, curr.Dynasty, curr.ChamberLength)
		}
	}
}

func TestGetShipTypeSpecs_Count(t *testing.T) {
	specs := GetShipTypeSpecs()
	if len(specs) != 7 {
		t.Errorf("expected 7 ship type specs, got %d", len(specs))
	}
}

func TestGetShipTypeSpecs_AllTypes(t *testing.T) {
	specs := GetShipTypeSpecs()
	expected := []ShipType{ShipTypeGrain, ShipTypeCargo, ShipTypePassenger, ShipTypeTribute,
		ShipTypeMilitary, ShipTypeFishing, ShipTypeRoyal}
	found := map[ShipType]bool{}
	for _, s := range specs {
		found[s.ShipType] = true
	}
	for _, e := range expected {
		if !found[e] {
			t.Errorf("missing ship type: %s", e)
		}
	}
}

func TestGetShipTypeSpecs_PhysicalConsistency(t *testing.T) {
	specs := GetShipTypeSpecs()
	for _, s := range specs {
		if s.LengthMin <= 0 || s.LengthMax <= 0 {
			t.Errorf("ship %s: length should be positive", s.ShipType)
		}
		if s.LengthMin > s.LengthMax {
			t.Errorf("ship %s: length min (%f) > max (%f)", s.ShipType, s.LengthMin, s.LengthMax)
		}
		if s.WidthMin > s.WidthMax {
			t.Errorf("ship %s: width min (%f) > max (%f)", s.ShipType, s.WidthMin, s.WidthMax)
		}
		if s.DraftMin > s.DraftMax {
			t.Errorf("ship %s: draft min (%f) > max (%f)", s.ShipType, s.DraftMin, s.DraftMax)
		}
		if s.CapacityTon <= 0 {
			t.Errorf("ship %s: capacity should be positive", s.ShipType)
		}
		if s.EntryTimeS < 0 || s.ExitTimeS < 0 {
			t.Errorf("ship %s: entry/exit time should be non-negative", s.ShipType)
		}
		if s.WaterFactor <= 0 {
			t.Errorf("ship %s: water factor should be positive", s.ShipType)
		}
		if s.BasePriority < 1 || s.BasePriority > 6 {
			t.Errorf("ship %s: base priority %d should be in [1,6]", s.ShipType, s.BasePriority)
		}
	}
}

func TestGetShipTypeSpecs_DraftWidthRatio(t *testing.T) {
	specs := GetShipTypeSpecs()
	for _, s := range specs {
		ratio := s.DraftMax / s.WidthMax
		if ratio > 1.0 {
			t.Errorf("ship %s: draft/width ratio (%f) seems too high", s.ShipType, ratio)
		}
	}
}

func TestGetDefaultCanalSegments_Count(t *testing.T) {
	segs := GetDefaultCanalSegments()
	if len(segs) != 35 {
		t.Errorf("expected 35 canal segments, got %d", len(segs))
	}
}

func TestGetDefaultCanalSegments_Ordering(t *testing.T) {
	segs := GetDefaultCanalSegments()
	for i, s := range segs {
		if s.SegmentOrder != i+1 {
			t.Errorf("segment[%d]: expected order %d, got %d", i, i+1, s.SegmentOrder)
		}
	}
}

func TestGetDefaultCanalSegments_Continuity(t *testing.T) {
	segs := GetDefaultCanalSegments()
	for i := 1; i < len(segs); i++ {
		if segs[i].FromGateID != segs[i-1].ToGateID {
			t.Errorf("segment[%d]: from_gate_id=%d != prev to_gate_id=%d, break in continuity",
				i, segs[i].FromGateID, segs[i-1].ToGateID)
		}
	}
}

func TestGetDefaultCanalSegments_PhysicalConsistency(t *testing.T) {
	segs := GetDefaultCanalSegments()
	for i, s := range segs {
		if s.DistanceM <= 0 {
			t.Errorf("segment[%d]: distance should be positive", i)
		}
		if s.TravelTimeS <= 0 {
			t.Errorf("segment[%d]: travel time should be positive", i)
		}
		if s.MaxShips <= 0 {
			t.Errorf("segment[%d]: max ships should be positive", i)
		}
		if math.IsInf(s.DistanceM, 0) || math.IsNaN(s.DistanceM) {
			t.Errorf("segment[%d]: distance is Inf/NaN", i)
		}
	}
}

func TestDynastyString(t *testing.T) {
	tests := []struct {
		d Dynasty
		s string
	}{
		{DynastyTang, "tang"},
		{DynastySong, "song"},
		{DynastyQing, "qing"},
		{DynastyModern, "modern"},
	}
	for _, tt := range tests {
		if string(tt.d) != tt.s {
			t.Errorf("expected %s, got %s", tt.s, string(tt.d))
		}
	}
}

func TestShipTypeString(t *testing.T) {
	tests := []struct {
		s ShipType
		v string
	}{
		{ShipTypeGrain, "grain"},
		{ShipTypeCargo, "cargo"},
		{ShipTypePassenger, "passenger"},
		{ShipTypeMilitary, "military"},
		{ShipTypeFishing, "fishing"},
		{ShipTypeTribute, "tribute"},
		{ShipTypeRoyal, "royal"},
	}
	for _, tt := range tests {
		if string(tt.s) != tt.v {
			t.Errorf("expected %s, got %s", tt.v, string(tt.s))
		}
	}
}
