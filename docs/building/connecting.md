# Connecting

Once you have a Lupo server binary, starting it up is as simple as executing the binary. This will drop you into a Lupo CLI shell like so:
![lupo cli png](../assets/lupo_cli.png)


To connect to a wolfpack server instance via the Lupo client binary, follow these steps:
1. Start up a Lupo server
2. Start a new wolfpack server with `wolfpack start`
3. Setup a new wolfpack operator with `wolfpack register <username>`
4. Copy the config from the `wolfpack_configs` directory that gets generated.
5. Startup a Lupo client binary passing in the JSON config that was generated.