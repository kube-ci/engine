package v1alpha1

func (r Workflow) IsValid() error {
	return nil
}

func (r *Workflow) SetDefaults() (*Workflow, error) {
	if r.Spec.ExecutionOrder == "" {
		r.Spec.ExecutionOrder = ExecutionOrderSerial
	}
	return r, nil
}
