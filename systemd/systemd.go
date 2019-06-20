package systemd

import (
	"time"

	"github.com/coreos/go-systemd/dbus"
	"github.com/coreos/go-systemd/machine1"
	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/drivers"
)

var (
	dbusConn     *dbus.Conn
	machinedConn *machine1.Conn
)

// Machine Object in dbus.
//
// node /org/freedesktop/machine1/machine/fedora_2dtree {
//  interface org.freedesktop.machine1.Machine {
//    methods:
//      Terminate();
//      Kill(in  s who,
//           in  s signal);
//      GetAddresses(out a(iay) addresses);
//      GetOSRelease(out a{ss} fields);
//    signals:
//    properties:
//      readonly s Name = 'fedora-tree';
//      readonly ay Id = [0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00];
//      readonly t Timestamp = 1374193370484284;
//      readonly t TimestampMonotonic = 128247251308;
//      readonly s Service = 'nspawn';
//      readonly s Unit = 'machine-fedora\\x2dtree.scope';
//      readonly u Leader = 30046;
//      readonly s Class = 'container';
//      readonly s RootDirectory = '/home/lennart/fedora-tree';
//      readonly ai NetworkInterfaces = [7];
//      readonly s State = 'running';
//  };
//  interface org.freedesktop.DBus.Properties {
//  };
//  interface org.freedesktop.DBus.Peer {
//  };
//  interface org.freedesktop.DBus.Introspectable {
//  };
//};
type Machine struct {
	Name               string
	ID                 []byte
	Timestamp          time.Time
	TimestampMonotonic time.Time
	Service            string
	Unit               string
	Leader             int
	Class              string
	RootDirectory      string
	NetworkInterfaces  []int
	State              string
}

// Available state for machine.
const (
	MachineStateOpening = "opening"
	MachineStateRunning = "running"
	MachineStateClosing = "closing"
)

// Available class for machine.
const (
	MachineClassContainer = "container"
	MachineClassVM        = "vm"
)

// CreateMachine will create a new systemd-nspawn machine.
func (d *Driver) CreateMachine(cfg *drivers.TaskConfig, taskConfig TaskConfig) (m *Machine, err error) {
	panic("implement me")
}

// CreateMachineWithNetwork will create a new systemd-nspawn machine with network.
func (d *Driver) CreateMachineWithNetwork() {
	panic("implement me")
}

// GetMachine will get a new systemd-nspawn machine.
func (d *Driver) GetMachine() {
	panic("implement me")
}

// KillMachine will kill a new systemd-nspawn machine.
func (d *Driver) KillMachine() {
	panic("implement me")
}

// TerminateMachine will terminate a new systemd-nspawn machine.
func (d *Driver) TerminateMachine() {
	panic("implement me")
}

func (d *Driver) getMachine() {
	panic("implement me")
}

func init() {
	var err error
	dbusConn, err = dbus.New()
	if err != nil {
		log.Default().Error("systemd connected failed", err)
	}

	machinedConn, err = machine1.New()
	if err != nil {
		log.Default().Error("systemd-machined connected failed", err)
	}
}
