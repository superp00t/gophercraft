package worldserver

import (
	"runtime"
	"time"
)

type ServerStats struct {
	Allocated      uint64
	TotalAllocated uint64
	SystemMemory   uint64
	NumGCCycles    uint32
	Goroutines     int
	Uptime         time.Duration
}

func (ws *WorldServer) GetServerStats() *ServerStats {
	sstats := &ServerStats{}
	sstats.Goroutines = runtime.NumGoroutine()
	sstats.Uptime = time.Since(ws.StartTime)
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	sstats.Allocated = memStats.Alloc
	sstats.TotalAllocated = memStats.TotalAlloc
	sstats.SystemMemory = memStats.Sys
	sstats.NumGCCycles = memStats.NumGC
	return sstats
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func x_Stats(c *C) {
	stats := c.Session.WS.GetServerStats()
	c.Session.SystemChat("|cff34dceb|r ~~ Server Statistics ~~")
	c.Session.Warnf("System: %s %s %s", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	c.Session.Warnf("CPUs: %d", runtime.NumCPU())
	c.Session.Warnf("Server uptime %v", stats.Uptime)
	c.Session.Warnf("Number of active goroutines: %d", stats.Goroutines)
	c.Session.Warnf("Total bytes allocated by heap: %d MiB", bToMb(stats.TotalAllocated))
	c.Session.Warnf("Current memory allocated for server: %d MiB", bToMb(stats.SystemMemory))
	c.Session.Warnf("Current memory usage of server: %d MiB", bToMb(stats.Allocated))
	c.Session.Warnf("Total GC cycles: %d", stats.NumGCCycles)
}
