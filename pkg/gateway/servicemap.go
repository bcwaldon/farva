package gateway

import (
	"fmt"
	"strings"
)

type ServiceMapper interface {
	ServiceMap() (*ServiceMap, error)
}

type ServiceMap struct {
	ServiceGroups []ServiceGroup
	Aliases       *Aliases
}

type Service struct {
	Namespace  string
	Name       string
	TargetPort int
	Endpoints  []Endpoint
	Path       string
}

type ServiceGroup struct {
	Name      string
	Namespace string
	Aliases   *Aliases
	Services  []Service
}

func (svg *ServiceGroup) DefaultServerName(cz string) string {
	return fmt.Sprintf(
		"%s.%s.%s",
		svg.Name,
		svg.Namespace,
		cz,
	)
}

type Endpoint struct {
	Name string
	IP   string
	Port int
}

type Aliases struct {
	Data map[string]string
}

// Generates a list of aliases for a given ingress name and namespace.
func (a *Aliases) Collect(name string, ns string) []string {
	results := make([]string, 0)
	for key, value := range a.Data {
		if value == fmt.Sprintf("%s.%s", name, ns) {
			results = append(results, key)
		}
	}
	return results
}

// Generates a concatenated list of aliases for a given ingress name and namespace.
func (a *Aliases) AliasNames(name string, ns string) string {
	aliases := a.Collect(name, ns)
	return strings.Join(aliases, " ")
}
