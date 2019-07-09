package systemd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-systemd/dbus"
	"github.com/coreos/go-systemd/import1"
	"github.com/coreos/go-systemd/machine1"
	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/drivers"
)

var (
	dbusConn     *dbus.Conn
	machinedConn *machine1.Conn
	importdConn  *import1.Conn
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
	machineName := fmt.Sprintf("%s-%s", strings.Replace(cfg.Name, "/", "_", -1), cfg.AllocID)

	trans, err := importdConn.PullRaw(taskConfig.Image, machineName, "no", false)
	if err != nil {
		return
	}

	// FIXME: So stupid, let's use signal instead.
	for {
		ts, err := importdConn.ListTransfers()
		if err != nil {
			return nil, err
		}
		found := false
		for _, v := range ts {
			if v.Id == trans.Id {
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	// Create nspawn file.
	f, err := os.Create("/etc/systemd/nspawn/" + machineName)
	if err != nil {
		d.logger.Error("Create nspawn file failed", "error", err)
		return
	}
	defer f.Close()

	err = tmpl.Execute(f, taskConfig)
	if err != nil {
		d.logger.Error("Generate nspawn file failed", "error", err)
		return
	}

	// Start machine along with image and nspawn file.
	ch := make(chan string)
	defer close(ch)
	_, err = dbusConn.StartUnit(
		fmt.Sprintf("systemd-nspawn@%s.service", machineName), "replace", ch)
	if err != nil {
		d.logger.Error("Create machine unit failed", "error", err)
		return
	}

	job := <-ch
	if job != "done" {
		d.logger.Error("Start machine unit failed")
	}

	return
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

	importdConn, err = import1.New()
	if err != nil {
		log.Default().Error("systemd-importd connected failed", err)
	}
}
