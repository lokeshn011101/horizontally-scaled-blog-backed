package sql

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectToPostgres() *pgxpool.Pool {
	url := "postgres://postgres:root@localhost:5432/postgres"
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Connected to database")
	return pool
}

func ExecSelectQuery[T any](conn *pgx.Conn, table string) ([]T, error) {
	rows, _ := conn.Query(context.Background(), fmt.Sprintf("select * from %s", table))
	rowsData, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		fmt.Printf("CollectRows error: %v", err)
		return nil, err
	}
	return rowsData, nil
}
