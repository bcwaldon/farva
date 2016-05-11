package gateway

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Config struct {
	RefreshInterval time.Duration
	KubeconfigFile  string
	NGINXDryRun     bool
	NGINXHealthPort int
	FarvaHealthPort int
}

func New(cfg Config) (*Gateway, error) {
	kc, err := newKubernetesClient(cfg.KubeconfigFile)
	if err != nil {
		return nil, err
	}

	sm := newServiceMapper(kc)

	var nm NGINXManager
	if cfg.NGINXDryRun {
		nm = newLoggingNGINXManager(cfg.NGINXHealthPort)
	} else {
		nm = newNGINXManager(cfg.NGINXHealthPort)
	}

	gw := Gateway{
		cfg: cfg,
		sm:  sm,
		nm:  nm,
	}

	return &gw, nil
}

type Gateway struct {
	cfg Config
	sm  ServiceMapper
	nm  NGINXManager
}

func (gw *Gateway) start() error {
	ok, err := gw.nginxIsRunning()
	if err != nil {
		return err
	} else if ok {
		return nil
	}

	if err := gw.nm.WriteConfig(&ServiceMap{}); err != nil {
		return err
	}

	if err := gw.nm.Start(); err != nil {
		return err
	}

	gw.startHTTPServer()

	return nil
}

func (gw *Gateway) startHTTPServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Healthy!")
	})

	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", gw.cfg.FarvaHealthPort),
		Handler: mux,
	}

	go func() {
		log.Fatal(s.ListenAndServe())
	}()
}

func (gw *Gateway) nginxIsRunning() (bool, error) {
	log.Printf("Checking if nginx is running")
	st, err := gw.nm.Status()
	if err != nil {
		return false, err
	}
	return st == nginxStatusRunning, nil
}

func (gw *Gateway) refresh() error {
	log.Printf("Refreshing nginx config")
	sm, err := gw.sm.ServiceMap()
	if err != nil {
		return err
	}

	if err := gw.nm.WriteConfig(sm); err != nil {
		return err
	}
	if err := gw.nm.Reload(); err != nil {
		return err
	}

	return nil
}

func (gw *Gateway) Run() error {
	if err := gw.start(); err != nil {
		return err
	}

	log.Printf("Gateway started successfully, entering refresh loop")

	ticker := time.NewTicker(gw.cfg.RefreshInterval)

	for {
		if err := gw.refresh(); err != nil {
			log.Printf("Failed refreshing Gateway: %v", err)
		}

		//NOTE(bcwaldon): receive from the ticker at the
		// end of the loop to emulate do-while semantics.
		<-ticker.C
	}

	return nil
}
