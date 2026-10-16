# Spose - Squid Pivoting Open Port Scanner (Go Edition)
A fast, concurrent port scanner designed to detect open ports behind a Squid proxy. This is a complete Go rewrite of the original Python spose tool, optimized with goroutines to scan ports significantly faster.

# Installation
```
git clone https://github.com/YOUR_USERNAME/spose-go.git
cd spose-go
go build -o spose
```

# Usage
```
Spose - Squid Pivoting Open Port Scanner (Go edition)

usage: spose --proxy PROXY --target TARGET [--ports PORTS] [--allports]

options:
  -allports
        [Optional] Scan all 65535 TCP ports behind proxy
  -ports string
        [Optional] Define target ports behind proxy (comma-separated)
  -proxy string
        Define proxy address URL (http://x.x.x.x:3128)
  -target string
        Define target IP behind proxy
  -timeout int
        [Optional] HTTP request timeout in seconds (default: 10) (default 10)
  -workers int
        [Optional] Number of concurrent workers (default: 50) (default 50)
```

# Examples
1. Scan default common ports:
```
./spose --proxy http://10.10.10.10:3128 --target 127.0.0.1
```
2. Scan specific ports:
```
./spose --proxy http://10.10.10.10:3128 --target 127.0.0.1 --ports 22,80,443,8080
```
3. Scan all ports:
```
./spose --proxy http://10.10.10.10:3128 --target 127.0.0.1 --allports
```

# Disclaimer
This tool is intended for educational purposes and authorized penetration testing/CTFs only. Ensure you have explicit permission to scan the target network and proxy. Unauthorized scanning is illegal. The author takes no responsibility for misuse of this tool.

# Credits
Original Python tool: [aancw/spose](https://github.com/aancw/spose)
