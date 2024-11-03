package logging

type Service interface {
	Log(format string, v ...any)
	Fatalf(format string, v ...any)
}

func DefaultService() Service {
	return &defaultService{}
}

func TestService() Service {
	return &defaultService{}
}
