# kurz-url: A simple URL shortener, a la bitly/tinyURL
Kurz is a simple version of a URL shortener, written in Golang using `net/http`.

## Design
Kurz has two primary components: the backend redirection API, written in Go, and the redirection rules database.
Future improvements may include additional components for rate-limiting, caching, scaling, etc.

### Redirection API
The redirection API provides 3 fundamental operations:
* GET `/`: retrieves the basic homepage
* POST `/l`: adds a URL mapping to the database and returns the generated 6-character key
* GET `/l/{key}`: returns a redirect to the URL that `key` maps to

The API is written in Go. To run locally, once all dependencies are installed:
```bash
export PORT=8080
export DB_ENDPOINT=mongodb://localhost:27017
go run server
```

### Redirection Rules Database
The redirection rules database uses MongoDB to store documents of the following structure:
```json
{
  "pk": string,
  "url": string
}
```
The `pk` (primary key) is simply the randomly generated key that is used to retrieve the redirection URL (`url`).

To start the DB locally, simply run:
```bash
mongod
```
Use `--dbpath=path/to/data` to specify a filepath to the DB data.
