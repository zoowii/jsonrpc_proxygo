package registry

import "fmt"

type Service struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

func (s *Service) String() string {
	return fmt.Sprintf("Service<Name=%s, Url=%s>", s.Name, s.Url)
}
