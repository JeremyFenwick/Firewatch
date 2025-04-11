package unusualdatabase

import "sync"

type weirdDatase struct {
	data  map[string]string
	mutex sync.RWMutex
}

func (db *weirdDatase) insert(key, value string) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	db.data[key] = value
}

func (db *weirdDatase) retrieve(key string) (bool, string) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	value, ok := db.data[key]
	if !ok {
		return false, ""
	}
	return true, value
}
