package gateway

type ServiceMapper interface {
	ServiceMap() (*ServiceMap, error)
}

type ServiceMap struct {
	Services []Service
}

type Service struct {
	Namespace  string
	Name       string
	ListenPort int
	TargetPort int
	Endpoints  []Endpoint
	// Optional for rule-based backends.
	Path string
	Host string
}

type Endpoint struct {
	Name string
	IP   string
	Port int
}
