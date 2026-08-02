package ustils

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func PrintProtoJSON(msg proto.Message) {
	marshaler := protojson.MarshalOptions{
		Multiline: true,
		Indent:    "  ",
	}
	out, err := marshaler.Marshal(msg)
	if err != nil {
		fmt.Printf("Failed to format response: %v\n", err)
		return
	}
	fmt.Println(string(out))
}