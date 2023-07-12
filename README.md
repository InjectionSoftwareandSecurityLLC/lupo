# lupo
Modular C2 server to tame your pack of wolves.

<p align="center">
  <img width=400px src="docs/assets/lupo_logo.png" />
</p>


## Current Release
- [v1.0.5.2](https://github.com/InjectionSoftwareandSecurityLLC/lupo/releases/tag/v1.0.5.2) - Version 1.0.5.2 Release!

## Documentation
- [Usage Docs](./docs/README.md)
- [Source Code Docs](https://pkg.go.dev/github.com/InjectionSoftwareandSecurityLLC/lupo)
- [Contributing](contributing.md)

v1.0.5.2 Features:
- [X] Added `updateinterval` subcommand to Sessions CLI to allow for dynamic updates to implant check-in delays 
- [X] Added persistence handler updates to all listening modules
- [x] Implement data response and check in status intervals
- [x] Implement registering custom functions
- [x] Consider creating a "color" library in core to handle custom colors across the entire application
- [x] Port finished HTTP server to HTTPs
- [x] Enhance custom functions
- [x] Implement TCP listener
- [x] Implement "wolfpack" teamserver with client binary generation
- [x] Implement extended functions like upload/download and any other seemingly "universal" switches
- [x] Implement a web shell handler for bind web shells
- [x] Consider random PSK generation rather than a default base key
- [x] Add Exec command to allow local shell interaction while in the Lupo CLI
- [x] Reformat the ASCII art so it is printed a bit more cleanly
- [x] Document API
- [x] Document core features
- [x] Create demo implants to show off all the feature/functionality
- [x] Repo art update and open source!
- [x] Implement config file for lupo server to auto supply configs (done via metasploit-style resource file for simpler automation)
- [x] Implement optional encryption flag for TCP
- [x] wolfpack chat

v1.0.5.2 Bug Fixes:
- [X] Fixed some bugs in errors and text for the UpdateInterval functionality


Road Map:
- [ ] Consider Implementing UDP listener (Would be cool to come back to this, it's not hard, just tricky for implants to integrate with cleanly. Needs a seamless standard/API)
- [ ] Consider Implementing Proxying (Cool for v2 should be easy with a go revproxy lib)
- [ ] Implement Stager support for things like multi staged implants and DLL injection
- [ ] Implement Procdump API to support process dumping (can use the stager API)
- [ ] Web interface for wolfpack server
- [ ] Implement Github Actions to get automated builds for future releases
