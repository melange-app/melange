// Package go_melange is an autogenerated binder stub for package melange.
//   gobind -lang=go getmelange.com/mobile/melange
//
// File is generated by gobind. Do not edit.
package go_melange

import (
	"getmelange.com/mobile/melange"
	"golang.org/x/mobile/bind/seq"
)

func proxy_Run(out, in *seq.Buffer) {
	param_port := in.ReadInt()
	param_dataDir := in.ReadUTF16()
	param_version := in.ReadUTF16()
	param_platform := in.ReadUTF16()
	err := melange.Run(param_port, param_dataDir, param_version, param_platform)
	if err == nil {
		out.WriteUTF16("")
	} else {
		out.WriteUTF16(err.Error())
	}
}

func init() {
	seq.Register("melange", 1, proxy_Run)
}
