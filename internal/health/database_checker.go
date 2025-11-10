package health

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type DatabaseChecker struct {
	db *gorm.DB
}

func NewDatabaseChecker(db *gorm.DB) *DatabaseChecker {
	return &DatabaseChecker{db: db}
}

func (d *DatabaseChecker) Name() string {
	return "database"
}

func (d *DatabaseChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	sqlDB, err := d.db.DB()
	if err != nil {
		return CheckResult{
			Status:  CheckFail,
			Message: "Failed to get database instance",
		}
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return CheckResult{
			Status:  CheckFail,
			Message: "Database connection failed",
		}
	}

	var result int
	if err := d.db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error; err != nil {
		return CheckResult{
			Status:  CheckFail,
			Message: "Database query failed",
		}
	}

	duration := time.Since(start)
	status := CheckPass
	message := "Database connection healthy"

	if duration > 500*time.Millisecond {
		status = CheckFail
		message = "Database response time too slow"
	} else if duration > 100*time.Millisecond {
		status = CheckWarn
		message = "Database response time degraded"
	}

	return CheckResult{
		Status:       status,
		Message:      message,
		ResponseTime: fmt.Sprintf("%dms", duration.Milliseconds()),
	}
}
