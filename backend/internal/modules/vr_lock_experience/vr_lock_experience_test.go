package vr_lock_experience

import (
	"testing"

	"lingqu-dou-gate/internal/config"
	"lingqu-dou-gate/internal/models"
	"lingqu-dou-gate/internal/modules/hydraulic_sim"
)

func makeTestExperience() *VRLockExperience {
	config.AppConfig.HydraulicJSON = config.DefaultHydraulicJSONConfig()
	hydro := hydraulic_sim.NewHydraulicSimulator(1)
	return NewVRLockExperience(hydro)
}

func TestNewVRLockExperience(t *testing.T) {
	ve := makeTestExperience()
	if ve == nil {
		t.Fatal("VRLockExperience should not be nil")
	}
	if len(ve.scenarios) != 4 {
		t.Errorf("expected 4 scenarios, got %d", len(ve.scenarios))
	}
}

func TestListScenarios(t *testing.T) {
	ve := makeTestExperience()
	scenarios := ve.ListScenarios()

	if len(scenarios) != 4 {
		t.Errorf("expected 4 scenarios, got %d", len(scenarios))
	}
	ids := map[string]bool{}
	for _, s := range scenarios {
		ids[s.ScenarioID] = true
		if s.ScenarioName == "" {
			t.Errorf("scenario %s: name should not be empty", s.ScenarioID)
		}
		if s.Description == "" {
			t.Errorf("scenario %s: description should not be empty", s.ScenarioID)
		}
		if s.WaterLevelUp <= s.WaterLevelDown {
			t.Errorf("scenario %s: waterLevelUp %f should be > waterLevelDown %f",
				s.ScenarioID, s.WaterLevelUp, s.WaterLevelDown)
		}
	}
	expected := []string{"tang_upstream", "song_downstream", "qing_royal", "modern_fishing"}
	for _, id := range expected {
		if !ids[id] {
			t.Errorf("missing expected scenario: %s", id)
		}
	}
}

func TestGetScenario_Exists(t *testing.T) {
	ve := makeTestExperience()
	scenario, ok := ve.GetScenario("tang_upstream")

	if !ok {
		t.Fatal("should find tang_upstream scenario")
	}
	if scenario.Dynasty != models.DynastyTang {
		t.Errorf("expected tang dynasty, got %s", scenario.Dynasty)
	}
	if scenario.Direction != "upstream" {
		t.Errorf("expected upstream, got %s", scenario.Direction)
	}
}

func TestGetScenario_NotExists(t *testing.T) {
	ve := makeTestExperience()
	_, ok := ve.GetScenario("nonexistent")

	if ok {
		t.Error("should not find nonexistent scenario")
	}
}

func TestNewSession(t *testing.T) {
	ve := makeTestExperience()
	sess, err := ve.NewSession("tang_upstream", "user_123")

	if err != nil {
		t.Fatalf("new session should not error: %v", err)
	}
	if sess == nil {
		t.Fatal("session should not be nil")
	}
	if sess.SessionID == "" {
		t.Error("session ID should not be empty")
	}
	if sess.UserID != "user_123" {
		t.Errorf("expected user_123, got %s", sess.UserID)
	}
	if sess.State != "idle" {
		t.Errorf("expected idle state, got %s", sess.State)
	}
	if sess.ScoreTotal != 100 {
		t.Errorf("expected score 100, got %d", sess.ScoreTotal)
	}
	if len(sess.Errors) != 0 {
		t.Errorf("expected empty errors, got %d", len(sess.Errors))
	}
}

func TestNewSession_DefaultScenario(t *testing.T) {
	ve := makeTestExperience()
	sess, _ := ve.NewSession("nonexistent", "")

	if sess == nil {
		t.Fatal("should return default scenario session")
	}
}

func TestGetSession_Exists(t *testing.T) {
	ve := makeTestExperience()
	sess, _ := ve.NewSession("tang_upstream", "u1")
	found, ok := ve.GetSession(sess.SessionID)

	if !ok {
		t.Fatal("should find session by ID")
	}
	if found.SessionID != sess.SessionID {
		t.Errorf("session ID mismatch: %s vs %s", found.SessionID, sess.SessionID)
	}
}

func TestGetSession_NotExists(t *testing.T) {
	ve := makeTestExperience()
	_, ok := ve.GetSession("nonexistent_id")

	if ok {
		t.Error("should not find nonexistent session")
	}
}

func TestSimulateScenario_Valid(t *testing.T) {
	ve := makeTestExperience()
	result, err := ve.SimulateScenario("tang_upstream")

	if err != nil {
		t.Fatalf("simulate should not error: %v", err)
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.TotalPassageS <= 0 {
		t.Errorf("total passage time should be positive, got %f", result.TotalPassageS)
	}
	if result.TotalWaterVolume <= 0 {
		t.Errorf("total water volume should be positive, got %f", result.TotalWaterVolume)
	}
}

func TestSimulateScenario_NotFound(t *testing.T) {
	ve := makeTestExperience()
	result, err := ve.SimulateScenario("nonexistent")

	if err != nil {
		t.Fatalf("should not error for missing scenario, returns default: %v", err)
	}
	if result == nil {
		t.Fatal("result should not be nil (falls back to default)")
	}
}

func TestScenarios_CoverAllDynasties(t *testing.T) {
	ve := makeTestExperience()
	scenarios := ve.ListScenarios()

	dynastySet := map[models.Dynasty]bool{}
	for _, s := range scenarios {
		dynastySet[s.Dynasty] = true
	}
	expected := []models.Dynasty{
		models.DynastyTang, models.DynastySong,
		models.DynastyQing, models.DynastyModern,
	}
	for _, d := range expected {
		if !dynastySet[d] {
			t.Errorf("missing dynasty scenario: %s", d)
		}
	}
}

func TestScenarios_Bidirectional(t *testing.T) {
	ve := makeTestExperience()
	scenarios := ve.ListScenarios()

	hasUp := false
	hasDown := false
	for _, s := range scenarios {
		if s.Direction == "upstream" {
			hasUp = true
		}
		if s.Direction == "downstream" {
			hasDown = true
		}
	}
	if !hasUp {
		t.Error("should have at least one upstream scenario")
	}
	if !hasDown {
		t.Error("should have at least one downstream scenario")
	}
}

func TestScenarios_MultipleShipTypes(t *testing.T) {
	ve := makeTestExperience()
	scenarios := ve.ListScenarios()

	shipSet := map[models.ShipType]bool{}
	for _, s := range scenarios {
		shipSet[s.ShipType] = true
	}
	if len(shipSet) < 3 {
		t.Errorf("should have at least 3 distinct ship types, got %d", len(shipSet))
	}
}
