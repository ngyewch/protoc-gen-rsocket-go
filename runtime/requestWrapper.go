package runtime

import (
	"bufio"
	"encoding/binary"
	"io"
)

type RequestWrapper struct {
	Selector   uint64
	MethodName string
	Payload    []byte
}

func (reqWrapper *RequestWrapper) Marshal(w io.Writer) error {
	outputBytes := make([]byte, 0)
	outputBytes = binary.AppendUvarint(outputBytes, reqWrapper.Selector)
	methodNameBytes := []byte(reqWrapper.MethodName)
	outputBytes = binary.AppendUvarint(outputBytes, uint64(len(methodNameBytes)))
	_, err := w.Write(outputBytes)
	if err != nil {
		return err
	}
	_, err = w.Write(methodNameBytes)
	if err != nil {
		return err
	}
	_, err = w.Write(reqWrapper.Payload)
	if err != nil {
		return err
	}
	return nil
}

func (reqWrapper *RequestWrapper) Unmarshal(r io.Reader) error {
	br := bufio.NewReader(r)
	selector, err := binary.ReadUvarint(br)
	if err != nil {
		return err
	}
	methodNameLen, err := binary.ReadUvarint(br)
	if err != nil {
		return err
	}
	methodNameBytes := make([]byte, methodNameLen)
	_, err = br.Read(methodNameBytes)
	if err != nil {
		return err
	}
	methodName := string(methodNameBytes)
	reqBytes, err := io.ReadAll(br)
	if err != nil {
		return err
	}
	reqWrapper.Selector = selector
	reqWrapper.MethodName = methodName
	reqWrapper.Payload = reqBytes
	return nil
}
