# TalentLand Workshop

[![Software License](https://img.shields.io/badge/license-BSD3-red.svg)](LICENSE)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-v2.0-ff69b4.svg)](.github/CODE_OF_CONDUCT.md)

This repository provides a sample project to provide a "hands on" exercise to follow
along during the workshop at TalentLand.

To work through the session and/or run this locally you'll need a few tools installed
on your local machine:

- A working Go setup. (<https://go.dev/doc/install>)
- A way to run containers, usually Docker. (<https://docs.docker.com/engine/install/>)
- An IDE of your choice; we'll use VS Code. <https://code.visualstudio.com>
- Optionally, "delve" to use as a Debugger. (<https://github.com/go-delve/delve>)

## How To

1. First you'll need to build a "debuggable" version of the application in a container: `make debugger-build`

2. Then start the application through a debugger server instance: `make debugger-run`
