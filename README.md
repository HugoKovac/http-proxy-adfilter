# Router AdFilter in Go

## Project Overview

The goal of this project is to create a proxy that can:
  - Proxy HTTP traffic
  - (Optionally) Proxy HTTPS and WebSocket traffic
  - Persistently track clients by MAC address to enforce the following rules:
    - Subscribe clients to categorized block lists
    - Block HTTP(S) and WebSocket traffic to specified domains
    - (Temporarily disable all rules for a client, e.g., for a specific time window)
    - (Implement admin authentication?)
    - (Block traffic to associated IP addresses?)

## File Structure

- **`cmd`**: Contains main entry points
- **`internal/data`**: Handles data manipulation for the database and API
- **`internal/db`**: Database setup and configuration
- **`internal/pkg`**: Core business logic
- **`internal/types`**: Definitions of structs and interfaces
- **`tests`**: Unit and integration tests

## Storage Considerations

Choosing the optimal storage mechanism for the block lists depends on the following constraints:
- Limited RAM, CPU, and file I/O resources
- Persistent storage needs

The proxy needs to quickly check the domain state for each request, so fast retrieval is crucial. Given these constraints, **SQLite** appears to be the most suitable choice for its balance between performance and resource efficiency.

Since SQLite stores data in a file, batch inserts will be preferred for bulk operations, such as importing large block lists.

## HTTP Traffic Proxying
### GL-iNet Example

To redirect all HTTP traffic from a specific client (e.g., with IP `192.168.10.207`) to the proxy:

```bash
iptables -t nat -A PREROUTING -s 192.168.10.207 -p tcp --dport 80 -j REDIRECT --to-port 8888
```

[proxy.sh](https://github.com/elazarl/goproxy/blob/master/examples/goproxy-transparent/proxy.sh)

## Performance Comparison

| Name       | Language | Binary Size (kB) | Shared Lib (kB) | Memory Usage (Idle) (kB) | Memory Usage (Under Load) (kB) |
|------------|----------|------------------|-----------------|--------------------------|--------------------------------|
| **Eyeo SQLITE**   | Go       | 13365            | 4               | 168376                   | 179908                         |
| **Eyeo BOLT**   | Go       | 13285            | 4               | 153656                   | 170188                         |
| **tinyproxy** | C        | 130.8            | 652             | 39376                    | 40128                          |

### HTTP Request Time Comparison

| Proxy      | Normal Request Time | Proxied Request Time | Difference  |
|------------|---------------------|----------------------|-------------|
| **Eyeo SQLITE**   | 0.112908s            | 0.211491s            |  +0.097079s |
| **Eyeo BOLT**   | 0.112908s            | 0.114412s            | +0.001504s  |
| **tinyproxy** | 0.112908s            | 0.235008s            | +0.1221s  |

## Next Steps for Production

- Implement domain caching for faster lookups
- Set up SSL certificates for HTTPS support
- Add support for HTTPS and WebSocket proxying
- Implement IP blocking for domains

## TLS implementation

To implement TLS for downstream and upstream we can use an existing proxy.

Either:
### [gomitmproxy](https://github.com/AdguardTeam/gomitmproxy) by AdGuard

#### GPL-3.0 license

Our project has to be open source and GPL-3.0

#### Binary Size

Without the filtering part: 8.7M


### [goproxy](https://github.com/elazarl/goproxy)

#### BSD-3-Clause license

MIT project + The original project and its contributors cannot be used for commercial or advertising purposes.

#### Binary Size

Without the filtering part: 8.1M

---

The binaries size seems fair, considering that we should share the same libraries (to be confirmed).

We could also build the TLS part since we already have the HTTP part. But that would mean more testings to do on the TLS part, but also on the HTTP one.

