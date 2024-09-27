# Router AdFilter in GO

The purpose of this project create a proxy that can:
  - Proxy HTTP
  - Proxy HTTPS
  - Proxy Web Socket
  - Save persitent client (via MAC adresses) to apply following rules:
    - Subscribe a client to categorized block list(s)
    - Block HTTP(S) and WS traffic to custom domains
    - Disabling all list for x amount of time for SELF
    - (Add auth with admin?)

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