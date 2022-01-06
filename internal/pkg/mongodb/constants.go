package mongodb

const (
	keyTable       = "key"
	logTable       = "log"
	operationTable = "operation"
	keyMapTable    = "map"
	keyMapDB       = "keyMap"
)

var indexData = []IndexData{
	newIndexData(keyTable, []string{"key", "manual"}, true),
	newIndexData(logTable, []string{"key"}, false),
	newIndexData(operationTable, []string{"operationID"}, true),
}

var keyMapIndexData = []IndexData{
	newIndexData(keyMapTable, []string{"key"}, true),
	newIndexData(keyMapTable, []string{"externalID"}, true),
}
