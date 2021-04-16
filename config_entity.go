package nrt_configs

import "encoding/json"

type ConfigValue struct {
	Name  string      `firestore:"name"`
	Value interface{} `firestore:"value"`
}

type configEntity struct {
	values map[string]interface{}
}

func (ce configEntity) Get(name string) interface{} {
	return ce.values[name]
}

func (ce configEntity) GetInt(name string) int {
	return ce.values[name].(int)
}

func (ce configEntity) GetInt64(name string) int64 {
	return ce.values[name].(int64)
}

func (ce configEntity) GetString(name string) string {
	return ce.values[name].(string)
}

func (ce configEntity) Exists(name string) bool {
	_, exists := ce.values[name]
	return exists
}

func (ce configEntity) Parse(name string, v interface{}) error {
	data, err := json.Marshal(ce.values[name])
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
