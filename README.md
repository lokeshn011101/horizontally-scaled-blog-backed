# Horizontally Scaled Blog's Backend

This is a blog's backend that is horizontally scaled. It is written in Go and uses Postgres as a database.
Nginx load balancer was used to handle the traffic and the GETs are load tested using locust.

## How to run

1. Install Postgres and create a database with any name.
2. Replace the postgres connection string in `sql/sql.go` at line 13 with the connection string of your database.
3. Install Golang
4. Get all Go dependencies `go mod download`
5. Install Air `go install github.com/cosmtrek/air@latest`
6. Run the server `air -- 8080`
