package handler

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.bryk.io/pkg/errors"
	otelApi "go.bryk.io/pkg/otel/api"
)

// ServiceOperator provides an implementation for all functional
// requirements (i.e., business logic) required to cover the service
// scope.
type ServiceOperator struct{}

// New service operator instance.
func New() (*ServiceOperator, error) {
	return &ServiceOperator{}, nil
}

// Ping provides a basic reachability test.
func (so *ServiceOperator) Ping() bool {
	return true
}

// Ready returns "true" if the service is able to receive and
// process requests.
func (so *ServiceOperator) Ready() bool {
	return true
}

// Echo returns the same message received as input.
func (so *ServiceOperator) Echo(ctx context.Context, msg string) (string, error) {
	span := otelApi.Start(ctx, "echo handler")
	defer span.End(nil)

	return fmt.Sprintf("you said: %s", msg), nil
}

// Slow is a method that exhibit a random latency between 10 and 200ms.
func (so *ServiceOperator) Slow(ctx context.Context) error {
	span := otelApi.Start(ctx, "slow handler")
	defer span.End(nil)

	delay := rand.Intn(180) + 20 // nolint:gosec
	attrs := otelApi.AsWarning()
	attrs.Set("app.delay", delay)
	span.Event("waiting for slow operation", attrs)
	<-time.After(time.Duration(delay) * time.Millisecond)

	return nil
}

// Faulty is a method that returns an error roughly
// 50% of the time.
func (so *ServiceOperator) Faulty(ctx context.Context) error {
	span := otelApi.Start(ctx, "faulty handler")
	defer span.End(nil)

	check := rand.Intn(20) + 1 // nolint:gosec
	if check%2 == 0 {
		err := errors.New("random error")
		attrs := otelApi.AsWarning()
		attrs.Set("random.value", check)
		span.Event("bad luck", attrs)
		span.End(err)
		return err
	}

	attrs := otelApi.AsInfo()
	attrs.Set("random.value", check)
	span.Event("good luck", attrs)
	return nil
}

// Reload the operator instance by refreshing or re-establishing
// any internal resources or dependencies.
func (so *ServiceOperator) Reload() error {
	return nil
}

// Close the service operator gracefully and free any used/locked
// resources.
func (so *ServiceOperator) Close() error {
	return nil
}
