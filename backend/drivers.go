package backend

import (
	"context"
	"errors"
	"fmt"

	"git.timschuster.info/rls.moe/catgi/backend/types"
)

// Backend Type Alias for easier importing
type Backend types.Backend
type KVBackend types.KVBackend
type ContentBackend types.ContentBackend

type driverCreator func(map[string]interface{}, context.Context) (types.Backend, error)

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
	driver string, params map[string]interface{}, ctx context.Context) (types.Backend, error) {
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
	case func(map[string]interface{}, context.Context) (types.Backend, error):
		backendDrivers[driver] = v
		return nil
	default:
		return errors.New("Not a known driver type")
	}
}
