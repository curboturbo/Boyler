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
	for i, elem:= range w.Steps{
		if err := elem.Commit(); err != nil{
			for j := i; j >= 0; j-- {
				w.Steps[j].Rollback()
				if i == j{
					break
				}
			}
			return fmt.Errorf("Failed on %d workflow step: %v", i, err)
		}else{
			continue
		}
	}
	return nil
}

