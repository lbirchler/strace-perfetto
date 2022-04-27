package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type Strace struct {
	DefaultArgs []string
	UserArgs    []string
	Timeout     int64
}

func (s Strace) Run() {
	args := append(s.DefaultArgs, s.UserArgs...)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.Timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "strace", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("[!] Strace timeout reached: %s\n", err)
	}
}
