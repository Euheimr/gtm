# gtm

_golang task manager - text-based UI (TUI) task manager and resource monitor written entirely in [Go](https://go.dev/)._

---

## Table of Contents

1. [About](#about)
   1. [Why](#why)
   2. [Requirements](#requirements)
2. [Development](#development)
   1. [Building](#building)
      1. [Prerequisites](#prerequisites)
      2. [Build steps](#build)
   2. [TODO](#TODO)
   3. [Reference](#reference)
      1. [General](#general)
      2. [GPU](#gpu)
      3. [Example Projects](#example-projects)
3. [Contributing](#contributing)
---

## About

### Why

- Similar existing tools use `Python` (requires python interpreter) or `C` (difficult to work with, when it matters)
- Similar existing tools are _platform-specific_ (operating system and/or CPU architecture-restrictive) and compiled or written needing external dependencies
- `htop`, `btop` and `top` are great but don't have everything I need (also platform-specific)

**Most importantly**: I love writing [Go](https://go.dev/) & this project has **absolutely** made me a better programmer.

_Also, I totally haven't redesigned this whole project 5 times... definitely not... (... I did)_


### Requirements

1. I want a task manager and resource monitor with alerts
2. Must be usable from a terminal via SSH
3. Must be **FULLY** cross-platform (Linux, Windows & macOS)
4. Operate as expected in `BASH`, `zsh`, `wt` (_new [microsoft windows terminal](https://github.com/microsoft/terminal) packaged with Win11_) & `Powershell`
5. (_preferably_) Entirely written in "pure" [Go](https://go.dev/) ([without using cgo](https://dave.cheney.net/2016/01/18/cgo-is-not-go))

**_Please note_**: I do not plan on supporting `cmd` aka `conhost.exe`. Please use `Powershell` OR the replacement of `cmd`: [Microsoft Windows Terminal](https://github.com/microsoft/terminal)

`cmd` does **not support unicode** & works completely differently from every other console out there. [Windows Terminal](https://github.com/microsoft/terminal) exists as the replacement of `cmd` in Win10+ and `Powershell` works great too.

--- 

## Development

<pre>
Source Code Structure

gtm/
 ├─── cmd/
 │     └── main.go
 ├─── scripts/
 │     ├── run.sh
 │     ├── log.sh
 │     └── pprof.sh
 ├─── ui/
 │     ├── common.go
 │     ├── cpu.go
 │     ├── disk.go
 │     ├── gpu.go
 │     ├── mem.go
 │     ├── net.go
 │     └── proc.go
 ├─── config.go
 ├─── devices.go
 └─── log.go

</pre>

This project uses BASH/zsh shell scripts (within `scripts/`) to run & build the app:
  - For `Linux` or `macOS`, you can just use your standard shell
  - On `Windows`, you can use [Cygwin](https://cygwin.com/) to get those GNU tools (`sh`, `ls`, `tail`, `tree`, ...)

There is a run & build script `run.sh` in the root directory of this project and `log.sh` uses `tail` to track the latest log file entries in your terminal.

`cgo` is disabled (`CGO_ENABLED=0`) in the `run.sh` script. It is not needed for this project, but is a design requirement (as noted above in [Requirements](#Requirements)) for [specific reasons](https://dave.cheney.net/2016/01/18/cgo-is-not-go).
This is to ensure I don't accidentally manage to use `cgo` and then run into weird, complex issues down the road.

### Building

#### Prerequisites:
1. [Go 1.21+](https://go.dev/)
2. **Windows Only**: requires [Cygwin](https://cygwin.com/) to run the `run.sh` script. (the script uses cleans up old binaries, etc ... might make a `Powershell` script later to get around this)

Changing directory to this project, then executing `sh ./run.sh` first finds the `/bin` folder in the root directory of this project, and if a binary exists.
If a binary does exist, the script just runs that existing binary.
Otherwise, it compiles a new binary for your system and then runs it for you.

#### Build Steps:
1. `git clone https://github.com/euheimr/gtm`
2. Open up a shell:
   1. `Powershell` or `wt` (Windows)
   2. `BASH` or `zsh` (Linux/macOS)
3. `cd <PROJECT DIRECTORY>` (ie. `cd ~/Downloads/gtm`)
4. `sh ./scripts/run.sh`

You may force a build & run the executable _even if a binary exists_ by running:

  `sh ./scripts/run.sh build` or 
  `sh ./scripts/run.sh -b`

You may also force a build and NOT run the executable by running:

  `sh ./scripts/run.sh build-only` or
  `sh ./scripts/run.sh -bo`


### TODO

- CPU - `gopsutil`
  - [ ] Device data
  - UI
    - [ ] bars
    - [ ] graphs
    - [ ] alerts (over-temp)
- Disk - `gopsutil`
  - [ ] Device data
  - UI
    - [ ] bars
    - [ ] graphs
- GPU - `nvidia-smi` & `rocm-smi`
  - [ ] Device data
  - UI
    - [ ] bars
    - [ ] graphs
    - [ ] alerts (over-temp)
- Memory - `gopsutil`
  - [x] Device data
  - UI
    - [ ] bars
    - [ ] graphs
    - [ ] alerts (excessive paging / usage >85%)
- [ ] Networking - `gopsutil`
  - [ ] Device data (by interface)
  - UI
    - [ ] bars
    - [ ] graphs (braille like `htop` ?)
- [ ] Processes - `gopsutil`
  - [ ] Device data
  - UI
    - [ ] Table
    - [ ] Tree view
  - Control
    - [ ] Kill
    - [ ] Priority
    - [ ] Open file location

### Reference

#### General

 - [Awesome Go](https://awesome-go.com/) - curated list of awesome Go frameworks, libraries, and software
 - [Logging in Go with Slog](https://betterstack.com/community/guides/logging/logging-in-go/) - structured logging
 - [pprof - performance profiling](https://jvns.ca/blog/2017/09/24/profiling-go-with-pprof/) - profiling go code with pprof

#### GPU

 - [nvidia-smi](https://developer.nvidia.com/system-management-interface) - command line interface (CLI) utility for management and monitoring of NVIDIA GPU devices
 - [rocm-smi](https://rocm.docs.amd.com/projects/amdsmi/en/latest/how-to/using-AMD-SMI-CLI-tool.html) - CLI tool for telemetry / monitoring AMD devices

#### Example projects

 - [ZanMax/gpu-stats](https://github.com/ZanMax/gpu-stats/blob/3197b24cebfd/main.go) - a project using `nvidia-smi` & `rocm-smi`

## Contributing

**I'm not taking pull requests for now**, but will personally take on reported issues.

_I will change this in the future (sorry, not ready for this yet)._