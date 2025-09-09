package net_scan

import (
	"context"
	"os/exec"
)

func ping(ctx context.Context, address string) bool {
	cmd := exec.CommandContext(ctx, "ping", "-c", "1", "-W", "5", address)

	return cmd.Run() == nil
}
