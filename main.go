package main

import (
    "flag"
    "fmt"
    "net/http"
    "net/url"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"
)

const (
    colorReset  = "\033[0m"
    colorGreen  = "\033[32m"
    colorYellow = "\033[33m"
    colorCyan   = "\033[36m"
)

// defaultPorts mirrors the Python script's default port list
var defaultPorts = []int{
    21, 22, 23, 25, 53, 69, 80, 109, 110, 123, 137, 138, 139, 143, 156, 389, 443,
    546, 547, 995, 993, 2086, 2087, 2082, 2083, 3306, 8080, 8443, 10000,
}

func colorize(color, text string) string {
    return color + text + colorReset
}

// scanPort sends an HTTP request through the Squid proxy to http://target:port.
// A port is considered OPEN when Squid returns 200, 404, or 401.
// (Squid typically returns 503 when the backend port is closed.)
func scanPort(client *http.Client, target string, port int) bool {
    targetURL := fmt.Sprintf("http://%s:%d", target, port)
    resp, err := client.Get(targetURL)
    if err != nil {
        return false
    }
    defer resp.Body.Close()

    code := resp.StatusCode
    return code == 200 || code == 404 || code == 401
}

func main() {
    var (
        proxy     string
        target    string
        portsFlag string
        allPorts  bool
        workers   int
        timeout   int
    )

    flag.StringVar(&proxy, "proxy", "", "Define proxy address URL (http://x.x.x.x:3128)")
    flag.StringVar(&target, "target", "", "Define target IP behind proxy")
    flag.StringVar(&portsFlag, "ports", "", "[Optional] Define target ports behind proxy (comma-separated)")
    flag.BoolVar(&allPorts, "allports", false, "[Optional] Scan all 65535 TCP ports behind proxy")
    flag.IntVar(&workers, "workers", 50, "[Optional] Number of concurrent workers (default: 50)")
    flag.IntVar(&timeout, "timeout", 10, "[Optional] HTTP request timeout in seconds (default: 10)")

    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Spose - Squid Pivoting Open Port Scanner (Go edition)\n\n")
        fmt.Fprintf(os.Stderr, "usage: spose --proxy PROXY --target TARGET [--ports PORTS] [--allports]\n\n")
        fmt.Fprintf(os.Stderr, "options:\n")
        flag.PrintDefaults()
    }

    flag.Parse()

    if proxy == "" || target == "" {
        flag.Usage()
        os.Exit(1)
    }

    proxyURL, err := url.Parse(proxy)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Invalid proxy URL: %v\n", err)
        os.Exit(1)
    }

    var ports []int
    switch {
    case allPorts:
        ports = make([]int, 65535)
        for i := 0; i < 65535; i++ {
            ports[i] = i + 1
        }
        fmt.Println(colorize(colorYellow, "Scanning all 65,535 TCP ports"))
    case portsFlag != "":
        for _, p := range strings.Split(portsFlag, ",") {
            p = strings.TrimSpace(p)
            if p == "" {
                continue
            }
            port, err := strconv.Atoi(p)
            if err != nil {
                fmt.Fprintf(os.Stderr, "Invalid port: %s\n", p)
                os.Exit(1)
            }
            ports = append(ports, port)
        }
        fmt.Println(colorize(colorYellow, fmt.Sprintf("Scanning specified ports: %s", portsFlag)))
    default:
        ports = defaultPorts
        fmt.Println(colorize(colorYellow, "Scanning default common ports"))
    }

    fmt.Println(colorize(colorCyan, fmt.Sprintf("Using proxy address %s", proxy)))
    fmt.Printf("Scanning with %d concurrent workers (timeout: %ds per request)\n\n", workers, timeout)

    transport := &http.Transport{
        Proxy:             http.ProxyURL(proxyURL),
        DisableKeepAlives: true,
    }
    client := &http.Client{
        Transport: transport,
        Timeout:   time.Duration(timeout) * time.Second,
    }

    portCh := make(chan int, workers)
    var mu sync.Mutex
    var openPorts []int
    var wg sync.WaitGroup

    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for port := range portCh {
                if scanPort(client, target, port) {
                    mu.Lock()
                    openPorts = append(openPorts, port)
                    mu.Unlock()
                    fmt.Println(colorize(colorGreen, fmt.Sprintf("%s:%d seems OPEN", target, port)))
                }
            }
        }()
    }

    // Feed ports
    for _, p := range ports {
        portCh <- p
    }
    close(portCh)

    wg.Wait()

    fmt.Println()
    mu.Lock()
    defer mu.Unlock()
    if len(openPorts) > 0 {
        fmt.Println(colorize(colorGreen, fmt.Sprintf("Found %d open port(s):", len(openPorts))))
        for _, p := range openPorts {
            fmt.Printf("  %s:%d\n", target, p)
        }
    } else {
        fmt.Println(colorize(colorYellow, "No open ports detected."))
    }
}
