//go:build linux

// LINUX(ChassisModule) ok
// WINDOWS(ChassisModule) no
// MACOS(ChassisModule) ?
// ROOT(ChassisModule) ?
package modules

import (
	"context"
	"fmt"
	"os"

	"github.com/godbus/dbus/v5"
	"github.com/situation-sh/situation/pkg/models"
)

func init() {
	registerModule(&ChassisModule{})
}

// ChassisModule fills host chassis information
//
// Currently it only works under linux. It uses DBUS and the "org.freedesktop.hostname1"
// service to get the type of the chassis (like laptop, vm, desktop etc.)
// In the future it may rather rely on [ghw] but at that time
// it does not fully get the info on windows.
//
// [ghw]: https://github.com/jaypipes/ghw/
type ChassisModule struct {
	BaseModule
}

func (m *ChassisModule) Name() string {
	return "chassis"
}

func (m *ChassisModule) Dependencies() []string {
	// depends on host-basic to get the host machine
	return []string{"host-basic"}
}

func (m *ChassisModule) Run(ctx context.Context) error {
	logger := getLogger(ctx, m)
	storage := getStorage(ctx)

	if !isSocketAvailable() {
		logger.Warn("dbus system bus is not available, skipping")
		return &notApplicableError{"dbus system bus is not available"}
	}

	hostID := storage.GetHostID(ctx)
	if hostID == 0 {
		return fmt.Errorf("no host found in storage")
	}

	// Connect to the system bus
	logger.Debug("opening system bus")
	conn, err := dbus.SystemBus()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Call the method to get the property
	chassis := ""
	obj := conn.Object("org.freedesktop.hostname1", "/org/freedesktop/hostname1")
	logger.Debug("getting chassis from org.freedesktop.hostname1")
	err = obj.Call("org.freedesktop.DBus.Properties.Get", 0, "org.freedesktop.hostname1", "Chassis").
		Store(&chassis)
	if err != nil {
		return err
	}

	logger.
		WithField("chassis", chassis).
		Info("chassis found through dbus")
	_, err = storage.DB().
		NewUpdate().
		Model((*models.Machine)(nil)).
		Where("id = ?", hostID).
		Set("chassis = ?", chassis).
		Exec(ctx)

	return err
}

func isSocketAvailable() bool {
	_, err := os.Stat("/var/run/dbus/system_bus_socket")
	return !os.IsNotExist(err)
}
