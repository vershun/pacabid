package main

type meanRevision struct {
	budget       float64
	marketStream <-chan bar
	orders       chan<- order
}

func (mr *meanRevision) symbolsToWatch() []string {
	return []string{"CERN"}
}

func (mr *meanRevision) prepare(
	budget float64,
	marketStream <-chan bar,
	orders chan<- order,
) {
	mr.budget = budget
	mr.marketStream = marketStream
	mr.orders = orders
}

func (mr *meanRevision) run(quit chan struct{}) error {
	for {
		select {
		case <-quit:
			return nil
		case update := <-mr.marketStream:
			if err := mr.handleUpdate(&update); err != nil {
				return err
			}
		}
	}
}

func (mr *meanRevision) handleUpdate(update *bar) error {

	return nil
}
