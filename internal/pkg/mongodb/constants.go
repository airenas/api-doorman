package mongodb

const (
	store    = "doorman"
	keyTable = "key"
	logTable = "log"
)

var indexData = []IndexData{
	newIndexData(keyTable, []string{"key", "manual"}, true),
	newIndexData(logTable, []string{"key"}, false)}
