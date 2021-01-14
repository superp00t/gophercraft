package commands

import (
	"runtime"

	"github.com/superp00t/gophercraft/realm"
)

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func cmdStats(s *realm.Session) {
	stats := s.WS.GetServerStats()
	s.SystemChat("|cff34dceb|r ~~ Server Statistics ~~")
	s.Warnf("System: %s %s %s", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	s.Warnf("CPUs: %d", runtime.NumCPU())
	s.Warnf("Server uptime %v", stats.Uptime)
	s.Warnf("Number of active goroutines: %d", stats.Goroutines)
	s.Warnf("Total bytes allocated by heap: %d MiB", bToMb(stats.TotalAllocated))
	s.Warnf("Current memory allocated for server: %d MiB", bToMb(stats.SystemMemory))
	s.Warnf("Current memory usage of server: %d MiB", bToMb(stats.Allocated))
	s.Warnf("Total GC cycles: %d", stats.NumGCCycles)
}
