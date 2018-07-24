package v1alpha1

import (
	core "k8s.io/api/core/v1"
)

func (wf *Workflow) ObjectReference() *core.ObjectReference {
	return &core.ObjectReference{
		APIVersion:      SchemeGroupVersion.String(),
		Kind:            ResourceKindWorkflow,
		Namespace:       wf.Namespace,
		Name:            wf.Name,
		UID:             wf.UID,
		ResourceVersion: wf.ResourceVersion,
	}
}

func (wp *Workplan) ObjectReference() *core.ObjectReference {
	return &core.ObjectReference{
		APIVersion:      SchemeGroupVersion.String(),
		Kind:            ResourceKindWorkplan,
		Namespace:       wp.Namespace,
		Name:            wp.Name,
		UID:             wp.UID,
		ResourceVersion: wp.ResourceVersion,
	}
}
