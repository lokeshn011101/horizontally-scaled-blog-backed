package main

import (
	"blog/sql"
	"blog/types"
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jaswdr/faker"
	"golang.org/x/exp/slices"
)

func SetupFakeData() {
	DropTables()
	CreateTables()
	GenerateAndInsertFakeUsers(10)
	GenerateAndInsertFakeBlogs(1000, 10)
	GenerateAndInsertFakeTags(10)
	AssociateBlogsWithTags(1000, 10)
}

func DropTables() {
	fmt.Println("Dropping existing tables")
	pool := sql.ConnectToPostgres()

	conn, err := pool.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Cannot acquire conn from pool %s", err)
		return
	}

	query := `
DROP TABLE IF EXISTS "public"."blog_tags" CASCADE;
DROP TABLE IF EXISTS "public"."blog_users" CASCADE;
DROP TABLE IF EXISTS "public"."blog_blogs" CASCADE;
DROP TABLE IF EXISTS "public"."blogs_tags";
DROP SEQUENCE IF EXISTS "blog_users_id_seq";
DROP SEQUENCE IF EXISTS "blog_blogs_id_seq";
DROP SEQUENCE IF EXISTS "blog_tags_id_seq";
`
	_, err = conn.Exec(context.Background(), query)
	if err != nil {
		fmt.Println(err)
		return
	}
	conn.Release()
	fmt.Println("Tables dropped")
}

func CreateTables() {
	fmt.Println("Creating tables")
	query := `
CREATE SEQUENCE IF NOT EXISTS blog_users_id_seq;

-- Table Definition
CREATE TABLE "public"."blog_users" (
    "id" int4 NOT NULL DEFAULT nextval('blog_users_id_seq'::regclass),
    "count" int4,
    "name" varchar,
    "interested_tags" _int4 DEFAULT '{}'::integer[],
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS blog_blogs_id_seq;

-- Table Definition
CREATE TABLE "public"."blog_blogs" (
    "id" int4 NOT NULL DEFAULT nextval('blog_blogs_id_seq'::regclass),
    "title" varchar,
    "text" text,
    "images" _varchar,
    "user_id" int4 NOT NULL,
    "is_deleted" bool DEFAULT false,
    "excerpt" varchar,
    "modified_at" int8 DEFAULT (EXTRACT(epoch FROM now()))::bigint,
    "created_at" int8 DEFAULT (EXTRACT(epoch FROM now()))::bigint,
    CONSTRAINT "users_id" FOREIGN KEY ("user_id") REFERENCES "public"."blog_users"("id") ON DELETE CASCADE,
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS blog_tags_id_seq;

-- Table Definition
CREATE TABLE "public"."blog_tags" (
    "id" int4 NOT NULL DEFAULT nextval('blog_tags_id_seq'::regclass),
    "name" varchar NOT NULL,
    PRIMARY KEY ("id")
);

CREATE TABLE "public"."blogs_tags" (
    "blog_id" int4 NOT NULL,
    "tag_id" int4 NOT NULL,
    CONSTRAINT "blogs_tags_tag_id_fkey" FOREIGN KEY ("tag_id") REFERENCES "public"."blog_tags"("id"),
    CONSTRAINT "blogs_tags_blog_id_fkey" FOREIGN KEY ("blog_id") REFERENCES "public"."blog_blogs"("id"),
    PRIMARY KEY ("blog_id","tag_id")
);

CREATE OR REPLACE FUNCTION update_modified_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = EXTRACT(EPOCH FROM NOW())::BIGINT;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER update_blogs_modified_at
BEFORE UPDATE ON blog_blogs
FOR EACH ROW
EXECUTE FUNCTION update_modified_at();
`
	pool := sql.ConnectToPostgres()

	conn, err := pool.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Cannot acquire conn from pool %s", err)
		return
	}

	_, err = conn.Exec(context.Background(), query)
	if err != nil {
		fmt.Println(err)
		return
	}
	conn.Release()
	fmt.Println("Tables created")
}

func GenerateAndInsertFakeTags(count int) {
	fmt.Println("Generating and inserting fake tags")
	pool := sql.ConnectToPostgres()

	conn, err := pool.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Cannot acquire conn from pool %s", err)
		return
	}

	generateAndInsertFakeTags(count, conn)
	// printTable[Blog_tags](conn, "blog_tags")

	conn.Release()
	fmt.Println("Fake tags generated and inserted")
}

func GenerateAndInsertFakeUsers(count int) {
	fmt.Println("Generating and inserting fake users")
	pool := sql.ConnectToPostgres()

	conn, err := pool.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Cannot acquire conn from pool %s", err)
		return
	}

	generateAndInsertFakeUsers(count, conn)
	printTable[types.Blog_users](conn, "blog_users")

	conn.Release()
	fmt.Println("Fake users generated and inserted")
}

func GenerateAndInsertFakeBlogs(count int, usersCount int) {
	fmt.Println("Generating and inserting fake blogs")
	pool := sql.ConnectToPostgres()

	conn, err := pool.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Cannot acquire conn from pool %s", err)
		return
	}

	generateAndInsertFakeBlogs(count, usersCount, conn)
	updateUserBlogsCount(conn, count, usersCount)
	// printTable[Blog_blogs](conn, "blog_blogs")

	conn.Release()
	fmt.Println("Fake blogs generated and inserted")
}

func AssociateBlogsWithTags(blogCount int, tagCount int) {
	fmt.Println("Associating blogs with tags")
	pool := sql.ConnectToPostgres()

	conn, err := pool.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Cannot acquire conn from pool %s", err)
		return
	}

	for i := 1; i <= blogCount; i++ {
		for j := 1; j <= tagCount; j++ {
			insertIntoTable(conn, "blogs_tags", "(blog_id, tag_id)", fmt.Sprintf("(%d, %d)", i, j))
		}
	}
	// printTable[Blogs_Tags](conn, "blogs_tags")

	conn.Release()
	fmt.Println("Blogs associated with tags")
}

func generateAndInsertFakeBlogs(count int, usersCount int, conn *pgxpool.Conn) {
	fakeBlogs := generateFakeBlogs(count, usersCount)
	insertIntoTable(conn, "blog_blogs", "(title, text, user_id, excerpt, created_at)", fakeBlogs)
}

func updateUserBlogsCount(conn *pgxpool.Conn, count int, usersCount int) {
	_, err := conn.Exec(context.Background(), "update blog_users set count = 10")
	if err != nil {
		fmt.Println(err)
		return
	}
}

func generateFakeBlogs(count int, usersCount int) string {
	fake := faker.New()
	var values string
	for i := 0; i < count; i++ {
		for j := 1; j <= usersCount; j++ {
			title := fake.Lorem().Sentence(4 + rand.Intn(4))
			text := fake.Lorem().Words(1500 + rand.Intn(500))
			user_id := j
			values += fmt.Sprintf("('%s', '%s', %d, '%s'),", strings.Title(title), text, user_id, strings.ToLower(title))
		}
	}
	return strings.TrimSuffix(values, ",")
}

func generateAndInsertFakeTags(count int, conn *pgxpool.Conn) {
	fakeTags := generateFakeTags(count)
	insertIntoTable(conn, "blog_tags", "(name)", fakeTags)
}

func generateFakeTags(count int) string {
	fake := faker.New()
	var values string
	for i := 0; i < count; i++ {
		name := fake.Language().ProgrammingLanguage()
		values += fmt.Sprintf("('%s'),", name)
	}
	return strings.TrimSuffix(values, ",")
}

func insertIntoTable(conn *pgxpool.Conn, table string, columns string, values string) {
	tx, err := conn.Begin(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), fmt.Sprintf("insert into %s %s values %s", table, columns, values))
	if err != nil {
		fmt.Println(err)
		return
	}

	err = tx.Commit(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
}

func generateAndInsertFakeUsers(count int, conn *pgxpool.Conn) {
	fakeUsers := generateFakeUsers(count)
	insertIntoTable(conn, "blog_users", "(name, count)", fakeUsers)
}

func generateFakeUsers(count int) string {
	fake := faker.New()
	var values string
	for i := 0; i < count; i++ {
		name := fake.Person().FirstName()
		values += fmt.Sprintf("('%s', %d, ARRAY[%s]),", name, 0, strings.Join(GenerateInterestedTagsOfUser(10), ","))
	}
	return strings.TrimSuffix(values, ",")
}

func GenerateInterestedTagsOfUser(tagCount int) []string {
	fmt.Println("Generating interested tags for user")

	totalTagsForUser := 1 + rand.Intn(tagCount)
	interestedTagsForUser := []string{}
	for j := 1; j <= totalTagsForUser; j++ {
		fakeTag := strconv.Itoa(1 + rand.Intn(tagCount))
		for idx := slices.IndexFunc(interestedTagsForUser, func(i string) bool { return i == fakeTag }); idx != -1; idx = slices.IndexFunc(interestedTagsForUser, func(i string) bool { return i == fakeTag }) {
			fakeTag = strconv.Itoa(1 + rand.Intn(tagCount))
		}
		interestedTagsForUser = append(interestedTagsForUser, fakeTag)
	}

	fmt.Println("Interested tags for user generated")
	return interestedTagsForUser
}

func printTable[T any](conn *pgxpool.Conn, table string) {
	rows, _ := conn.Query(context.Background(), fmt.Sprintf("select * from %s", table))
	rowsData, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		fmt.Printf("CollectRows error: %v", err)
		return
	}

	for _, p := range rowsData {
		fmt.Printf("%v\n", p)
	}
}
