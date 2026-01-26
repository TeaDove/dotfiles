package net_system

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/teadove/teasutils/utils/time_utils"
)

const (
	pingColAddress = "address"
	pingColDur     = "dur (ms)"
	pingColSucFail = "success/failed"
)

var (
	addressesToPing = []string{ //nolint:gochecknoglobals // is ok
		"google.com:80",
		"ya.ru:80",
		"mts.ru:80",
		"gov.ge:80",
		"vultr.com:80",
	}
	pingCols = []string{pingColAddress, pingColDur, pingColSucFail} //nolint:gochecknoglobals // is ok
)

func (r *Service) pingsView(ctx context.Context) {
	var pingsWg sync.WaitGroup

	for _, address := range addressesToPing {
		r.model.pingsTableData.Locker().Lock()
		r.model.pingsTableData.Set(pingColAddress, address, color.New(color.FgCyan, color.Faint).Sprint(address))
		r.model.pingsTableData.Locker().Unlock()

		pingsWg.Go(func() {
			var (
				totalDur = time.Duration(int64(0))
				failed   uint64
				success  uint64
				avg      time.Duration
				dialer   = net.Dialer{Timeout: time.Second * 2}
			)

			tillTimer := time.NewTimer(1 * time.Minute)
			ticker := time.NewTicker(500 * time.Millisecond)

			for {
				select {
				case <-tillTimer.C:
					return
				case <-ticker.C:
					t0 := time.Now()

					conn, err := dialer.DialContext(ctx, "tcp", address)
					if err != nil {
						failed++

						r.model.pingsTableData.Locker().Lock()
						r.model.pingsTableData.Set(pingColSucFail, address, fmt.Sprintf("%d/%d", success, failed))
						r.model.pingsTableData.Locker().Unlock()

						continue
					}

					dur := time.Since(t0)
					totalDur += dur
					success++
					avg = time.Duration(uint64(totalDur) / success)

					r.model.pingsTableData.Locker().Lock()
					r.model.pingsTableData.Set(pingColDur, address, time_utils.RoundDuration(avg))
					r.model.pingsTableData.Set(pingColSucFail, address, fmt.Sprintf("%d/%d", success, failed))
					r.model.pingsTableData.Locker().Unlock()

					err = conn.Close()
					if err != nil {
						panic("failed to close connection")
					}
				}
			}
		})
	}

	pingsWg.Wait()
}
