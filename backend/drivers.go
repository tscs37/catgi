package backend

import (
	"context"
	"errors"
	"fmt"

	"git.timschuster.info/rls.moe/catgi/backend/common"
)

// Backend Type Alias for easier importing
type Backend common.Backend

type driverCreator func(map[string]interface{}, context.Context) (common.Backend, error)

var backendDrivers = map[string]driverCreator{}

type noDriverError struct {
	drvName string
}

func (n noDriverError) Error() string {
	return fmt.Sprintf("Driver '%s' not installed", n.drvName)
}

func newNoDriverError(drv string) error {
	return noDriverError{drvName: drv}
}

func NewBackend(
	driver string, params map[string]interface{}, ctx context.Context) (common.Backend, error) {
	if f, ok := backendDrivers[driver]; ok {
		return f(params, ctx)
	}
	return nil, newNoDriverError(driver)
}

func InstalledDrivers() []string {
	var list = []string{}
	for v := range backendDrivers {
		list = append(list, v)
	}
	return list
}

func NewDriver(driver string, dfunc interface{}) error {
	switch v := dfunc.(type) {
	case func(map[string]interface{}, context.Context) (common.Backend, error):
		backendDrivers[driver] = v
		return nil
	default:
		return errors.New("Not a known driver type")
	}
}
