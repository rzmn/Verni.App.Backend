package pathProvider

type Service interface {
	AbsolutePath(relative string) string
}
