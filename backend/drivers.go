package backend

import (
	"context"
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

// NewBackend initializes the backend named via driver-name with the given
// parameter mapping. The context is used for logging purposes.
// If the driver does not exist it returns an error.
func NewBackend(
	driver string, params map[string]interface{}, ctx context.Context) (common.Backend, error) {
	if f, ok := backendDrivers[driver]; ok {
		return f(params, ctx)
	}
	return nil, newNoDriverError(driver)
}

// InstalledDrivers returns a list of all drivers that are currently installed.
func InstalledDrivers() []string {
	var list = []string{}
	for v := range backendDrivers {
		list = append(list, v)
	}
	return list
}

// NewDriver accepts a backend init function and saves it into the list
// of installed backend drivers.
func NewDriver(driver string,
	dfunc func(map[string]interface{}, context.Context) (common.Backend, error)) {
	backendDrivers[driver] = dfunc
}
