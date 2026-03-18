package main

import (
	"fmt"
	"net"
	"os"
)

type portCheck struct {
	Name string
	Port int
}

var checks = []portCheck{
	{Name: "app server", Port: 2222},
	{Name: "mailpit smtp", Port: 1025},
	{Name: "mailpit ui", Port: 8025},
	{Name: "caddy http", Port: 9080},
}

func main() {
	failed := false
	for _, check := range checks {
		if err := ensurePortFree(check.Port); err != nil {
			failed = true
			fmt.Fprintf(os.Stderr, "dev preflight failed: %s port %d is in use (%v)\n", check.Name, check.Port, err)
		}
	}

	if failed {
		os.Exit(1)
	}

	fmt.Println("dev preflight ok: required ports are free")
}

func ensurePortFree(port int) error {
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer func() { _ = ln.Close() }()
	return nil
}
