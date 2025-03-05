//go:build linux
// +build linux

// LINUX(ChassisModule) ok
// WINDOWS(ChassisModule) no
// MACOS(ChassisModule) ?
// ROOT(ChassisModule) ?
package modules

import (
	"os"

	"github.com/godbus/dbus/v5"
	"github.com/situation-sh/situation/store"
)

func init() {
	RegisterModule(&ChassisModule{})
}

// ChassisModule fills host chassis information
//
// Currently it only works under linux. It uses DBUS and the "org.freedesktop.hostname1"
// service to get the type of the chassis (like laptop, vm, desktop etc.)
// In the future it may rather rely on [ghw] but at that time
// it does not fully get the info on windows.
//
// [ghw]: https://github.com/jaypipes/ghw/
type ChassisModule struct{}

func (m *ChassisModule) Name() string {
	return "chassis"
}

func (m *ChassisModule) Dependencies() []string {
	// depends on host-basic to get the host machine
	return []string{"host-basic"}
}

func (m *ChassisModule) Run() error {
	logger := GetLogger(m)

	if !isSocketAvailable() {
		logger.Warn("dbus system bus is not available, skipping")
		return &notApplicableError{"dbus system bus is not available"}
	}

	host := store.GetHost()

	// Connect to the system bus
	logger.Debug("opening system bus")
	conn, err := dbus.SystemBus()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Call the method to get the property
	obj := conn.Object("org.freedesktop.hostname1", "/org/freedesktop/hostname1")
	logger.Debug("getting chassis from org.freedesktop.hostname1")
	err = obj.Call("org.freedesktop.DBus.Properties.Get", 0, "org.freedesktop.hostname1", "Chassis").Store(&host.Chassis)
	if err != nil {
		return err
	}
	logger.WithField("chassis", host.Chassis).Info("chassis found through dbus")

	return nil
}

func isSocketAvailable() bool {
	_, err := os.Stat("/var/run/dbus/system_bus_socket")
	return !os.IsNotExist(err)
}
