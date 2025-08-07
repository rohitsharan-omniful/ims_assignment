package serializer

import "github.com/vmihailenco/msgpack"

type MsgpackSerializer struct {
}

func NewMsgpackSerializer() *MsgpackSerializer {
	return &MsgpackSerializer{}
}

func (m *MsgpackSerializer) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (m *MsgpackSerializer) Unmarshal(b []byte, v interface{}) error {
	return msgpack.Unmarshal(b, v)
}
