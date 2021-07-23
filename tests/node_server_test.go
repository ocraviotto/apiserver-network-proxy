package tests

import (
	"context"
	"net"
	"testing"
	"time"

	"sigs.k8s.io/apiserver-network-proxy/pkg/agent"
)

// Just to test that the echoserver istself is actually working :D
func TestEchoServerOnly(t *testing.T) {
	e := EchoServer{
		Protocol: "tcp",
		Address:  "0.0.0.0:9999",
	}
	e.Run()
	time.Sleep(100 * time.Second)
}

func TestNodeServerProxy(t *testing.T) {
	e := EchoServer{
		Protocol: "tcp",
		Address:  "0.0.0.0:9999",
	}
	e.Run()

	stopCh := make(chan struct{})
	defer close(stopCh)

	proxy, cleanup, err := runGRPCProxyServer()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	cs := runAgent(proxy.agent, stopCh)
	pf := &agent.PortForwarder{
		ClientSet:  cs,
		ListenHost: "127.0.0.1",
	}
	pf.Serve(context.TODO(), agent.PortMapping{
		LocalPort:  9998,
		RemoteHost: "localhost",
		RemotePort: 9999,
	})
	// Wait for agent to register on proxy server
	time.Sleep(time.Second)

	// run test client
	conn, err := net.Dial("tcp", "localhost:9998")
	if err != nil {
		t.Error(err)
	}

	msg := "1234567890123456789012345"
	n, err := conn.Write([]byte(msg))
	if err != nil {
		t.Error(err)
	}
	if n != len(msg) {
		t.Errorf("expect write %d; got %d", len(msg), n)
	}

	var data [10]byte

	n, err = conn.Read(data[:])
	if err != nil {
		t.Error(err)
	}
	if string(data[:n]) != msg[:10] {
		t.Errorf("expect %s; got %s", msg[:10], string(data[:n]))
	}

	n, err = conn.Read(data[:])
	if err != nil {
		t.Error(err)
	}
	if string(data[:n]) != msg[10:20] {
		t.Errorf("expect %s; got %s", msg[10:20], string(data[:n]))
	}

	msg2 := "1234567"
	n, err = conn.Write([]byte(msg2))
	if err != nil {
		t.Error(err)
	}
	if n != len(msg2) {
		t.Errorf("expect write %d; got %d", len(msg2), n)
	}

	n, err = conn.Read(data[:])
	if err != nil {
		t.Error(err)
	}
	if string(data[:n]) != msg[20:] {
		t.Errorf("expect %s; got %s", msg[20:], string(data[:n]))
	}

	n, err = conn.Read(data[:])
	if err != nil {
		t.Error(err)
	}
	if string(data[:n]) != msg2 {
		t.Errorf("expect %s; got %s", msg, string(data[:n]))
	}

	if err := conn.Close(); err != nil {
		t.Error(err)
	}
}
