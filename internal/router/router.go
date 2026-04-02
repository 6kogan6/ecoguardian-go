package router

import (
	"net/http"
	"time"

	"ecoguardian-go/internal/model"
	"ecoguardian-go/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, simulator *service.SimulatorService) *gin.Engine {
	r := gin.Default()

	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "./web/static")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "EcoGuardian Dashboard",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	api := r.Group("/api")
	{
		api.GET("/info", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"service": "EcoGuardian",
				"status":  "running",
				"version": "0.2.0",
			})
		})

		api.GET("/dashboard/summary", func(c *gin.Context) {
			var sensorsCount int64
			var readingsCount int64
			var activeAlertsCount int64

			if err := db.Model(&model.Sensor{}).Count(&sensorsCount).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count sensors"})
				return
			}

			if err := db.Model(&model.Reading{}).Count(&readingsCount).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count readings"})
				return
			}

			if err := db.Model(&model.Alert{}).Where("status = ?", "active").Count(&activeAlertsCount).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count active alerts"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"sensors_count":       sensorsCount,
				"readings_count":      readingsCount,
				"active_alerts_count": activeAlertsCount,
			})
		})

		api.GET("/sensors", func(c *gin.Context) {
			var sensors []model.Sensor

			if err := db.Order("id desc").Find(&sensors).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch sensors"})
				return
			}

			c.JSON(http.StatusOK, sensors)
		})

		api.GET("/sensors/:id", func(c *gin.Context) {
			var sensor model.Sensor

			if err := db.First(&sensor, c.Param("id")).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "sensor not found"})
				return
			}

			c.JSON(http.StatusOK, sensor)
		})

		api.POST("/sensors", func(c *gin.Context) {
			var sensor model.Sensor

			if err := c.ShouldBindJSON(&sensor); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if sensor.Name == "" || sensor.Location == "" || sensor.SensorType == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "name, location and sensor_type are required"})
				return
			}

			if sensor.CreatedAt.IsZero() {
				sensor.CreatedAt = time.Now()
			}

			if err := db.Create(&sensor).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create sensor"})
				return
			}

			c.JSON(http.StatusCreated, sensor)
		})

		api.GET("/readings", func(c *gin.Context) {
			var readings []model.Reading

			if err := db.Preload("Sensor").Order("recorded_at desc").Limit(100).Find(&readings).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch readings"})
				return
			}

			c.JSON(http.StatusOK, readings)
		})

		api.GET("/readings/latest", func(c *gin.Context) {
			var readings []model.Reading

			if err := db.Preload("Sensor").Order("recorded_at desc").Limit(10).Find(&readings).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch latest readings"})
				return
			}

			c.JSON(http.StatusOK, readings)
		})

		api.POST("/readings", func(c *gin.Context) {
			var reading model.Reading

			if err := c.ShouldBindJSON(&reading); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if reading.SensorID == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "sensor_id is required"})
				return
			}

			var sensor model.Sensor
			if err := db.First(&sensor, reading.SensorID).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "sensor not found"})
				return
			}

			if reading.RecordedAt.IsZero() {
				reading.RecordedAt = time.Now()
			}

			if err := db.Create(&reading).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create reading"})
				return
			}

			alerts := service.BuildAlerts(reading)
			if len(alerts) > 0 {
				if err := db.Create(&alerts).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "reading saved, but failed to create alerts"})
					return
				}
			}

			var createdReading model.Reading
			if err := db.Preload("Sensor").First(&createdReading, reading.ID).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "reading created, but failed to load reading"})
				return
			}

			c.JSON(http.StatusCreated, gin.H{
				"reading":        createdReading,
				"created_alerts": alerts,
			})
		})

		api.GET("/alerts", func(c *gin.Context) {
			var alerts []model.Alert

			if err := db.Preload("Sensor").Order("created_at desc").Limit(100).Find(&alerts).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch alerts"})
				return
			}

			c.JSON(http.StatusOK, alerts)
		})

		api.POST("/alerts/:id/resolve", func(c *gin.Context) {
			var alert model.Alert

			if err := db.First(&alert, c.Param("id")).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
				return
			}

			alert.Status = "resolved"

			if err := db.Save(&alert).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update alert"})
				return
			}

			if err := db.Preload("Sensor").First(&alert, alert.ID).Error; err != nil {
				c.JSON(http.StatusOK, alert)
				return
			}

			c.JSON(http.StatusOK, alert)
		})

		api.POST("/integrations/export", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "Data prepared for export to external environmental monitoring system",
				"sent_at": time.Now(),
			})
		})

		api.GET("/simulator/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, simulator.Status())
		})

		api.POST("/simulator/start", func(c *gin.Context) {
			if err := simulator.Start(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start simulator"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "Simulator started",
				"status":  simulator.Status(),
			})
		})

		api.POST("/simulator/stop", func(c *gin.Context) {
			if err := simulator.Stop(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to stop simulator"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "Simulator stopped",
				"status":  simulator.Status(),
			})
		})

		api.POST("/simulator/generate-once", func(c *gin.Context) {
			count, alerts, err := simulator.GenerateOnce()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate data"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message":          "Simulation cycle completed",
				"generated":        count,
				"generated_alerts": alerts,
				"status":           simulator.Status(),
			})
		})
	}

	return r
}
