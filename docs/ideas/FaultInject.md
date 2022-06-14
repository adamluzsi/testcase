
type FaultInject struct {
	On  string
	Err error
}

func (FaultInject) Check() error {
}

