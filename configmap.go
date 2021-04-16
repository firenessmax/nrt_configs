package nrt_configs

import (
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	"sync"
)

type configmap struct {
	rwmu    sync.RWMutex
	configs map[string]configEntity
	workers []*RTWorker
}

func NewConfigmap() *configmap {
	return &configmap{
		configs: map[string]configEntity{},
	}
}

func (c *configmap) Register(name string, ref *firestore.CollectionRef) error {
	if ref == nil {
		return errors.New("nil Reference")
	}
	//TODO: handle context
	ctx := context.Background()

	rookie := NewRTWorker(ctx, ref.Snapshots(ctx))
	rookie.OnChange(func(docs []*firestore.DocumentSnapshot) error {

		values := map[string]interface{}{}
		for _, data := range docs {
			val := ConfigValue{}
			err := data.DataTo(&val)
			if err != nil {
				return err
			}
			values[val.Name] = val.Value
		}
		c.rwmu.Lock()
		defer c.rwmu.Unlock()
		c.configs[name] = configEntity{values: values}
		return nil
	})
	c.workers = append(c.workers, rookie)
	return nil
}

func (cm *configmap) ListenAndWait() {
	wg := sync.WaitGroup{}
	for i := range cm.workers {
		go func(i int) {
			wg.Add(1)
			cm.workers[i].Listen()
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func (cm *configmap) Listen() {
	for _, w := range cm.workers {
		go w.Listen()
	}
}

func (cm *configmap) Get(cnfg, name string) interface{} {
	//thread safe
	cm.rwmu.RLock()
	defer cm.rwmu.RUnlock()

	if ent, ok := cm.configs[cnfg]; ok {
		return ent.Get(name)
	}
	return nil
}

func (cm *configmap) GetInt(cnfg, name string) int {
	//thread safe
	cm.rwmu.RLock()
	defer cm.rwmu.RUnlock()

	if ent, ok := cm.configs[cnfg]; ok {
		return ent.GetInt(name)
	}
	return 0
}

func (cm *configmap) GetInt64(cnfg, name string) int64 {
	//thread safe
	cm.rwmu.RLock()
	defer cm.rwmu.RUnlock()

	if ent, ok := cm.configs[cnfg]; ok {
		return ent.GetInt64(name)
	}
	return 0
}

func (cm *configmap) GetString(cnfg, name string) string {
	//thread safe
	cm.rwmu.RLock()
	defer cm.rwmu.RUnlock()

	if ent, ok := cm.configs[cnfg]; ok {
		return ent.GetString(name)
	}
	return ""
}

func (cm *configmap) GetStruct(cnfg, name string, out interface{}) error {
	//thread safe
	cm.rwmu.RLock()
	defer cm.rwmu.RUnlock()

	if ent, ok := cm.configs[cnfg]; ok {
		return ent.Parse(name, out)
	}
	return errors.New("not found")
}

func (cm *configmap) Exists(cnfg, name string) bool {
	//thread safe
	cm.rwmu.RLock()
	defer cm.rwmu.RUnlock()
	if ent, ok := cm.configs[cnfg]; ok {
		return ent.Exists(name)
	}
	return false
}
