package agent

import (
	"bytes"
	"encoding/json"
	"io"
	"net"

	"github.com/amitbet/teleporter/common"
	"github.com/amitbet/teleporter/logger"
)

type TaskType uint8

const (
	TaskTypeSocks = iota
	TaskTypeUpdateConfig
	TaskTypeVpn
)

type TaskInfo struct {
	Type          TaskType
	TargetAddress string //final target address (intermediate steps decided by network configurations)
	TargetPort    string
}

func writeTaskInfo(conn io.Writer, tInfo *TaskInfo) error {
	jstr, err := json.Marshal(tInfo)
	if err != nil {
		logger.Error("writeTaskHeader: Problem in TaskInfo json marshaling: ", err)
		return err
	}
	err = common.WriteString(conn, string(jstr))
	if err != nil {
		logger.Error("writeTaskHeader: Problem in writing TaskInfo: ", err)
		return err
	}
	return nil
}

func readTaskInfo(conn io.Reader) (*TaskInfo, error) {
	taskInfoStr, err := common.ReadString(conn)
	if err != nil {
		logger.Error("Client connect, failed while reading client header: %s\n", err)
		return nil, err
	}

	logger.Debug("Received task info string: ", taskInfoStr)
	cconfig := TaskInfo{}
	err = json.Unmarshal([]byte(taskInfoStr), &cconfig)
	if err != nil {
		logger.Error("Error unmarshaling taskInfo: %s\n", err)
		return nil, err
	}
	return &cconfig, nil
}

type TunnelTask struct {
	net.Conn
	Header  *TaskInfo
	preSend *bytes.Buffer // any bytes that need to be sent to the other side before piping the connections together
}

// ReadTunnelTask reads the task details from the connection and returns a new TunnelTask object
func ReadTunnelTask(conn net.Conn) (*TunnelTask, error) {
	//myType := []byte{0}
	//_, err := conn.Read(myType)
	//_, err := io.ReadAtLeast(conn, myType, 1)
	header, err := readTaskInfo(conn)
	if err != nil {
		logger.Info("Error in reading task type: ", err)
		return nil, err
	}
	task := NewTunnelTask(conn, header)

	return task, nil
}

// NewTunnelTask creates a new TunnelTask
func NewTunnelTask(conn net.Conn, taskInfo *TaskInfo) *TunnelTask {
	t := TunnelTask{Conn: conn}
	t.Header = taskInfo
	//t.Type = taskType
	t.preSend = &bytes.Buffer{}
	return &t
}

// ReadPresend returns the bytes that are in the presend buffer and zeroes it
func (t *TunnelTask) ReadPresend() []byte {
	b := t.preSend.Bytes()
	t.preSend.Reset()
	return b
}

// PrefixTaskInfo the task info onto the head of the stream to be sent
func (t *TunnelTask) PrefixTaskInfo() error {
	b := &bytes.Buffer{}
	err := writeTaskInfo(b, t.Header)
	if err != nil {
		return err
	}
	_, err = t.PrefixSend(b.Bytes())
	return err
}

// PrefixSend adds bytes to be sent before the connections are welded
func (t *TunnelTask) PrefixSend(b []byte) (int, error) {
	newPreSend := &bytes.Buffer{}
	n, err := newPreSend.Write(b)
	if err != nil {
		return n, err
	}
	_, err = newPreSend.Write(t.preSend.Bytes())

	t.preSend = newPreSend
	return n, err
}

// Read overrides the connection read to send the prefixed content first
func (t *TunnelTask) Read(b []byte) (int, error) {
	var n, n1 int
	var err error
	if t.preSend.Len() > 0 {
		n, err = t.preSend.Read(b)
		//logger.Debugf("Read: got from presend: %v len = %d newPresendLen = %d ", b[:n], n, t.preSend.Len())
		if err != nil {
			logger.Error("Read: returning (presend): err ", err)
			return n, err
		}
	}
	if n < len(b) {
		//n1, err := io.ReadAtLeast(t.Conn, b[n:], len(b)-n)
		n1, err = t.Conn.Read(b[n:])
		//logger.Debugf("Read: got from conn: %v len = %d newPresendLen = %d ", b[n:n+n1], n1, t.preSend.Len())
		if err != nil {
			logger.Error("Read: returning (from conn):", err)
			return n + n1, err
		}
	}

	//logger.Debugf("Read OK: returning: %v len = %d newPresendLen = %d ", b[:n], n, t.preSend.Len())
	return n + n1, nil
}

func (t *TunnelTask) Write(b []byte) (int, error) {
	//logger.Debugf("TunnelTask: writing: %v len = %d ", b[:len(b)], len(b))
	return t.Conn.Write(b)
}
