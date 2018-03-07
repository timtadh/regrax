//go:generate fs2-generic --output=wrapper.go --package-name=ints_int bptree --key-type=[]int32 --key-serializer=github.com/timtadh/regrax/stores/ints_ints/SerializeInt32s --key-deserializer=github.com/timtadh/regrax/stores/ints_ints/DeserializeInt32s --value-type=int32 --value-empty=0 --value-size=4 --value-serializer=github.com/timtadh/regrax/stores/int_int/SerializeInt32 --value-deserializer=github.com/timtadh/regrax/stores/int_int/DeserializeInt32
package ints_int
