# PortScan

A fast and powerful multithreaded TCP port scanner written in **Go** ⚡. Clean design, colorful output, IPv4/IPv6 support, clipboard copy, and customizable settings.

<img width="689" height="548" alt="image" src="https://github.com/user-attachments/assets/a6806ed3-5248-4908-814b-1631912e6b83" />

---

## 🚀 Features

* ⚙️ Highly concurrent port scanning with adjustable worker count
* 🌐 Supports both IPv4 and IPv6 targets
* 🎨 Colorful and clean terminal UI
* 📋 Auto-copy open ports to clipboard (`xclip` required)
* 🔢 Custom port ranges and timeouts
* 💻 Domain and IP address resolution

---

## 📦 Installation

```bash
sudo apt update && sudo apt install -y figlet lolcat xclip golang golang-go upx
git clone https://github.com/OusH4x/PortScan
cd PortScan
go build -ldflags "-s -w" PortScan.go && upx PortScan
```

---

## ▶️ Usage

```bash
./PortScan <host> [options]
```

### Options

| Flag        | Description                          | Default   |
| ----------- | ------------------------------------ | --------- |
| `-p<range>` | Port range to scan (e.g. `-p1-1000`) | `1-65535` |
| `-w<num>`   | Number of concurrent workers         | `1000`    |
| `-t<ms>`    | Timeout per port in milliseconds     | `500`     |
| `-c`        | Copy open ports to clipboard         | —         |
| `-h`        | Show help                            | —         |

---

## 🧪 Example

```bash
./PortScan 192.168.1.1
./PortScan scanme.nmap.org -p20-1024 -w500 -t300 -c
```

---

## ✅ Requirements

* Go 1.18 or higher
* `figlet`, `lolcat`, and `xclip` for full visual and clipboard support (optional)

---

## 🧠 Author

**OusH4x**
