package workflow

import "fmt"


type ApplyTransaction func() error
type CancelTransaction func() error

type Workflow struct{
	Steps []SagaStep
}

type SagaStep struct{
	Commit ApplyTransaction
	Rollback CancelTransaction
}

func NewSagaStep(commit ApplyTransaction, rollback CancelTransaction) SagaStep {
    return SagaStep{
        Commit: commit,
    	Rollback: rollback,
    }
}


func NewWorkflow() *Workflow {
	return &Workflow{
		Steps: make([]SagaStep, 0),
	}
}


func (w *Workflow) Add(saga SagaStep) {
	w.Steps = append(w.Steps, saga)
}


func (w *Workflow) Execute() error {
	for i, step := range w.Steps {
		if err := step.Commit(); err != nil {
			for j := i - 1; j >= 0; j-- {
				if rbErr := w.Steps[j].Rollback(); rbErr != nil {
					fmt.Printf("rollback step %d failed: %v\n", j, rbErr)
				}
			}
			return fmt.Errorf("workflow failed at step %d: %w", i, err)
		}
	}
	return nil
}
