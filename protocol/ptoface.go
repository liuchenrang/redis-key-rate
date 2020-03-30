package protocol

import (
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
)

type PtoFace interface {
	tcpassembly.StreamFactory
	GetPort() string
	GetFilter() string
	GetFace() string
	Init()
	WrapperTcp(tcp *layers.TCP) *layers.TCP
}
