package tests

import (
	"net"

	"k8s.io/klog/v2"
)

type EchoServer struct {
	Protocol string
	Address  string

	listen net.Listener
}

func (e *EchoServer) Run() error {
	ln, err := net.Listen(e.Protocol, e.Address)
	if err != nil {
		return err
	}
	e.listen = ln

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				klog.Info(err)
				break
			}
			go e.echo(conn)
		}
	}()

	return nil
}

func (e *EchoServer) echo(conn net.Conn) {
	var data [256]byte

	for {
		n, err := conn.Read(data[:])
		if err != nil {
			klog.Info(err)
			return
		}

		_, err = conn.Write(data[:n])
		if err != nil {
			klog.Info(err)
			return
		}
	}
}
