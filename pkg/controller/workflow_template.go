package controller

// we just need lister for workflow templates
func (c *Controller) initWorkflowTemplateWatcher() {
	c.wtLister = c.kubeciInformerFactory.Engine().V1alpha1().WorkflowTemplates().Lister()
}
