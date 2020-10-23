package mongodb

const (
	store        = "doorman"
	keyTable     = "key"
	requestTable = "request"
)

var indexData = []IndexData{
	newIndexData(keyTable, "key", true)}
