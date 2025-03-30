package net_stats

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	pingColAddress = "address"
	pingColDur     = "dur (ms)"
	pingColSucFail = "success/failed"
)

var (
	addressesToPing = []string{
		"google.com:80",
		"ya.ru:80",
		"mts.ru:80",
		"kodiki-hack.ru:8080",
	}
	pingCols = []string{pingColAddress, pingColDur, pingColSucFail}
)

func (r *NetStats) pingsView(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	var pingsWg sync.WaitGroup

	for _, address := range addressesToPing {
		r.model.pingsData.Set(pingColAddress, address, address)

		pingsWg.Add(1)
		go func() {
			defer pingsWg.Done()

			var (
				totalDur = time.Duration(int64(0))
				failed   uint64
				success  uint64
				avg      time.Duration
			)

			tillTimer := time.NewTimer(10 * time.Second)
			ticker := time.NewTicker(500 * time.Millisecond)
			for {
				select {
				case <-tillTimer.C:
					return
				case <-ticker.C:
					t0 := time.Now()
					conn, err := net.DialTimeout("tcp", address, time.Second*2)
					if err != nil {
						failed++
						r.model.pingsData.Set(pingColSucFail, address, fmt.Sprintf("%d/%d", success, failed))

						continue
					}
					dur := time.Since(t0)
					totalDur += dur
					success++
					avg = time.Duration(uint64(totalDur) / success)

					r.model.pingsData.Set(pingColDur, address, avg)
					r.model.pingsData.Set(pingColSucFail, address, fmt.Sprintf("%d/%d", success, failed))

					err = conn.Close()
					if err != nil {
						panic("failed to close connection")
					}
				}
			}
		}()
	}

	pingsWg.Wait()
}
