package proto

import (
	"strings"

	"google.golang.org/protobuf/encoding/protowire"
)

type Accomulator struct {
	Payloads [][]byte
}

func (a *Accomulator) Append(v any) {
	if payload, ok := v.([]uint8); ok {
		a.Payloads = append(a.Payloads, payload)
	}
}

func (a *Accomulator) IsProtoPayload() bool {
	return len(a.Payloads) > 0
}

func (a *Accomulator) ToString(delim string) string {
	var res strings.Builder
	for i, v := range a.Payloads {
		res.Write(v)
		if i < len(a.Payloads)-1 {
			res.WriteString(delim)
		}
	}
	return res.String()
}

type Field struct {
	Tag Tag
	Val Val
}

type Tag struct {
	Num  int32
	Type protowire.Type
}

type Val struct {
	Payload interface{}
	Length  int
}

func ParseUnknown(b []byte, acc *Accomulator) []Field {
	fields := make([]Field, 0)
	for len(b) > 0 {
		n, t, fieldlen := protowire.ConsumeField(b)
		if fieldlen < 1 {
			return nil
		}
		field := Field{
			Tag: Tag{Num: int32(n), Type: t},
		}

		_, _, taglen := protowire.ConsumeTag(b[:fieldlen])
		if taglen < 1 {
			return nil
		}

		var (
			v    interface{}
			vlen int
		)
		switch t {
		case protowire.VarintType:
			v, vlen = protowire.ConsumeVarint(b[taglen:fieldlen])

		case protowire.Fixed64Type:
			v, vlen = protowire.ConsumeFixed64(b[taglen:fieldlen])

		case protowire.BytesType:
			v, vlen = protowire.ConsumeBytes(b[taglen:fieldlen])
			sub := ParseUnknown(v.([]byte), acc)
			if sub != nil {
				v = sub
			}

		case protowire.StartGroupType:
			v, vlen = protowire.ConsumeGroup(n, b[taglen:fieldlen])
			sub := ParseUnknown(v.([]byte), acc)
			if sub != nil {
				v = sub
			}

		case protowire.Fixed32Type:
			v, vlen = protowire.ConsumeFixed32(b[taglen:fieldlen])
		}

		if vlen < 1 {
			return nil
		}

		acc.Append(v)
		field.Val = Val{Payload: v, Length: vlen - taglen}

		fields = append(fields, field)
		b = b[fieldlen:]
	}
	return fields
}
