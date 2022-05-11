package service

import (
	"encoding/binary"
	"encoding/json"
	"io"
)

type Package struct {
	Type       []byte
	TypeLength int16
	Data       []byte
	DataLength int16
}

func NewPackage(tp string, data any) (*Package, error) {
	pack := &Package{}
	pack.Type = []byte(tp)
	pack.TypeLength = int16(len(pack.Type))
	if data != nil {
		if str, ok := data.(string); ok {
			pack.Data = []byte(str)

		} else {
			buf, err := json.Marshal(data)
			if err != nil {
				return nil, err
			}
			pack.Data = buf
		}
		pack.DataLength = int16(len(pack.Data))
	}
	return pack, nil
}

func (pack *Package) Unmarshal(v any) error {
	return json.Unmarshal(pack.Data, v)
}

func (pack *Package) String() string {
	return string(pack.Data)
}

func (p *Package) Pack(writer io.Writer) error {
	if err := binary.Write(writer, binary.BigEndian, &p.TypeLength); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.BigEndian, &p.Type); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.BigEndian, &p.DataLength); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.BigEndian, &p.Data); err != nil {
		return err
	}
	return nil
}

func (p *Package) Unpack(reader io.Reader) error {
	if err := binary.Read(reader, binary.BigEndian, &p.TypeLength); err != nil {
		return err
	}
	p.Type = make([]byte, p.TypeLength)
	if err := binary.Read(reader, binary.BigEndian, &p.Type); err != nil {
		return err
	}
	if err := binary.Read(reader, binary.BigEndian, &p.DataLength); err != nil {
		return err
	}
	p.Data = make([]byte, p.DataLength)
	if err := binary.Read(reader, binary.BigEndian, &p.Data); err != nil {
		return err
	}
	return nil
}
