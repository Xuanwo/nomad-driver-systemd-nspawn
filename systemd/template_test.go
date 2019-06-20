package systemd

import (
	"bytes"
	"testing"
)

const result = `[Exec]
Boot=on
Ephemeral=off
ProcessTwo=off
Parameters=1,2,3
Environment=1=2
Environment=a=b
User=abc
WorkingDirectory=
PivotRoot=
Capability=1 2 3
DropCapability=
NoNewPrivileges=off
KillSignal=127
Personality=
MachineID=
PrivateUsers=
NotifyReady=off
SystemCallFilter=
LimitCPU=
LimitFSIZE=
LimitDATA=
LimitSTACK=
LimitCORE=
LimitRSS=
LimitNOFILE=
LimitAS=
LimitNPROC=
LimitMEMLOCK=
LimitLOCKS=
LimitSIGPENDING=
LimitMSGQUEUE=
LimitNICE=
LimitRTPRIO=
LimitRTTIME=
OOMScoreAdjust=1
CPUAffinity=
Hostname=
ResolvConf=
Timezone=
LinkJournal=

[Files]
ReadOnly=off
Volatile=
Overlay=1:2:3
Overlay=2:4:6
PrivateUsersChown=off

[Network]
Private=off
VirtualEthernet=off
Interface=1 2 3
MACVLAN=
IPVLAN=
Bridge=
Zone=
`

func TestTemplate(t *testing.T) {
	data := TaskConfig{
		Boot:       true,
		Parameters: []string{"1", "2", "3"},
		Environment: map[string]string{
			"a": "b",
			"1": "2",
		},
		User:           "abc",
		Capability:     []string{"1", "2", "3"},
		KillSignal:     127,
		OOMScoreAdjust: 1,
		Overlay:        [][]string{{"1", "2", "3"}, {"2", "4", "6"}},
	}

	buf := bytes.NewBuffer(make([]byte, 0))

	err := tmpl.Execute(buf, data)
	if err != nil {
		t.Error(err)
	}
	t.Log(buf.String())

	if buf.String() != result {
		t.Error("template generated wrongly")
	}
}
