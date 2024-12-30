package mongodb

const (
	keyTable       = "key"
	logTable       = "log"
	operationTable = "operation"
	keyMapTable    = "map"
	keyMapDB       = "keyMap"
	settingTable   = "setting"
)

var indexData = []IndexData{
	newIndexData(keyTable, []string{"key", "manual"}, true),
	newIndexData(keyTable, []string{"keyID"}, false), // expected to be true, but it needs to support old functionality
	newIndexData(keyTable, []string{"updated"}, false),
	newIndexData(logTable, []string{"key"}, false),
	newIndexData(logTable, []string{"keyID"}, false),
	newIndexData(logTable, []string{"date"}, false),
	newIndexData(logTable, []string{"requestID"}, false),
	newIndexData(operationTable, []string{"operationID"}, true),
}

var keyMapIndexData = []IndexData{
	newIndexData(keyMapTable, []string{"keyHash"}, true),
	newIndexData(keyMapTable, []string{"externalID"}, true),
}
