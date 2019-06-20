package systemd

import (
	"strings"
	"text/template"
)

var funcMaps = template.FuncMap{
	"join": strings.Join,
}

const nspawnTemplate = `[Exec]
Boot={{if .Boot}}on{{else}}off{{end}}
Ephemeral={{if .Ephemeral}}on{{else}}off{{end}}
ProcessTwo={{if .ProcessTwo}}on{{else}}off{{end}}
Parameters={{join .Parameters ","}}
{{- range $k, $v := .Environment }}
Environment={{$k}}={{$v}}
{{- end }}
User={{ .User }}
WorkingDirectory={{ .WorkingDirectory }}
PivotRoot={{ .PivotRoot }}
Capability={{join .Capability " "}}
DropCapability={{join .DropCapability " "}}
NoNewPrivileges={{if .NoNewPrivileges}}on{{else}}off{{end}}
KillSignal={{ .KillSignal }}
Personality={{ .Personality }}
MachineID={{ .MachineID }}
PrivateUsers={{ .PrivateUsers }}
NotifyReady={{if .NotifyReady}}on{{else}}off{{end}}
SystemCallFilter={{join .SystemCallFilter " "}}
LimitCPU={{ .LimitCPU }}
LimitFSIZE={{ .LimitFSIZE }}
LimitDATA={{ .LimitDATA }}
LimitSTACK={{ .LimitSTACK }}
LimitCORE={{ .LimitCORE }}
LimitRSS={{ .LimitRSS }}
LimitNOFILE={{ .LimitNOFILE }}
LimitAS={{ .LimitAS }}
LimitNPROC={{ .LimitNPROC }}
LimitMEMLOCK={{ .LimitMEMLOCK }}
LimitLOCKS={{ .LimitLOCKS }}
LimitSIGPENDING={{ .LimitSIGPENDING }}
LimitMSGQUEUE={{ .LimitMSGQUEUE }}
LimitNICE={{ .LimitNICE }}
LimitRTPRIO={{ .LimitRTPRIO }}
LimitRTTIME={{ .LimitRTTIME }}
OOMScoreAdjust={{ .OOMScoreAdjust }}
CPUAffinity={{join .CPUAffinity ","}}
Hostname={{ .Hostname }}
ResolvConf={{ .ResolvConf }}
Timezone={{ .Timezone }}
LinkJournal={{ .LinkJournal }}

[Files]
ReadOnly={{if .ReadOnly}}on{{else}}off{{end}}
Volatile={{ .Volatile}}
{{- range $_, $v := .Bind }}
Bind={{$v}}
{{- end }}
{{- range $_, $v := .BindReadOnly }}
BindReadOnly={{$v}}
{{- end }}
{{- range $_, $v := .TemporaryFileSystem }}
TemporaryFileSystem={{$v}}
{{- end }}
{{- range $_, $v := .Inaccessible }}
Inaccessible={{$v}}
{{- end }}
{{- range $_, $v := .Overlay }}
Overlay={{join $v ":"}}
{{- end }}
{{- range $_, $v := .OverlayReadOnly }}
OverlayReadOnly={{join $v ":"}}
{{- end }}
PrivateUsersChown={{if .PrivateUsersChown}}on{{else}}off{{end}}

[Network]
Private={{if .Private}}on{{else}}off{{end}}
VirtualEthernet={{if .VirtualEthernet}}on{{else}}off{{end}}
{{- range $_, $v := .VirtualEthernetExtra }}
VirtualEthernetExtra={{$v}}
{{- end }}
Interface={{join .Parameters " "}}
MACVLAN={{join .MACVLAN " "}}
IPVLAN={{join .IPVLAN " "}}
Bridge={{.Bridge}}
Zone={{.Zone}}
{{- range $_, $v := .Port }}
Port={{$v}}
{{- end }}
`

var tmpl = template.Must(template.New("nspawn").Funcs(funcMaps).Parse(nspawnTemplate))
