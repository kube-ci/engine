package controller

import (
	"sync"

	"github.com/appscode/go/encoding/json/types"
	"github.com/appscode/kutil/meta"
	api "kube.ci/kubeci/apis/kubeci/v1alpha1"
)

type observedWorkflows struct {
	lock  sync.RWMutex
	items map[string]*types.IntHash
}

func (c *observedWorkflows) alreadyObserved(wf *api.Workflow) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.items[wf.Key()].Equal(types.NewIntHash(wf.Generation, meta.GenerationHash(wf)))
}

func (c *observedWorkflows) set(wf *api.Workflow) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.items[wf.Key()] = types.NewIntHash(wf.Generation, meta.GenerationHash(wf))
}

func (c *observedWorkflows) delete(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.items, key)
}
