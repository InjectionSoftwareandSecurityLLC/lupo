# Implants

Implants are critical part of using any C2, but are a unique part of using Lupo specifically. While Lupo does provide some sample implants to get folks started, there is purposefully _no such thing_ as a "lupo implant/payload/agent/etc...". This a modular API driven C2. This means the C2 provides an open and documented API that anything can integrate with as an implant as long as it speaks a supported protocol. The expectation is that those that want to use Lupo can easily integrate with the API in any language/framework/platform of their choice. Often times "implants" are one of the "crown jewels" of an offensive operation. These are programs that we want to avoid detection, bypass controls, and help us maintain access, meaning they can be critical to an operation. Many C2 tools come with their own implants to make usage easy, but unfortunately over time their implants begin to get flagged by defensive controls. 

Some try to get around this with obfuscation, UPX packing, new language implementations, and regular re-compilation with signature breaking changes. All of these techniques are great to use, but whenever those mechanisms are open sourced they can also be detected by defenders over time. Lupo seeks to solve this problem by doing all the heavy lifting on the C2 side, and all you have to do is implement the correct life cycle loop and API calls in your implant. Lupo keeps the required parameters limited to simple command execution and authentication to help quickly develop implants. That said it supports far more functionality should one choose to implement it. This extended functionality can enable file uploads/downloads and custom functions that implants can hook into the C2 server.

By keeping your implants private, you'll be able to bypass more defensive controls and maintain persistence more effectively. This is the biggest feature of Lupo allowing for powerful collaborative cross-platform C2.

Samples that fully implement the Lupo C2 API can be found in the [samples](../../samples) directory of the Lupo C2 repository. Currently the following are provided:
- HTTP/HTTPS Golang Based Implant
- Bind Connector PHP Based Web Shell

Feel free to use any provided sample implants but keep in mind these are meant to be examples for you to write your own implants with. They _will not_ be updated/maintained to evade protections or implement unique features. They _will_ be updated to implement new API features as they are made available, but this is purely to help continue to provide examples on how to integrate with those features.