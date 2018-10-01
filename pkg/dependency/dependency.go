package dependency

import (
	"fmt"

	"github.com/philopon/go-toposort"
	"kube.ci/engine/apis/engine/v1alpha1"
)

// TODO: check in validation webhook
// TODO: set default in using mutation webhook
func ResolveDependency(workflowSteps []v1alpha1.Step, preSteps, postSteps []v1alpha1.Step, order v1alpha1.ExecutionOrder) ([]v1alpha1.Task, error) {
	if order != v1alpha1.ExecutionOrderDAG {
		for _, step := range workflowSteps {
			if len(step.Dependency) != 0 {
				return nil, fmt.Errorf("dependencies are valid only when ExecutionOrder is dag")
			}
		}
	}

	var err error
	var layers [][]v1alpha1.Step

	switch order {
	case v1alpha1.ExecutionOrderDAG:
		stepsMap := make(map[string]v1alpha1.Step, 0)
		for _, step := range workflowSteps {
			stepsMap[step.Name] = step
		}
		if layers, err = dagToLayers(stepsMap); err != nil {
			return nil, err
		}
	case v1alpha1.ExecutionOrderParallel:
		layers = append(layers, workflowSteps) // add all steps in one layer
	default: // default is v1alpha1.ExecutionOrderTypeSerial
		for _, step := range workflowSteps {
			layers = append(layers, []v1alpha1.Step{step}) // add each step in new layer
		}
	}

	// pre and post steps in new layers
	if len(preSteps) > 0 {
		layers = append([][]v1alpha1.Step{preSteps}, layers...)
	}
	if len(postSteps) > 0 {
		layers = append(layers, postSteps)
	}

	return layersToTasks(layers), nil
}

func layersToTasks(layers [][]v1alpha1.Step) []v1alpha1.Task {
	var tasks []v1alpha1.Task
	taskIndex := 0

	newTask := func() {
		tasks = append(tasks, v1alpha1.Task{
			SerialSteps:   []v1alpha1.Step{},
			ParallelSteps: []v1alpha1.Step{},
		})
	}

	for ii, layer := range layers {
		if len(tasks) != taskIndex+1 {
			newTask()
		}

		// number of steps in a layer:
		// case-1: only one step ---> append into serialSteps
		// case-2: only one step but it is the last-layer ---> add as a parallelStep
		// case-3: multiple steps ---> add as parallelSteps and increment taskIndex

		if len(layer) == 1 && ii < len(layers)-1 {
			tasks[taskIndex].SerialSteps = append(tasks[taskIndex].SerialSteps, layer[0])
		} else {
			for i := range layer {
				tasks[taskIndex].ParallelSteps = append(tasks[taskIndex].ParallelSteps, layer[i])
			}
			taskIndex++
		}
	}

	return tasks
}

func dagToLayers(stepsMap map[string]v1alpha1.Step) ([][]v1alpha1.Step, error) {
	// topological sort
	graph := toposort.NewGraph(len(stepsMap))
	for _, step := range stepsMap {
		graph.AddNode(step.Name)
	}
	for _, step := range stepsMap {
		for _, parent := range step.Dependency {
			if ok := graph.AddEdge(parent, step.Name); !ok {
				return nil, fmt.Errorf("can't resolve dependency %s for step %s", parent, step.Name)
			}
		}
	}
	sortedNodes, ok := graph.Toposort()
	if !ok {
		return nil, fmt.Errorf("can't resolve dependency, reason: cycle detected")
	}

	// dag-to-layers
	levels := make(map[string]int, 0)
	maxLevel := 0

	for _, node := range sortedNodes {
		if levels[node] != 0 {
			return nil, fmt.Errorf("can't resolve dependency, reason: topsort corrupted")
		}

		maxParentLevel := 0
		for _, parent := range stepsMap[node].Dependency {
			if levels[parent] == 0 {
				return nil, fmt.Errorf("can't resolve dependency, reason: topsort corrupted")
			}
			if levels[parent] > maxParentLevel {
				maxParentLevel = levels[parent]
			}
		}

		levels[node] = maxParentLevel + 1
		if levels[node] > maxLevel {
			maxLevel = levels[node]
		}
	}

	layers := make([][]v1alpha1.Step, maxLevel)
	for node, level := range levels {
		layers[level-1] = append(layers[level-1], stepsMap[node])
	}

	return layers, nil
}
