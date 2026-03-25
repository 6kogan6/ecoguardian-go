package router

import (
	"fmt"
	"net/http"
	"time"

	"ecoguardian-go/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
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
				"version": "0.1.0",
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

			alerts := buildAlerts(reading)
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
	}

	return r
}

func buildAlerts(reading model.Reading) []model.Alert {
	alerts := make([]model.Alert, 0)

	addAlert := func(level string, message string) {
		alerts = append(alerts, model.Alert{
			SensorID:  reading.SensorID,
			Level:     level,
			Message:   message,
			Status:    "active",
			CreatedAt: time.Now(),
		})
	}

	if reading.PM25 > 55 {
		addAlert("danger", fmt.Sprintf("Critical PM2.5 level: %.2f", reading.PM25))
	} else if reading.PM25 > 35 {
		addAlert("warning", fmt.Sprintf("High PM2.5 level: %.2f", reading.PM25))
	}

	if reading.CO2 > 1500 {
		addAlert("danger", fmt.Sprintf("Critical CO2 level: %.2f", reading.CO2))
	} else if reading.CO2 > 1000 {
		addAlert("warning", fmt.Sprintf("High CO2 level: %.2f", reading.CO2))
	}

	if reading.Noise > 85 {
		addAlert("danger", fmt.Sprintf("Critical noise level: %.2f", reading.Noise))
	} else if reading.Noise > 70 {
		addAlert("warning", fmt.Sprintf("High noise level: %.2f", reading.Noise))
	}

	if reading.WaterPH > 0 && (reading.WaterPH < 6.5 || reading.WaterPH > 8.5) {
		addAlert("warning", fmt.Sprintf("Water pH out of normal range: %.2f", reading.WaterPH))
	}

	return alerts
}
