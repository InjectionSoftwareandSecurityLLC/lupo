# lupo
Modular C2 server to tame your pack of wolves.

## Build Status:
- Main:

  ![main build workflow](https://github.com/InjectionSoftwareandSecurityLLC/lupo/actions/workflows/build_release.yml/badge.svg?branch=main)
- Current Release v0.1.1 (BETA):

  ![beta v0.1.1 build workflow](https://github.com/InjectionSoftwareandSecurityLLC/lupo/actions/workflows/build_release.yml/badge.svg?branch=dev/v0.1.1-beta-release)

<p align="center">
  <img width=400px src="docs/assets/lupo_logo.png" />
</p>


## Current Release
- [v0.1.1 (BETA)](https://github.com/InjectionSoftwareandSecurityLLC/lupo/releases/tag/v0.1.1-beta) - Beta release, see release notes for more details.

## Documentation
- [Usage Docs](./docs/README.md)
- [Source Code Docs](https://pkg.go.dev/github.com/InjectionSoftwareandSecurityLLC/lupo@v0.1.0)
- [Contributing](contributing.md)


v0.1.1 (BETA) Features:
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
- [x] implement automated builds with Travis CI

Known Bugs:
- [ ] Implement a cleaner mechanism for handling server shutdown if the user accidentally starts a second Wolfpack server so a dangling instance isn't left.


Version 1.0 TODO:
- [ ] Implement config file for lupo server to auto supply configs
- [ ] Implement optional encryption flag for TCP/UDP
- [ ] wolfpack chat
- [ ] config parser for server to improve automation capabilities
- [x] implement automated releases with Github actions


Road Map:
- [ ] Consider Implementing UDP listener (Would be cool to come back to this, it's not hard, just tricky for implants to integrate with cleanly. Needs a seamless standard/API)
- [ ] Consider Implementing Proxying (Cool for v2 should be easy with a go revproxy lib)
- [ ] Web interface for wolfpack server
