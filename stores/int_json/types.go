//go:generate fs2-generic --output=wrapper.go --package-name=int_json bptree --key-type=int32 --key-size=4 --key-empty=0 --key-serializer=github.com/timtadh/sfp/stores/int_int/SerializeInt32 --key-deserializer=github.com/timtadh/sfp/stores/int_int/DeserializeInt32 --value-type=map[string]interface{} --value-serializer=SerializeJson --value-deserializer=DeserializeJson
package int_json

import (
	"bytes"
	"encoding/json"
)

func SerializeJson(obj map[string]interface{}) []byte {
	data, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	return data
}

func DeserializeJson(data []byte) (obj map[string]interface{}) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&obj); err != nil {
		panic(err)
	}
	return obj
}
