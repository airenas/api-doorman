package mongodb

const (
	keyTable = "key"
	logTable = "log"
)

var indexData = []IndexData{
	newIndexData(keyTable, []string{"key", "manual"}, true),
	newIndexData(logTable, []string{"key"}, false)}
