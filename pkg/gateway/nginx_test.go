package gateway

import (
	"strings"
	"testing"
)

type fakeServiceMapper struct {
	sm  ServiceMap
	err error
}

func (fsm *fakeServiceMapper) ServiceMap() (*ServiceMap, error) {
	return &fsm.sm, fsm.err
}

func TestRenderHTTPServiceGroups(t *testing.T) {
	fsm := fakeServiceMapper{
		sm: ServiceMap{
			HTTPServiceGroups: []*HTTPServiceGroup{
				&HTTPServiceGroup{
					name:      "ing1",
					namespace: "svc1",
					Aliases: []string{
						"apps.example.com",
					},
					Services: []HTTPService{
						HTTPService{
							Namespace: "ns1",
							Name:      "svc1",
							Endpoints: []TCPEndpoint{
								TCPEndpoint{Name: "pod1", IP: "10.0.0.1"},
								TCPEndpoint{Name: "pod2", IP: "10.0.0.2"},
								TCPEndpoint{Name: "pod3", IP: "10.0.0.3"},
							},
						},
						HTTPService{
							Namespace: "ns1",
							Name:      "svc2",
							Path:      "/v0",
							Endpoints: []TCPEndpoint{
								TCPEndpoint{Name: "pod1", IP: "10.0.0.4"},
								TCPEndpoint{Name: "pod2", IP: "10.0.0.5"},
								TCPEndpoint{Name: "pod3", IP: "10.0.0.6"},
							},
						},
					},
				},
			},
		},
	}

	cfg := DefaultNGINXConfig
	got, err := renderConfig(&cfg, &fsm.sm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `
pid /var/run/nginx.pid;
daemon on;

events {
    worker_connections 512;
}

http {
    server_names_hash_bucket_size 128;

    server {
        listen 7332;
        location /health {
            return 200 'Healthy!';
        }
    }

    server {
        listen 7331;
        return 444;
    }

    server {
        listen 7331;
        server_name ing1.svc1.example.com apps.example.com;

        location / {
            proxy_pass http://ns1__ing1__svc1;
        }
        location /v0 {
            proxy_pass http://ns1__ing1__svc2;
        }
    }

    upstream ns1__ing1__svc1 {

        server 10.0.0.1:0;  # pod1
        server 10.0.0.2:0;  # pod2
        server 10.0.0.3:0;  # pod3
    }
    upstream ns1__ing1__svc2 {

        server 10.0.0.4:0;  # pod1
        server 10.0.0.5:0;  # pod2
        server 10.0.0.6:0;  # pod3
    }

}

stream {

}
`
	if want != string(got) {
		t.Fatalf(
			"unexpected output: want=%sgot=%s",
			strings.Replace(want, " ", "÷", -1),
			strings.Replace(string(got), " ", "÷", -1),
		)
	}
}

func TestRenderTCPServices(t *testing.T) {
	fsm := fakeServiceMapper{
		sm: ServiceMap{
			TCPServices: []*TCPService{
				&TCPService{
					name:       "svc1",
					namespace:  "ns1",
					ListenPort: 19822,
					Endpoints: []TCPEndpoint{
						TCPEndpoint{Name: "pod1", IP: "10.0.0.1"},
						TCPEndpoint{Name: "pod2", IP: "10.0.0.2"},
						TCPEndpoint{Name: "pod3", IP: "10.0.0.3"},
					},
				},
				&TCPService{
					name:       "svc2",
					namespace:  "ns2",
					ListenPort: 19305,
					Endpoints: []TCPEndpoint{
						TCPEndpoint{Name: "pod1", IP: "10.0.0.4"},
						TCPEndpoint{Name: "pod2", IP: "10.0.0.5"},
						TCPEndpoint{Name: "pod3", IP: "10.0.0.6"},
					},
				},
			},
		},
	}

	cfg := DefaultNGINXConfig
	got, err := renderConfig(&cfg, &fsm.sm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `
pid /var/run/nginx.pid;
daemon on;

events {
    worker_connections 512;
}

http {
    server_names_hash_bucket_size 128;

    server {
        listen 7332;
        location /health {
            return 200 'Healthy!';
        }
    }

    server {
        listen 7331;
        return 444;
    }

}

stream {

    server {
        listen 19822;
        proxy_pass ns1__svc1;
    }

    upstream ns1__svc1 {

        server 10.0.0.1:0;  # pod1
        server 10.0.0.2:0;  # pod2
        server 10.0.0.3:0;  # pod3
    }

    server {
        listen 19305;
        proxy_pass ns2__svc2;
    }

    upstream ns2__svc2 {

        server 10.0.0.4:0;  # pod1
        server 10.0.0.5:0;  # pod2
        server 10.0.0.6:0;  # pod3
    }

}
`
	if want != string(got) {
		t.Fatalf(
			"unexpected output: want=%sgot=%s",
			strings.Replace(want, " ", "÷", -1),
			strings.Replace(string(got), " ", "÷", -1),
		)
	}
}
