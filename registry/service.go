package registry

import "fmt"

type Service struct {
	Name string
	Url  string
}

func (s *Service) String() string {
	return fmt.Sprintf("Service<Name=%s, Url=%s>", s.Name, s.Url)
}
