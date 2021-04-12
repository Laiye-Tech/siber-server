package libs

import (
	"encoding/json"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

type JsonWrapper struct {
	V interface{}
}

func (j JsonWrapper) String() string {
	d, err := json.Marshal(j.V)
	if err != nil {
		return err.Error()
	}
	return string(d)
}

var marshal = &jsonpb.Marshaler{}

type JsonPbWrapper struct {
	V proto.Message
}

func (j JsonPbWrapper) String() string {
	d, err := marshal.MarshalToString(j.V)
	if err != nil {
		return err.Error()
	}
	return d
}

type JsonPbWrapperWithMarshal struct {
	V proto.Message
	M jsonpb.Marshaler
}

func (j JsonPbWrapperWithMarshal) String() string {
	d, err := j.M.MarshalToString(j.V)
	if err != nil {
		return err.Error()
	}
	return d
}
