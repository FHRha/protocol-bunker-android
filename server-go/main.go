package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	defaultPort := envInt("PORT", 8080)
	defaultHost := os.Getenv("HOST")
	if defaultHost == "" {
		defaultHost = "0.0.0.0"
	}
	defaultIdentityMode := normalizeIdentityMode(os.Getenv("BUNKER_IDENTITY_MODE"))

	exeDir := "."
	if exePath, err := os.Executable(); err == nil {
		exeDir = filepath.Dir(exePath)
	}
	cwd, _ := os.Getwd()

	defaultAssets := resolvePath(
		absPath(filepath.Join(cwd, "assets")),
		absPath(filepath.Join(cwd, "..", "assets")),
	)
	defaultClient := resolvePath(
		absPath(filepath.Join(cwd, "client", "dist")),
		absPath(filepath.Join(cwd, "..", "client", "dist")),
	)
	defaultScenarios := resolvePath(
		absPath(filepath.Join(cwd, "scenarios")),
		absPath(filepath.Join(cwd, "..", "scenarios")),
	)
	defaultSpecials := resolvePath(
		absPath(filepath.Join(defaultScenarios, "classic", "SPECIAL_CONDITIONS.json")),
		absPath(filepath.Join(cwd, "scenarios", "classic", "SPECIAL_CONDITIONS.json")),
	)

	port := flag.Int("port", defaultPort, "listen port")
	host := flag.String("host", defaultHost, "listen host")
	assetsRoot := flag.String("assets-root", defaultAssets, "assets root directory")
	clientDist := flag.String("client-dist", defaultClient, "client dist directory")
	scenariosRoot := flag.String("scenarios-root", defaultScenarios, "scenarios root directory")
	specialsFile := flag.String("specials-file", defaultSpecials, "SPECIAL_CONDITIONS.json file path")
	enableDevScenarios := flag.Bool("enable-dev-scenarios", envFlag(os.Getenv("BUNKER_ENABLE_DEV_SCENARIOS")), "enable dev scenarios")
	identityMode := flag.String("identity-mode", defaultIdentityMode, "identity mode: prod|dev_tab")
	flag.Parse()

	cfg := config{
		Host:                *host,
		Port:                *port,
		AssetsRoot:          *assetsRoot,
		ClientDistRoot:      *clientDist,
		IdentityMode:        normalizeIdentityMode(*identityMode),
		EnableDevScenarios:  *enableDevScenarios,
		ScenariosSourceRoot: resolvePath(
			absPath(*scenariosRoot),
			absPath(filepath.Join(exeDir, "..", "scenarios")),
		),
		SpecialsFile: resolvePath(
			absPath(*specialsFile),
			absPath(filepath.Join(exeDir, "..", "scenarios", "classic", "SPECIAL_CONDITIONS.json")),
		),
	}

	srv, err := newServer(cfg)
	if err != nil {
		log.Fatalf("server init failed: %v", err)
	}

	listenAddr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Printf("__BUNKER_PORT__=%d", cfg.Port)
	log.Printf("Server listening on http://%s", listenAddr)
	log.Printf("Assets root: %s (decks: %d)", cfg.AssetsRoot, len(srv.assets.Decks))
	log.Printf("Client dist: %s", cfg.ClientDistRoot)
	log.Printf("Scenarios root: %s", cfg.ScenariosSourceRoot)
	log.Printf("Specials file: %s", cfg.SpecialsFile)
	log.Printf("Identity mode: %s", cfg.IdentityMode)
	log.Printf("Scenarios: %v", availableScenarios(cfg.EnableDevScenarios))
	log.Printf("Mode: LAN only (Wi-Fi/hotspot)")

	httpServer := &http.Server{
		Addr:    listenAddr,
		Handler: srv.routes(),
	}
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen failed: %v", err)
	}
}
