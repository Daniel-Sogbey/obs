package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/Daniel-Sogbey/obs/obs"
)

const defaultAddr = "http://localhost:7070/debug/obs"

// ANSI Colors
const (
	reset = "\033[0m"
	bold  = "\033[1m"

	neonBlue   = "\033[38;5;39m"
	neonGreen  = "\033[38;5;46m"
	neonPink   = "\033[38;5;213m"
	neonOrange = "\033[38;5;208m"
	neonPurple = "\033[38;5;141m"
	gray       = "\033[38;5;245m"
	red        = "\033[31m"
	yellow     = "\033[33m"
)

func colorState(s obs.State) string {
	switch s {
	case obs.StateRunning:
		return neonGreen + "● RUNNING" + reset
	case obs.StateIdle:
		return yellow + "● IDLE" + reset
	case obs.StateCompleted:
		return gray + "● COMPLETED" + reset
	default:
		return red + "● UNKNOWN" + reset
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	return d.String()
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	switch os.Args[1] {

	case "tree":
		treeCmd := flag.NewFlagSet("tree", flag.ExitOnError)
		addr := treeCmd.String("addr", defaultAddr, "observability endpoint address")
		watch := treeCmd.Bool("watch", false, "live update view")
		interval := treeCmd.Duration("interval", 1*time.Second, "refresh interval")
		_ = treeCmd.Parse(os.Args[2:])

		log.Println("INTERVAL-- ", interval)

		handleError(runTree(*addr, *watch, *interval))

	case "list":
		listCmd := flag.NewFlagSet("list", flag.ExitOnError)
		addr := listCmd.String("addr", defaultAddr, "observability endpoint address")
		_ = listCmd.Parse(os.Args[2:])
		handleError(runList(*addr))

	case "slow":
		slowCmd := flag.NewFlagSet("slow", flag.ExitOnError)
		addr := slowCmd.String("addr", defaultAddr, "observability endpoint address")
		threshold := slowCmd.Duration("threshold", 2*time.Second, "duration threshold (e.g. 2s, 500ms)")
		_ = slowCmd.Parse(os.Args[2:])
		handleError(runSlow(*addr, *threshold))

	case "leaks":
		leaksCmd := flag.NewFlagSet("leaks", flag.ExitOnError)
		addr := leaksCmd.String("addr", defaultAddr, "observability endpoint address")
		_ = leaksCmd.Parse(os.Args[2:])
		handleError(runLeaks(*addr))

	default:
		usage()
	}
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  obs tree [--addr=URL]")
	fmt.Println("  obs list [--addr=URL]")
	fmt.Println("  obs slow --threshold=2s [--addr=URL]")
	fmt.Println("  obs leaks [--addr=URL]")
}

func handleError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func fetchSnapshots(addr string) ([]obs.TrackerSnapshot, error) {
	resp, err := http.Get(addr)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var snapshots []obs.TrackerSnapshot

	if err := json.NewDecoder(resp.Body).Decode(&snapshots); err != nil {
		return nil, err
	}

	return snapshots, nil
}

func runTree(addr string, watch bool, interval time.Duration) error {

	for {
		snapshots, err := fetchSnapshots(addr)
		if err != nil {
			return err
		}

		roots := obs.BuildTree(snapshots)

		clearScreen()
		printHeader()

		fmt.Printf("%sActive Goroutines: %d%s\n\n",
			neonBlue,
			len(snapshots),
			reset,
		)

		printTree(roots, 0)

		if !watch {
			return nil
		}

		time.Sleep(interval)
	}
}

func clearScreen() {
	fmt.Print("\033c")
}

func runList(addr string) error {
	snapshots, err := fetchSnapshots(addr)
	if err != nil {
		return err
	}

	printHeader()
	fmt.Println(bold + neonPink + "\nGoroutine Snapshot\n" + reset)

	fmt.Printf("%-5s %-25s %-15s %-10s\n",
		"ID", "NAME", "STATE", "DURATION")

	for _, s := range snapshots {
		fmt.Printf("%-5d %-25s %-15s %-10s\n",
			s.Id,
			s.Name,
			colorState(s.State),
			formatDuration(s.Duration),
		)
	}

	return nil
}

func runSlow(addr string, threshold time.Duration) error {
	snapshots, err := fetchSnapshots(addr)
	if err != nil {
		return err
	}
	printHeader()
	fmt.Println(bold + neonOrange + "\nSlow Goroutines\n" + reset)

	for _, s := range snapshots {
		if s.Duration > threshold && s.State != obs.StateCompleted {
			fmt.Printf("%s %-25s %s %s\n",
				neonBlue,
				s.Name,
				colorState(s.State),
				formatDuration(s.Duration),
			)
		}
	}

	return nil
}

func runLeaks(addr string) error {
	snapshots, err := fetchSnapshots(addr)
	if err != nil {
		return err
	}

	const leakThreshold = 30 * time.Second

	printHeader()
	fmt.Println(bold + red + "\nPotential Leaks\n" + reset)

	for _, s := range snapshots {
		if s.State != obs.StateCompleted && s.Duration > leakThreshold {
			fmt.Printf("%s %-25s %s %s\n",
				red,
				s.Name,
				colorState(s.State),
				formatDuration(s.Duration),
			)
		}
	}

	return nil
}

func printTree(nodes []*obs.TreeNode, level int) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Snapshot.Duration > nodes[j].Snapshot.Duration
	})

	for _, node := range nodes {

		indent := ""
		for i := 0; i < level; i++ {
			indent += "  "
		}

		name := neonBlue + node.Snapshot.Name + reset
		state := colorState(node.Snapshot.State)
		duration := neonPurple + formatDuration(node.Snapshot.Duration) + reset

		fmt.Printf("%s%s %s %s\n",
			indent,
			name,
			state,
			duration,
		)

		printTree(node.Children, level+1)
	}
}

func printHeader() {
	fmt.Println(neonPurple + bold + "OBSERVABILITY" + reset)
	fmt.Println(gray + time.Now().Format(time.RFC1123) + reset)
	fmt.Println()
}
