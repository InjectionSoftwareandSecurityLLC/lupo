# lupo
Modular C2 server to tame your pack of wolves

<p align="center">
  <img src="docs/lupo_logo.png" />
</p>



## Documenation
- [Usage Docs](./docs/README.md)
- [Source Code Docs]() # ADD GODOC LINK

TODO:
- [x] Implement data response and check in status intervals
- [x] Implement registering custom functions
- [x] Consider creating a "color" library in core to handle custom colors across the entire application
- [x] Port finished HTTP server to HTTPs
- [x] Enhance custom functions
- [x] Implement TCP listener
- [ ] ~~Consider Implementing UDP listener~~ (Would be cool to come back to this, it's not hard, just tricky for implants to integrate with cleanly. Needs a seamless standard/API)
- [ ] ~~Consider Implementing Proxying~~ (Cool for v2 should be easy with a go revproxy lib)
- [x] Implement "wolfpack" teamserver with client binary generation
- [x] Implement extended functions like upload/download and any other seemingly "universal" switches
- [ ] Implement config file for lupo server to auto supply configs
- [x] Implement a webshell handler for bind webshells
- [ ] Implement optional encryption flag for TCP/UDP
- [x] Consider random PSK generation rather than a default base key
- [ ] Document "API" and consider pre-generating API documentation.
- [ ] Document core features: TLS generation, custom functions (part of API but notable), implant baseline implementation
- [x] Add Exec command to allow local shell interaction while in the Lupo CLI
- [x] Reformat the ASCII art so it is printed a bit more cleanly
- [ ] Create demo implants to show off all the feature/functionality - documentation too
- [ ] Repo art update and open source!