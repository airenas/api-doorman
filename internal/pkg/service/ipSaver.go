package service

type ipSaver struct {
	saver IPManager
	limit float64
}

func (is *ipSaver) Save(ip string) error {
	return is.saver.CheckCreate(ip, is.limit)
}
