package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	portRange = 65535
	batchSize = 500
)

var timeout = 500 * time.Millisecond

var (
	green   = "\033[32m"
	red     = "\033[31m"
	blue    = "\033[34m"
	yellow  = "\033[33m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	reset   = "\033[0m"
	bold    = "\033[1m"
)

func main() {
	showBanner()

	if len(os.Args) < 2 || os.Args[1] == "-h" {
		showHelp()
		return
	}

	input := os.Args[1]
	ips, err := net.LookupIP(input)
	if err != nil || len(ips) == 0 {
		fmt.Printf("%s[-] Failed to resolve host: %s%s\n", red, input, reset)
		return
	}
	target := ips[0].String()

	startPort := 1
	endPort := portRange
	autoCopy := false
	workers := 1000

	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "-c":
			autoCopy = true
		case strings.HasPrefix(arg, "-p"):
			ports := strings.Split(arg[2:], "-")
			if len(ports) == 2 {
				start, _ := strconv.Atoi(ports[0])
				end, _ := strconv.Atoi(ports[1])
				if start > 0 && end <= portRange {
					startPort = start
					endPort = end
				}
			}
		case strings.HasPrefix(arg, "-w"):
			if w, err := strconv.Atoi(arg[2:]); err == nil && w > 0 {
				workers = w
			}
		case strings.HasPrefix(arg, "-t"):
			if t, err := strconv.Atoi(arg[2:]); err == nil && t > 0 {
				timeout = time.Duration(t) * time.Millisecond
			}
		}
	}

	fmt.Printf("%s[+] Target:%s %s (%s)\n", cyan, reset, input, target)
	fmt.Printf("%s[+] Port Range:%s %d-%d\n", cyan, reset, startPort, endPort)
	fmt.Printf("%s[+] Workers:%s %d\n", cyan, reset, workers)
	fmt.Printf("%s[+] Timeout:%s %dms\n\n", cyan, reset, timeout.Milliseconds())

	fmt.Printf("%s[!] Starting scan in 2 seconds...%s\n", yellow, reset)
	time.Sleep(2 * time.Second)

	startTime := time.Now()
	openPorts := scanPorts(target, startPort, endPort, workers)
	elapsed := time.Since(startTime)

	var copied bool
	if autoCopy && len(openPorts) > 0 {
		copied = copyToClipboard(openPorts)
	}
	printResults(input, openPorts, elapsed, copied)
}

func showBanner() {
	cmd := exec.Command("sh", "-c", "figlet PortScan | lolcat")
	cmd.Stdout = os.Stdout
	cmd.Run()
	fmt.Printf("%s%sMade by OusH4x%s\n\n", bold, magenta, reset)
}

func showHelp() {
	fmt.Printf("%sUsage:%s %s <IP|Host> [-p1-1000] [-w1000] [-t500] [-c] [-h]\n", green, reset, os.Args[0])
	fmt.Printf("%sOptions:%s\n", yellow, reset)
	fmt.Printf("  %s-p<start-end>%s  Port range (e.g., -p1-1000) (default: 1-65535)\n", blue, reset)
	fmt.Printf("  %s-w<workers>%s    Number of workers (default: 1000)\n", blue, reset)
	fmt.Printf("  %s-t<ms>%s         Timeout per port in milliseconds (default: 500)\n", blue, reset)
	fmt.Printf("  %s-c%s            Copy open ports to clipboard\n", blue, reset)
	fmt.Printf("  %s-h%s            Show this help\n\n", blue, reset)
}

func scanPorts(target string, start, end, workers int) []int {
	var openPorts []int
	var mutex sync.Mutex
	var wg sync.WaitGroup
	ports := make(chan int, batchSize)
	var done int32
	total := end - start + 1

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for port := range ports {
				if isPortOpen(target, port) {
					mutex.Lock()
					openPorts = append(openPorts, port)
					mutex.Unlock()
				}
				atomic.AddInt32(&done, 1)
			}
		}()
	}

	go func() {
		for port := start; port <= end; port++ {
			ports <- port
		}
		close(ports)
	}()

	go func() {
		for {
			current := atomic.LoadInt32(&done)
			percent := float64(current) / float64(total) * 100
			fmt.Printf("\033[2K\r%sScanning: %d/%d (%.1f%%)%s", blue, current, total, percent, reset)
			if int(current) >= total {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	wg.Wait()
	fmt.Printf("\033[2K\r")
	return openPorts
}

func isPortOpen(ip string, port int) bool {
	var network string
	if strings.Contains(ip, ":") {
		network = "tcp6"
		ip = "[" + ip + "]"
	} else {
		network = "tcp4"
	}

	addr := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout(network, addr, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	fmt.Printf("\033[2K\r%s[+] Port %d OPEN%s\n", green, port, reset)
	return true
}

func printResults(target string, openPorts []int, elapsed time.Duration, copied bool) {
	fmt.Printf("\n%s╔══════════════════════════════════════════════════════════════════════╗\n", magenta)
	fmt.Printf("║ %s%-66s%s ║\n", bold+green, "SCAN COMPLETED!", magenta)
	fmt.Printf("╠══════════════════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║ %sTarget:       %s%-54s%s ║\n", cyan, reset, target, magenta)
	fmt.Printf("║ %sOpen ports:   %s%-54d%s ║\n", cyan, reset, len(openPorts), magenta)
	fmt.Printf("║ %sTime elapsed: %s%-54v%s ║\n", cyan, reset, elapsed.Round(time.Second), magenta)

	if len(openPorts) > 0 {
		sort.Ints(openPorts)
		portsStr := strings.Trim(strings.Replace(fmt.Sprint(openPorts), " ", ",", -1), "[]")
		if len(portsStr) > 54 {
			portsStr = portsStr[:51] + "..."
		}
		fmt.Printf("║ %sPorts:        %s%-54s%s ║\n", cyan, reset, portsStr, magenta)
	}
	status := "❌"
	if copied {
		status = "✅"
	}
	fmt.Printf("║ %sCopied to clipboard: %s%-47s%s ║\n", cyan, reset, status, magenta)

	fmt.Printf("╚══════════════════════════════════════════════════════════════════════╝\n%s", reset)
}

func copyToClipboard(ports []int) bool {
	portsStr := strings.Trim(strings.Replace(fmt.Sprint(ports), " ", ",", -1), "[]")
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo -n '%s' | xclip -sel clip", portsStr))
	if err := cmd.Run(); err == nil {
		fmt.Printf("\n%s[+] Ports copied to clipboard%s\n", green, reset)
		return true
	} else {
		fmt.Printf("\n%s[-] Failed to copy (install xclip)%s\n", red, reset)
		return false
	}
}
