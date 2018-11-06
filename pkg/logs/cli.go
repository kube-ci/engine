package logs

import (
	"bufio"
	"fmt"

	"gopkg.in/AlecAivazis/survey.v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	api "kube.ci/engine/apis/engine/v1alpha1"
)

func (c *LogController) GetLogs(query Query) error {
	if err := c.prepare(&query); err != nil {
		return err
	}
	reader, err := c.getLogReader(query)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Println(scanner.Text()) // write log to stdout
	}
	return nil
}

// if Workplan name not provided, ask Workflow name, then ask to select a Workplan
// if Workplan name provided, we don't need Workflow
func (c *LogController) prepare(query *Query) error {
	if query.Namespace == "" {
		if err := c.askNamespace(query); err != nil {
			return err
		}
	}
	if query.Workplan == "" {
		if query.Workflow == "" {
			if err := c.askWorkflow(query); err != nil {
				return err
			}
		}
		if err := c.askWorkplan(query); err != nil {
			return err
		}
	}
	if query.Step == "" {
		return c.askStep(query)
	}
	return nil
}

func (c *LogController) askNamespace(query *Query) error {
	var namespaceNames []string
	namespaces, err := c.kubeClient.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, ns := range namespaces.Items {
		namespaceNames = append(namespaceNames, ns.Name)
	}
	if len(namespaceNames) == 0 {
		return fmt.Errorf("no namespace found")
	}

	qs := []*survey.Question{
		{
			Name: "Namespace",
			Prompt: &survey.Select{
				Message: "Choose a Namespace:",
				Options: namespaceNames,
			},
		},
	}
	return survey.Ask(qs, query)
}

func (c *LogController) askWorkflow(query *Query) error {
	// list all workflows in the given Namespace
	var workflowNames []string
	workflows, err := c.kubeciClient.EngineV1alpha1().Workflows(query.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, wf := range workflows.Items {
		workflowNames = append(workflowNames, wf.Name)
	}
	if len(workflowNames) == 0 {
		return fmt.Errorf("no workflow found")
	}

	qs := []*survey.Question{
		{
			Name: "Workflow",
			Prompt: &survey.Select{
				Message: "Choose a Workflow:",
				Options: workflowNames,
			},
		},
	}
	return survey.Ask(qs, query)
}

func (c *LogController) askWorkplan(query *Query) error {
	// list all workplans in the given Namespace for a specific Workflow
	var workplanNames []string
	workplans, err := c.kubeciClient.EngineV1alpha1().Workplans(query.Namespace).List(metav1.ListOptions{
		LabelSelector: labels.FormatLabels(map[string]string{"workflow": query.Workflow}),
	})
	if err != nil {
		return err
	}
	for _, wp := range workplans.Items {
		workplanNames = append(workplanNames, wp.Name)
	}
	if len(workplanNames) == 0 {
		return fmt.Errorf("no workplan found")
	}

	qs := []*survey.Question{
		{
			Name: "Workplan",
			Prompt: &survey.Select{
				Message: "Choose a Workplan:",
				Options: workplanNames,
			},
		},
	}
	return survey.Ask(qs, query)
}

func (c *LogController) askStep(query *Query) error {
	// list all steps in the given Workplan along with their status
	workplanStatus, err := c.getWorkplanStatus(*query)
	if err != nil {
		return err
	}

	// list all running and terminated steps, logs are available only for those steps
	var stepNames []string
	for _, stepEntries := range workplanStatus.StepTree {
		for _, stepEntry := range stepEntries {
			if stepEntry.Status == api.ContainerRunning || stepEntry.Status == api.ContainerTerminated {
				stepNames = append(stepNames, stepEntry.Name)
			}
		}
	}

	qs := []*survey.Question{
		{
			Name: "Step",
			Prompt: &survey.Select{
				Message: "Choose a running/terminated step:",
				Options: stepNames,
			},
		},
	}
	return survey.Ask(qs, query)
}
