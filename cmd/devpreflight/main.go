package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type portCheck struct {
	Name string
	Port int
}

var checks = []portCheck{
	{Name: "app server", Port: 2222},
}

func main() {
	failed := false
	for _, check := range checks {
		if err := ensurePortFreeAndKillIfNeeded(check); err != nil {
			failed = true
			fmt.Fprintf(os.Stderr, "dev preflight failed: %s port %d could not be freed (%v)\n", check.Name, check.Port, err)
		}
	}

	if failed {
		os.Exit(1)
	}

	fmt.Println("dev preflight ok: required ports are free")
}

func ensurePortFreeAndKillIfNeeded(check portCheck) error {
	if err := ensurePortFree(check.Port); err == nil {
		return nil
	}

	pids, err := pidsUsingPort(check.Port)
	if err != nil {
		return err
	}
	if len(pids) == 0 {
		return fmt.Errorf("port is in use, but no process was found listening")
	}

	for _, pid := range pids {
		if err := signalPID(pid, syscall.SIGTERM); err != nil {
			return err
		}
	}

	time.Sleep(300 * time.Millisecond)
	if err := ensurePortFree(check.Port); err == nil {
		fmt.Printf("dev preflight: reclaimed %s port %d by stopping PID(s) %v\n", check.Name, check.Port, pids)
		return nil
	}

	for _, pid := range pids {
		if err := signalPID(pid, syscall.SIGKILL); err != nil {
			return err
		}
	}

	time.Sleep(200 * time.Millisecond)
	if err := ensurePortFree(check.Port); err != nil {
		return err
	}

	fmt.Printf("dev preflight: force-reclaimed %s port %d by killing PID(s) %v\n", check.Name, check.Port, pids)
	return nil
}

func pidsUsingPort(port int) ([]int, error) {
	cmd := exec.Command("lsof", "-nP", "-tiTCP:"+strconv.Itoa(port), "-sTCP:LISTEN")
	out, err := cmd.Output()
	if err != nil {
		var startErr *exec.Error
		if errors.As(err, &startErr) {
			return nil, fmt.Errorf("lsof is required to free ports automatically: %w", err)
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() == 1 {
				return nil, nil
			}
		}

		return nil, err
	}

	lines := strings.Fields(string(out))
	pids := make([]int, 0, len(lines))
	for _, line := range lines {
		pid, convErr := strconv.Atoi(strings.TrimSpace(line))
		if convErr != nil {
			return nil, fmt.Errorf("invalid pid from lsof output %q: %w", line, convErr)
		}
		pids = append(pids, pid)
	}

	return pids, nil
}

func signalPID(pid int, sig syscall.Signal) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}
	if err := proc.Signal(sig); err != nil {
		return fmt.Errorf("signal %s to pid %d: %w", sig.String(), pid, err)
	}
	return nil
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
