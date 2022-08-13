package domain

import "net"

type ConBytesMsg struct {
	Conn    net.Conn
	MsgBody []byte
}

type ByteMsg struct {
	MsgBody []byte
}
