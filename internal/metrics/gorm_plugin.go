package metrics

import (
	"time"

	"gorm.io/gorm"
)

const metricsCallbackName = "metrics"

type GormMetricsPlugin struct{}

func (p *GormMetricsPlugin) Name() string {
	return "prometheus_metrics"
}

func (p *GormMetricsPlugin) Initialize(db *gorm.DB) error {
	// Before callbacks \ store start time
	for _, op := range []string{"create", "query", "update", "delete", "row"} {
		callbackName := metricsCallbackName + ":before_" + op
		var err error
		switch op {
		case "create":
			err = db.Callback().Create().Before("gorm:create").Register(callbackName, beforeCallback)
		case "query":
			err = db.Callback().Query().Before("gorm:query").Register(callbackName, beforeCallback)
		case "update":
			err = db.Callback().Update().Before("gorm:update").Register(callbackName, beforeCallback)
		case "delete":
			err = db.Callback().Delete().Before("gorm:delete").Register(callbackName, beforeCallback)
		case "row":
			err = db.Callback().Row().Before("gorm:row").Register(callbackName, beforeCallback)
		}
		if err != nil {
			return err
		}
	}

	// After callbacks - record metrics
	for _, op := range []string{"create", "query", "update", "delete", "row"} {
		callbackName := metricsCallbackName + ":after_" + op
		opLabel := op
		if op == "row" {
			opLabel = "query"
		}
		afterFn := makeAfterCallback(opLabel)
		var err error
		switch op {
		case "create":
			err = db.Callback().Create().After("gorm:create").Register(callbackName, afterFn)
		case "query":
			err = db.Callback().Query().After("gorm:query").Register(callbackName, afterFn)
		case "update":
			err = db.Callback().Update().After("gorm:update").Register(callbackName, afterFn)
		case "delete":
			err = db.Callback().Delete().After("gorm:delete").Register(callbackName, afterFn)
		case "row":
			err = db.Callback().Row().After("gorm:row").Register(callbackName, afterFn)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

type gormStartTimeKey struct{}

func beforeCallback(db *gorm.DB) {
	db.Set("metrics:start_time", time.Now())
}

func makeAfterCallback(operation string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		table := db.Statement.Table
		if table == "" {
			table = "unknown"
		}

		// Record duration
		if v, ok := db.Get("metrics:start_time"); ok {
			if startTime, ok := v.(time.Time); ok {
				duration := time.Since(startTime).Seconds()
				DbQueryDuration.WithLabelValues(operation, table).Observe(duration)
			}
		}

		// Record query count
		DbQueriesTotal.WithLabelValues(operation, table).Inc()

		// Record errors
		if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
			DbErrorsTotal.WithLabelValues(operation, table).Inc()
		}
	}
}
