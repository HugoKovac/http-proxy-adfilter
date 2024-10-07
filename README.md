# Router AdFilter in GO

The purpose of this project create a proxy that can:
  - Proxy HTTP
  - (Proxy HTTPS)
  - (Proxy Web Socket)
  - Save persitent client (via MAC adresses) to apply following rules:
    - Subscribe a client to categorized block list(s)
    - Block HTTP(S) and WS traffic to custom domains
    - (Disabling all list for x amount of time for SELF)
    - (Add auth with admin?)
    - (Block associated IP addresses ?)

## File Structure

- **`cmd`**: For all mains
- **`internal/data`**: For DB and API data manipulation
- **`internal/db`**: Setting up DB
- **`internal/pkg`**: buisness logic
- **`internal/types`**: All structs + interfaces
- **`tests`**: All tests

## Test DB

 You can run `make db_up` to launch a psql local DB *`(localhost:5432)`* and pgAdmin *`(localhost:8888)`* to administrate it

 You can run `make seed` to seed the DB with dummy data

# Storage

To choose the storage type of the list we have to take in account the limited:

- RAM
- CPU
- File i/o
- Persitant storage

Also we'll have to check at the domain state for each of the requests so we want this to be retrived as fast as possible

Knowing all that SqlLite seems to be the best choice.

Since Sql Lite is storing the data in a file we'll prefer batch inserts for repeted inserts such as the lists imports.

# Proxy HTTP Traffic
## Glinet

Here is how to proxy all the http traffic of a client to our proxy

 `iptables -t nat -A PREROUTING -s 192.168.10.207 -p tcp --dport 80 -j REDIRECT --to-port 8888`

# Performance

Name|Language|Binary Size (kB)|shared lib (kB)|at rest memory data (kB)|stress test memory data (kB)
---|---|---|---|---|---
Eyeo SQLITE|Go|13365|4|168376|179908
Eyeo Bolt|Go|13285|4|153656|170188
tinyproxy|C|130.8|652|39376|40128

HTTP Request comparison

 Proxy|normal total request time|Proxied total request time|Difference
---|---|---|---
Eyeo|0.367237s|0.482646s|0.115409s
tinyproxy|0.366646s|0.419841s|0.053195

