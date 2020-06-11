package registry

import "fmt"

type Service struct {
	Name string `json:"name"`
	Url  string `json:"url"`
	Host string `json:"host"`
}

func (s *Service) String() string {
	return fmt.Sprintf("Service<Name=%s, Url=%s, Host=%s>", s.Name, s.Url, s.Host)
}
