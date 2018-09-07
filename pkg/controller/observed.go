package controller

import (
	"sync"

	"github.com/appscode/go/encoding/json/types"
	"github.com/appscode/kutil/meta"
	api "kube.ci/kubeci/apis/kubeci/v1alpha1"
)

type observedWorkflows struct {
	lock  sync.RWMutex
	items map[string]workflowState
}

type workflowState struct {
	hash              *types.IntHash
	observedResources map[api.ObjectReference]*types.IntHash
}

func (c *observedWorkflows) alreadyObserved(wf *api.Workflow) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	hash := types.NewIntHash(wf.Generation, meta.GenerationHash(wf))
	return c.items[wf.Key()].hash.Equal(hash)
}

func (c *observedWorkflows) resourceAlreadyObserved(workflowKey string, resource api.TriggeredFor) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.items[workflowKey].observedResources == nil {
		return false
	}
	observed, ok := c.items[workflowKey].observedResources[resource.ObjectReference]
	return ok && observed.Equal(resource.ResourceGeneration)
}

func (c *observedWorkflows) set(wf *api.Workflow) {
	c.lock.Lock()
	defer c.lock.Unlock()

	state := c.items[wf.Key()]
	state.hash = types.NewIntHash(wf.Generation, meta.GenerationHash(wf))
	c.items[wf.Key()] = state
}

func (c *observedWorkflows) setObservedResource(workflowKey string, resource api.TriggeredFor) {
	c.lock.Lock()
	defer c.lock.Unlock()

	state := c.items[workflowKey]
	if state.observedResources == nil {
		state.observedResources = make(map[api.ObjectReference]*types.IntHash)
	}
	state.observedResources[resource.ObjectReference] = resource.ResourceGeneration
	c.items[workflowKey] = state
}

// set ObservedResource if not exists, required for initial sync
func (c *observedWorkflows) upsertObservedResource(workflowKey string, resource api.TriggeredFor) {
	c.lock.Lock()
	defer c.lock.Unlock()

	state := c.items[workflowKey]
	if state.observedResources == nil {
		state.observedResources = make(map[api.ObjectReference]*types.IntHash)
	}
	if _, ok := state.observedResources[resource.ObjectReference]; !ok {
		state.observedResources[resource.ObjectReference] = resource.ResourceGeneration
		c.items[workflowKey] = state
	}
}

func (c *observedWorkflows) delete(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.items, key)
}
