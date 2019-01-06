package common

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/amitbet/teleporter/logger"
)

func ReadUint32(r io.Reader) (uint32, error) {
	var myUint uint32
	if err := binary.Read(r, binary.BigEndian, &myUint); err != nil {
		return 0, err
	}

	return myUint, nil
}

func ReadUint8(r io.Reader) (uint8, error) {
	var myUint uint8
	if err := binary.Read(r, binary.BigEndian, &myUint); err != nil {
		return 0, err
	}

	return myUint, nil
}

func ReadBytes(r io.Reader, count int) ([]byte, error) {
	buff := make([]byte, count)

	lengthRead, err := io.ReadFull(r, buff)

	if lengthRead != count {
		logger.Errorf("util.ReadBytes unable to read bytes: lengthRead=%d, countExpected=%d", lengthRead, count)
		return nil, errors.New("util.ReadBytes unable to read bytes")
	}

	if err != nil {
		logger.Errorf("util.ReadBytes error while reading bytes: ", err)
		return nil, err
	}

	return buff, nil
}

func ReadString(r io.Reader) (string, error) {
	size, err := ReadUint32(r)
	if err != nil {
		logger.Errorf("util.ReadString error while reading string size: ", err)
		return "", err
	}
	bytes, err := ReadBytes(r, int(size))
	if err != nil {
		logger.Errorf("util.ReadString error while reading string: ", err)
		return "", err
	}
	return string(bytes), nil
}

func WriteString(w io.Writer, str string) error {
	length := uint32(len(str))
	err := binary.Write(w, binary.BigEndian, length)
	if err != nil {
		logger.Errorf("util.WriteString error while writeing string length: ", err)
		return err
	}
	written, err := w.Write([]byte(str))
	if err != nil {
		logger.Errorf("util.WriteString error while writeing string: ", err)
		return err
	}
	if written != int(length) {
		logger.Errorf("util.WriteString error while writeing string: ", err)
		return errors.New("util.WriteString error while writeing string: written size too small")
	}
	return nil
}

func ReadShortString(r io.Reader) (string, error) {
	size, err := ReadUint8(r)
	if err != nil {
		logger.Errorf("util.ReadShortString error while reading string size: ", err)
		return "", err
	}
	bytes, err := ReadBytes(r, int(size))
	if err != nil {
		logger.Errorf("util.ReadShortString error while reading string: ", err)
		return "", err
	}
	return string(bytes), nil
}

func WriteShortString(w io.Writer, str string) error {
	length := uint8(len(str))
	err := binary.Write(w, binary.BigEndian, length)
	if err != nil {
		logger.Errorf("util.WriteShortString error while writeing string length: ", err)
		return err
	}
	written, err := w.Write([]byte(str))
	if err != nil {
		logger.Errorf("util.WriteShortString error while writeing string: ", err)
		return err
	}
	if written != int(length) {
		logger.Errorf("util.WriteShortString error while writeing string: ", err)
		return errors.New("util.WriteShortString error while writeing string: written size too small")
	}
	return nil
}
