package health

import "context"

type Checker interface {
	Name() string
	Check(ctx context.Context) CheckResult
}
