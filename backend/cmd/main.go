package main

import (
	"expvar"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"lingqu-dou-gate/internal/config"
	"lingqu-dou-gate/internal/handlers"
	"lingqu-dou-gate/internal/middleware"
	"lingqu-dou-gate/internal/modules/alarm_mqtt"
	"lingqu-dou-gate/internal/modules/cascade_scheduler"
	"lingqu-dou-gate/internal/modules/design_comparator"
	"lingqu-dou-gate/internal/modules/dtu_receiver"
	"lingqu-dou-gate/internal/modules/hydraulic_sim"
	"lingqu-dou-gate/internal/modules/scheduler_ga"
	"lingqu-dou-gate/internal/modules/vessel_analyzer"
	"lingqu-dou-gate/internal/modules/vr_lock_experience"
	"lingqu-dou-gate/internal/services"
)

var (
	buildTime = "unknown"
	gitHash   = "unknown"
	version   = "dev"
)

func main() {
	config.Load()

	log.Printf("Starting DouGate Scheduler | version=%s | git=%s | built=%s", version, gitHash, buildTime)

	services.InitDB()
	defer services.CloseDB()

	services.InitMQTT()
	defer services.CloseMQTT()

	dtuReceiver := dtu_receiver.NewDTUReceiver(2)
	hydraulicSim := hydraulic_sim.NewHydraulicSimulator(2)
	schedulerGA := scheduler_ga.NewGAScheduler(2)
	alarmMqtt := alarm_mqtt.NewAlarmMqtt(dtuReceiver.ValidatedDataChannel(), 2)

	designComparator := design_comparator.NewDesignComparator(hydraulicSim)
	cascadeScheduler := cascade_scheduler.NewCascadeScheduler()
	vesselAnalyzer := vessel_analyzer.NewVesselAnalyzer(hydraulicSim)
	vrLockExperience := vr_lock_experience.NewVRLockExperience(hydraulicSim)

	metrics := middleware.GetMetricsCollector()

	dtuReceiver.Start()
	hydraulicSim.Start()
	schedulerGA.Start()
	alarmMqtt.Start()

	defer func() {
		dtuReceiver.Stop()
		hydraulicSim.Stop()
		schedulerGA.Stop()
		alarmMqtt.Stop()
	}()

	handler := handlers.NewHandler(
		dtuReceiver,
		hydraulicSim,
		schedulerGA,
		alarmMqtt,
		metrics,
		designComparator,
		cascadeScheduler,
		vesselAnalyzer,
		vrLockExperience,
	)

	// ======== pprof + Prometheus + expvar 管理端点（6060端口）========
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/debug/vars", expvar.Handler().ServeHTTP)
		mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(200)
			w.Write([]byte(metrics.ExportPrometheus()))
		})
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"ok","version":"` + version + `","build":"` + buildTime + `"}`))
		})
		mux.HandleFunc("/build", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"version":"` + version + `","git_hash":"` + gitHash + `","build_time":"` + buildTime + `"}`))
		})
		adminAddr := "0.0.0.0:6060"
		log.Printf("Admin server [pprof/metrics/expvar] starting on %s", adminAddr)
		if err := http.ListenAndServe(adminAddr, mux); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Admin server error: %v", err)
		}
	}()

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// ======== Prometheus HTTP指标中间件 ========
	r.Use(gin.WrapH(metrics.HTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 作为gin的中间件栈嵌入；gin.WrapH会把它转换为gin handler
	}))))

	api := r.Group("/api")
	{
		api.GET("/gates", handler.GetGates)
		api.GET("/gates/:id", handler.GetGate)
		api.GET("/dynasties/:gateId", handler.GetDynastyDesigns)
		api.POST("/gates/:id/dynasty-comparison", handler.DynastyComparison)

		api.GET("/sensors/:gateId", handler.GetSensorData)
		api.GET("/sensors/:gateId/history", handler.GetSensorHistory)
		api.POST("/sensors", handler.PostSensorData)

		api.POST("/simulate", handler.SimulatePassage)
		api.GET("/simulation/:gateId", handler.GetSimulationData)

		api.POST("/optimize", handler.OptimizeSchedule)
		api.POST("/optimize/multi-stage", handler.OptimizeMultiStage)
		api.GET("/canal/segments", handler.GetCanalSegments)

		api.GET("/ship-types", handler.GetShipTypes)
		api.POST("/analysis/ship-type-efficiency", handler.ShipTypeEfficiencyAnalysis)

		api.GET("/alerts", handler.GetAlerts)
		api.POST("/alerts/:id/resolve", handler.ResolveAlert)
		api.POST("/alerts/test", handler.TestAlert)

		api.GET("/vr/scenarios", handler.ListVRScenarios)
		api.GET("/vr/scenarios/:id", handler.GetVRScenario)
		api.POST("/vr/sessions", handler.NewVRSession)
		api.GET("/vr/scenarios/:id/simulate", handler.SimulateVRScenario)
	}

	addr := config.AppConfig.Server.Host + ":" + config.AppConfig.Server.Port
	log.Printf("API server starting on %s", addr)
	log.Printf("Modules: DTU=running, HydraulicSim=running, SchedulerGA=running, AlarmMQTT=running, DesignComparator=ok, CascadeScheduler=ok, VesselAnalyzer=ok, VRLockExperience=ok")
	log.Printf("Endpoints: /debug/pprof /metrics /healthz on :6060")

	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatalf("Failed to start API server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gracefully...")
}
