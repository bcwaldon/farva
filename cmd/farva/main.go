package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/bcwaldon/farva/pkg/gateway"
)

func main() {
	fs := flag.NewFlagSet("farva", flag.ExitOnError)

	var cfg gateway.Config
	fs.DurationVar(&cfg.RefreshInterval, "refresh-interval", 30*time.Second, "Attempt to build and reload a new nginx config at this interval")
	fs.StringVar(&cfg.KubeconfigFile, "kubeconfig", "", "Set this to provide an explicit path to a kubeconfig, otherwise the in-cluster config will be used.")
	fs.BoolVar(&cfg.NGINXDryRun, "nginx-dry-run", false, "Log nginx management commands rather than executing them.")

	fs.Parse(os.Args[1:])

	gw, err := gateway.New(cfg)
	if err != nil {
		log.Fatalf("Gateway construction failed: %v", err)
	}

	if err := gw.Run(); err != nil {
		log.Printf("Gateway operation failed: %v", err)
	}

	log.Printf("Gateway shutting down")
}
