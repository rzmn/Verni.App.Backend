package repositories

type MutationWorkItem struct {
	Perform  func() error
	Rollback func() error
}

type MutationWorkItemWithReturnValue[T any] struct {
	Perform  func() (T, error)
	Rollback func() error
}
