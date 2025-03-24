package net_stats

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	"net"
	"strconv"
	"sync"
	"time"
)

var addressesToPing = [5]string{
	"google.com:80",
	"ya.ru:80",
	"mts.ru:80",
	"kodiki-hack.ru:8080",
	"inner.kodiki-hack.ru:80",
}

type PingStats struct {
	avg     time.Duration
	total   time.Duration
	success uint64
	failed  uint64
}

func (r *NetStats) flushPings(rows map[string]*PingStats) {
	tableRows := make([]table.Row, 0)
	for _, address := range addressesToPing {
		stats := rows[address]
		avg := elipsis
		if stats.avg != 0 {
			avg = strconv.FormatFloat(stats.avg.Seconds()*1000, 'f', 2, 64)
		}

		tableRows = append(tableRows,
			table.Row{
				address,
				avg,
				fmt.Sprintf("%d/%d", stats.success, stats.failed),
			})
	}
	r.model.pings.SetRows(tableRows)
}

func (r *NetStats) pingsView(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	var (
		rows    = make(map[string]*PingStats)
		pingsWg sync.WaitGroup
	)

	for _, address := range addressesToPing {
		rows[address] = &PingStats{}
		pingsWg.Add(1)
		go func() {
			defer pingsWg.Done()
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
						rows[address].failed += 1
						r.flushPings(rows)
						continue
					}
					dur := time.Since(t0)
					rows[address].success += 1
					rows[address].total += dur
					rows[address].avg = time.Duration(uint64(rows[address].total) / rows[address].success)
					r.flushPings(rows)

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
