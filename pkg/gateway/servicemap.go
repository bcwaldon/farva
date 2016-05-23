package gateway

import (
	"fmt"
)

type ServiceMapper interface {
	ServiceMap() (*ServiceMap, error)
}

type ServiceMap struct {
	ServiceGroups []ServiceGroup
	AliasMap      *AliasMap
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

type AliasMap map[string]string

// Generates a list of aliases for a given ingress name and namespace.
func (a *AliasMap) FilterByIngress(name string, ns string) []string {
	results := make([]string, 0)
	for key, value := range *a {
		if value == fmt.Sprintf("%s.%s", name, ns) {
			results = append(results, key)
		}
	}
	return results
}
