//go:generate fs2-generic --output=wrapper.go --package-name=itemset_int bptree --key-type=*github.com/timtadh/sfp/stores/itemsets/ItemSet --key-serializer=github.com/timtadh/sfp/stores/itemsets/ItemSetSerialize --key-deserializer=github.com/timtadh/sfp/stores/itemsets/ItemSetDeserialize --value-type=int32 --value-size=4 --value-empty=0 --value-serializer=github.com/timtadh/sfp/stores/intint/SerializeInt32 --value-deserializer=github.com/timtadh/sfp/stores/intint/DeserializeInt32

package itemset_int
