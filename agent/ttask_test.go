package agent

import (
	"io"
	"net"
	"testing"
)

func TestTunnelTask(t *testing.T) {
	// this pipe simulates the socks5 connection
	server, client := net.Pipe()
	task := NewTunnelTask(client,
		&TaskInfo{
			Type:          TaskTypeSocks,
			TargetPort:    "8080",
			TargetAddress: "localhost"})
	task.PrefixTaskInfo()

	task.PrefixSend([]byte("abcd"))
	go server.Write([]byte("12345678901234567890"))

	// this pipe simulates the relay mux connection
	server2, client2 := net.Pipe()
	go io.Copy(server2, task)

	//read the task from the relay connection
	gotTask, err := ReadTunnelTask(client2)
	if err != nil {
		t.Fatalf("error while reading task %s", err)
	}
	if gotTask.Header.Type != TaskTypeSocks {
		t.Fatalf("bad task type, %d: should be %d ....\n", gotTask.Header.Type, TaskTypeSocks)
	}

	buff := make([]byte, 15)
	//_, err = task.Read(buff)
	_, err = io.ReadAtLeast(gotTask, buff, 15)
	if err != nil {
		t.Fatalf("error while reading type %s", err)
	}

	actual := string(buff)
	expected := "abcd12345678901"
	if actual != expected {
		t.Fatalf("bad output, got: %s, should be %s ....\n", actual, expected)
	}
}
