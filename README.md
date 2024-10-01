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

# Limits

## Lists

### Importing
> API Fetch
>
? What's the max list size that we can fetch and store in RAM ?

### Parsing
> From JSON fetched to local storing (via struct)

? How fast can we 

We know that the i/o is very low. So we want to prioritize batch commit to persitant storage.

? If we are using an sqllite DB is it possible to import a raw json as a table or append it to a table ?

### Storing
> DB usage and file size

We want a:
- Persitant storage
- Low complexity
- Low ram consumption
- Low I/O consumption


### Reading
> Reading Complexity and Caching

Hash table are prefered because we want to check domains for each of the request and we are waiting the results for each of them
