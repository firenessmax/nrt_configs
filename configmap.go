package nrt_configs

import (
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	"fmt"
	"sync"
)

type eventHandler func(name string, value interface{})

type eventHandlerById func(id string, value *firestore.DocumentSnapshot)

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

func (c *configmap) RegisterHandlerById(handler eventHandlerById, ref *firestore.CollectionRef) error {
	if ref == nil {
		return errors.New("nil Reference")
	}
	ctx := context.Background()

	rookie := NewRTWorker(ctx, ref.Snapshots(ctx))
	rookie.OnChange(func(docs []*firestore.DocumentSnapshot) error {
		for _, data := range docs {
			if e := safeHandleById(handler, data.Ref.ID, data); e != nil {
				return e
			}
		}
		return nil
	})
	c.workers = append(c.workers, rookie)
	return nil
}

func (c *configmap) RegisterHandler(handler eventHandler, ref *firestore.CollectionRef) error {
	if ref == nil {
		return errors.New("nil Reference")
	}
	ctx := context.Background()

	rookie := NewRTWorker(ctx, ref.Snapshots(ctx))
	rookie.OnChange(func(docs []*firestore.DocumentSnapshot) error {
		for _, data := range docs {
			val := ConfigValue{}
			err := data.DataTo(&val)
			if err != nil {
				return err
			}
			if e := safeHandle(handler, val.Name, val.Value); e != nil {
				return e
			}
		}
		return nil
	})
	c.workers = append(c.workers, rookie)
	return nil
}

func (c *configmap) Register(name string, ref *firestore.CollectionRef) error {
	ctx := context.Background()
	return c.RegisterWithContext(ctx, name, ref)
}

func (c *configmap) RegisterWithCancel(name string, ref *firestore.CollectionRef) (context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return cancel, c.RegisterWithContext(ctx, name, ref)
}

func (c *configmap) RegisterWithContext(ctx context.Context, name string, ref *firestore.CollectionRef) error {
	if ref == nil {
		return errors.New("nil Reference")
	}
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

func (cm *configmap) GetAll(cnfg string, assigner func(name string, value interface{})) {
	for name, value := range cm.configs[cnfg].values {
		assigner(name, value)
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

func safeHandle(handler eventHandler, name string, value interface{}) (e error) {
	defer func() {
		if e := recover(); e != nil {
			e = fmt.Errorf("%s", e)
		}
	}()
	handler(name, value)
	return
}

func safeHandleById(handler eventHandlerById, id string, value *firestore.DocumentSnapshot) (e error) {
	defer func() {
		if e := recover(); e != nil {
			e = fmt.Errorf("%s", e)
		}
	}()
	handler(id, value)
	return
}
