package systemd

import (
	"context"
	"time"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/drivers/shared/eventer"
	"github.com/hashicorp/nomad/plugins/base"
	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/hashicorp/nomad/plugins/shared/hclspec"
)

const (
	// pluginName is the name of the plugin
	pluginName = "systemd-nspawn"
)

var (
	// pluginInfo is the response returned for the PluginInfo RPC
	pluginInfo = &base.PluginInfoResponse{
		Type:              base.PluginTypeDriver,
		PluginApiVersions: []string{drivers.ApiVersion010},
		PluginVersion:     "0.1.0",
		Name:              pluginName,
	}

	// configSpec is the hcl specification returned by the ConfigSchema RPC
	configSpec = hclspec.NewObject(map[string]*hclspec.Spec{
		"enabled": hclspec.NewDefault(
			hclspec.NewAttr("enabled", "bool", false),
			hclspec.NewLiteral("true"),
		),
	})

	// taskConfigSpec is the hcl specification for the driver config section of
	// a task within a job. It is returned in the TaskConfigSchema RPC
	taskConfigSpec = hclspec.NewObject(map[string]*hclspec.Spec{
		"template": hclspec.NewAttr("template", "string", true),
	})

	// capabilities is returned by the Capabilities RPC and indicates what
	// optional features this driver supports
	capabilities = &drivers.Capabilities{
		Exec: false,
	}
)

// Driver is a driver for running systemd nspawn containers
type Driver struct {
	// eventer is used to handle multiplexing of TaskEvents calls such that an
	// event can be broadcast to all callers
	eventer *eventer.Eventer

	// config is the driver configuration set by the SetConfig RPC
	config *Config

	// nomadConfig is the client config from nomad
	nomadConfig *base.ClientDriverConfig

	// ctx is the context for the driver. It is passed to other subsystems to
	// coordinate shutdown
	ctx context.Context

	// signalShutdown is called when the driver is shutting down and cancels the
	// ctx passed to any subsystems
	signalShutdown context.CancelFunc

	// logger will log to the Nomad agent
	logger log.Logger
}

// Config is the driver configuration set by the SetConfig RPC call
type Config struct {
	// Enabled is set to true to enable the systemd driver
	Enabled bool `codec:"enabled"`
}

// TaskConfig is the driver configuration of a task within a job
type TaskConfig struct {
	// Image section

	// Image is the image name.
	Image string

	// Exec section

	// Boot takes a boolean argument, which defaults to off.
	// If enabled, systemd-nspawn will automatically search for an init executable and invoke it.
	// In this case, the specified parameters using Parameters= are passed as additional arguments to the init process.
	// This option may not be combined with ProcessTwo=yes.
	Boot bool
	// Ephemeral takes a boolean argument, which defaults to off, If enabled, the container is run with a temporary
	// snapshot of its file system that is removed immediately when the container terminates.
	Ephemeral bool
	// ProcessTwo takes a boolean argument, which defaults to off.
	// If enabled, the specified program is run as PID 2.
	// A stub init process is run as PID 1.
	// This option may not be combined with Boot=yes.
	ProcessTwo bool
	// Parameters takes a space-separated list of arguments.
	// This is either a command line, beginning with the binary name to execute,
	// or – if Boot= is enabled – the list of arguments to pass to the init process.
	Parameters []string
	// Environment takes an environment variable assignment consisting of key and value.
	// Sets an environment variable for the main process invoked in the container.
	// This setting may be used multiple times to set multiple environment variables.
	Environment map[string]string
	// User takes a UNIX user name.
	// Specifies the user name to invoke the main process of the container as.
	// This user must be known in the container's user database.
	User string
	// WorkingDirectory selects the working directory for the process invoked in the container.
	// Expects an absolute path in the container's file system namespace.
	WorkingDirectory string
	// PivotRoot selects a directory to pivot to / inside the container when starting up.
	// Takes a single path, or a pair of two paths separated by a colon.
	// Both paths must be absolute, and are resolved in the container's file system namespace.
	PivotRoot string
	// Capability takes a list of Linux process capabilities (see capabilities(7) for details).
	// The Capability= setting specifies additional capabilities to pass on top of the default set of capabilities.
	// The DropCapability= setting specifies capabilities to drop from the default set.
	Capability []string
	// DropCapability used like Capability.
	DropCapability []string
	// NoNewPrivileges takes a boolean argument that controls the PR_SET_NO_NEW_PRIVS flag for the container payload.
	// ref: https://www.freedesktop.org/software/systemd/man/systemd-nspawn.html#--no-new-privileges=
	NoNewPrivileges bool
	// KillSignal specify the process signal to send to the container's PID 1 when nspawn itself receives SIGTERM,
	// in order to trigger an orderly shutdown of the container.
	// Defaults to SIGRTMIN+3 if Boot= is used (on systemd-compatible init systems SIGRTMIN+3 triggers an
	// orderly shutdown).
	// For a list of valid signals, see signal(7).
	KillSignal uint32
	// Personality configures the kernel personality for the container.
	// Currently, "x86" and "x86-64" are supported.
	// ref: https://www.freedesktop.org/software/systemd/man/systemd-nspawn.html#--personality=
	Personality string
	// MachineID configures the 128-bit machine ID (UUID) to pass to the container.
	MachineID string
	// PrivateUsers configures support for usernamespacing.
	// ref: https://www.freedesktop.org/software/systemd/man/systemd-nspawn.html#--private-users=
	PrivateUsers string
	// NotifyReady configures support for notifications from the container's init process.
	// ref: https://www.freedesktop.org/software/systemd/man/systemd-nspawn.html#--notify-ready=
	NotifyReady bool
	// SystemCallFilter configures the system call filter applied to containers.
	// ref: https://www.freedesktop.org/software/systemd/man/systemd-nspawn.html#--system-call-filter=
	SystemCallFilter []string
	// Configures various types of resource limits applied to containers.
	// Sets the specified POSIX resource limit for the container payload.
	// Expects an assignment of the form "SOFT:HARD" or "VALUE"
	// ref: https://www.freedesktop.org/software/systemd/man/systemd-nspawn.html#--rlimit=
	LimitCPU        string
	LimitFSIZE      string
	LimitDATA       string
	LimitSTACK      string
	LimitCORE       string
	LimitRSS        string
	LimitNOFILE     string
	LimitAS         string
	LimitNPROC      string
	LimitMEMLOCK    string
	LimitLOCKS      string
	LimitSIGPENDING string
	LimitMSGQUEUE   string
	LimitNICE       string
	LimitRTPRIO     string
	LimitRTTIME     string
	// OOMScoreAdjust changes the OOM ("Out Of Memory") score adjustment value for the container payload.
	// This controls /proc/self/oom_score_adj which influences the preference with which this container
	// is terminated when memory becomes scarce.
	// For details see proc(5).
	// Takes an integer in the range -1000…1000.
	OOMScoreAdjust int
	// CPUAffinity controls the CPU affinity of the container payload.
	// Takes a comma separated list of CPU numbers or number ranges (the latter's start and end value separated by
	// dashes).
	// See sched_setaffinity(2) for details.
	CPUAffinity []string
	// Hostname configures the kernel hostname set for the container.
	Hostname string
	// ResolvConf configures how /etc/resolv.conf inside of the container (i.e. DNS configuration synchronization from
	// host to container) shall be handled.
	// Takes one of "off", "copy-host", "copy-static", "bind-host", "bind-static", "delete" or "auto".
	// ref: https://www.freedesktop.org/software/systemd/man/systemd-nspawn.html#--resolv-conf=
	ResolvConf string
	// Timezone configures how /etc/localtime inside of the container (i.e. local timezone synchronization from host
	// to container) shall be handled.
	// Takes one of "off", "copy", "bind", "symlink", "delete" or "auto".
	// ref: https://www.freedesktop.org/software/systemd/man/systemd-nspawn.html#--timezone=
	Timezone string
	// LinkJournal controls whether the container's journal shall be made visible to the host system.
	// If enabled, allows viewing the container's journal files from the host (but not vice versa).
	// Takes one of "no", "host", "try-host", "guest", "try-guest", "auto".
	LinkJournal string

	// Files section

	// ReadOnly takes a boolean argument, which defaults to off.
	// If specified, the container will be run with a read-only file system.
	ReadOnly bool
	// Volatile takes "no", "yes", or the special value "state".
	// This configures whether to run the container with volatile state and/or configuration.
	// ref: https://www.freedesktop.org/software/systemd/man/systemd-nspawn.html#--volatile
	Volatile string
	// Bind adds a bind mount from the host into the container.
	// Takes a single path, a pair of two paths separated by a colon, or a triplet of two paths plus an
	// option string separated by colons.
	Bind         []string
	BindReadOnly []string
	// TemporaryFileSystem adds a "tmpfs" mount to the container.
	// Takes a path or a pair of path and option string, separated by a colon.
	TemporaryFileSystem []string
	// Inaccessible masks the specified file or directly in the container, by over-mounting it with an empty file node of
	// the same type with the most restrictive access mode.
	// Takes a file system path as arugment.
	Inaccessible []string
	// Overlay adds an overlay mount point.
	// Takes a colon-separated list of paths.
	Overlay         [][]string
	OverlayReadOnly [][]string
	// PrivateUsersChown configures whether the ownership of the files and directories in the container tree shall be adjusted
	// to the UID/GID range used, if necessary and user namespacing is enabled.
	PrivateUsersChown bool

	// Network section

	// Private takes a boolean argument, which defaults to off.
	// If enabled, the container will run in its own network namespace and not share network interfaces
	// and configuration with the host.
	Private bool
	// VirtualEthernet takes a boolean argument.
	// Configures whether to create a virtual Ethernet connection ("veth") between host and the container.
	// This setting implies Private=yes.
	VirtualEthernet bool
	// VirtualEthernetExtra takes a colon-separated pair of interface names.
	// Configures an additional virtual Ethernet connection ("veth") between host and the container.
	// The first specified name is the interface name on the host, the second the interface name in the container.
	// The latter may be omitted in which case it is set to the same name as the host side interface.
	// This setting implies Private=yes.
	// It is independent of VirtualEthernet=. This option is privileged.
	VirtualEthernetExtra []string
	// Interface takes a space-separated list of interfaces to add to the container.
	// This option implies Private=yes.
	Interface []string
	// MACVLAN and IPVLAN takes a space-separated list of interfaces to add MACLVAN or IPVLAN interfaces to,
	// which are then added to the container.
	// These options correspond to the --network-macvlan= and --network-ipvlan= command line switches and
	// imply Private=yes.
	// These options are privileged.
	MACVLAN []string
	IPVLAN  []string
	// Bridge takes an interface name.
	// This setting implies VirtualEthernet=yes and Private=yes and has the effect that the host side of the
	// created virtual Ethernet link is connected to the specified bridge interface.
	// This option is privileged.
	Bridge string
	// Zone takes a network zone name.
	// This setting implies VirtualEthernet=yes and Private=yes and has the effect that the host side of the
	// created virtual Ethernet link is connected to an automatically managed bridge interface named after
	// the passed argument, prefixed with "vz-".
	// This option is privileged.
	Zone string
	// Port exposes a TCP or UDP port of the container on the host.
	// If private networking is enabled, maps an IP port on the host onto an IP port on the container.
	// Takes a protocol specifier (either "tcp" or "udp"), separated by a colon from a host port number in the
	// range 1 to 65535, separated by a colon from a container port number in the range from 1 to 65535.
	// The protocol specifier and its separating colon may be omitted, in which case "tcp" is assumed.
	// The container port number and its colon may be omitted, in which case the same port as the host port is
	// implied.
	// This option is only supported if private networking is used, such as with --network-veth,
	// --network-zone= --network-bridge=.
	// This option is privileged.
	Port []string
}

// TaskState is the state which is encoded in the handle returned in
// StartTask. This information is needed to rebuild the task state and handler
// during recovery.
type TaskState struct {
	TaskConfig  *drivers.TaskConfig
	MachineName string
	StartedAt   time.Time
}

// NewSystemdNSpawnDriver returns a new DriverPlugin implementation
func NewSystemdNSpawnDriver(logger log.Logger) drivers.DriverPlugin {
	ctx, cancel := context.WithCancel(context.Background())
	logger = logger.Named(pluginName)
	return &Driver{
		eventer:        eventer.NewEventer(ctx, logger),
		config:         &Config{},
		ctx:            ctx,
		signalShutdown: cancel,
		logger:         logger,
	}
}

// PluginInfo implements BasePlugin's PluginInfo.
func (d *Driver) PluginInfo() (*base.PluginInfoResponse, error) {
	return pluginInfo, nil
}

// ConfigSchema implements BasePlugin's ConfigSchema.
func (d *Driver) ConfigSchema() (*hclspec.Spec, error) {
	return configSpec, nil
}

// SetConfig implements BasePlugin's SetConfig.
func (d *Driver) SetConfig(cfg *base.Config) error {
	var config Config
	if len(cfg.PluginConfig) != 0 {
		if err := base.MsgPackDecode(cfg.PluginConfig, &config); err != nil {
			return err
		}
	}

	d.config = &config
	if cfg.AgentConfig != nil {
		d.nomadConfig = cfg.AgentConfig.Driver
	}

	return nil
}

// Shutdown will shutdown current driver.
func (d *Driver) Shutdown(ctx context.Context) error {
	panic("implement me")
}

// TaskConfigSchema implements DriverPlugin's TaskConfigSchema.
func (d *Driver) TaskConfigSchema() (*hclspec.Spec, error) {
	panic("implement me")
}

// Capabilities implements DriverPlugin's Capabilities.
func (d *Driver) Capabilities() (*drivers.Capabilities, error) {
	panic("implement me")
}

// Fingerprint implements DriverPlugin's Fingerprint.
func (d *Driver) Fingerprint(ctx context.Context) (<-chan *drivers.Fingerprint, error) {
	panic("implement me")
}

// RecoverTask implements DriverPlugin's RecoverTask.
func (d *Driver) RecoverTask(handle *drivers.TaskHandle) error {
	panic("implement me")
}

// StartTask implements DriverPlugin's StartTask.
func (d *Driver) StartTask(cfg *drivers.TaskConfig) (*drivers.TaskHandle, *drivers.DriverNetwork, error) {
	panic("implement me")
}

// WaitTask implements DriverPlugin's WaitTask.
func (d *Driver) WaitTask(ctx context.Context, taskID string) (<-chan *drivers.ExitResult, error) {
	panic("implement me")
}

// StopTask implements DriverPlugin's StopTask.
func (d *Driver) StopTask(taskID string, timeout time.Duration, signal string) error {
	panic("implement me")
}

// DestroyTask implements DriverPlugin's DestroyTask.
func (d *Driver) DestroyTask(taskID string, force bool) error {
	panic("implement me")
}

// InspectTask implements DriverPlugin's InspectTask.
func (d *Driver) InspectTask(taskID string) (*drivers.TaskStatus, error) {
	panic("implement me")
}

// TaskStats implements DriverPlugin's TaskStats.
func (d *Driver) TaskStats(ctx context.Context, taskID string, interval time.Duration) (<-chan *drivers.TaskResourceUsage, error) {
	panic("implement me")
}

// TaskEvents implements DriverPlugin's TaskEvents.
func (d *Driver) TaskEvents(ctx context.Context) (<-chan *drivers.TaskEvent, error) {
	panic("implement me")
}

// SignalTask implements DriverPlugin's SignalTask.
func (d *Driver) SignalTask(taskID string, signal string) error {
	panic("implement me")
}

// ExecTask implements DriverPlugin's ExecTask.
func (d *Driver) ExecTask(taskID string, cmd []string, timeout time.Duration) (*drivers.ExecTaskResult, error) {
	panic("implement me")
}
