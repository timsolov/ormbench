package main_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/davecgh/go-spew/spew"
	ps "github.com/georgysavva/scany/pgxscan"
	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db *sql.DB
	_  = spew.UnsafeDisabled

	ozzoDB  *dbx.DB
	gormDB  *gorm.DB
	sqlxDB  *sqlx.DB
	pgxPool *pgxpool.Pool
)

// Add DB credentials here
var (
	host   string = "localhost"
	port   int    = 5432
	user   string = "postgres"
	pass   string = "pass"
	dbname string = "ormbench"
)

func init() {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbname)
	db, _ = sql.Open("pgx", dsn)
	db.SetMaxOpenConns(1)
	err := db.Ping()
	if err != nil {
		panic(err)
	}

	ozzoDB = dbx.NewFromDB(db, "pgx")
	gormDB, err = gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	sqlxDB = sqlx.NewDb(db, "pgx")

	// pgxpool
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic(err)
	}
	pgxPool, err = pgxpool.ConnectConfig(context.Background(), cfg)
	if err != nil {
		panic(err)
	}
}

type model struct {
	ID    int    `dbq:"id" gorm:"column:id" db:"id"`
	Name  string `dbq:"name" gorm:"column:name" db:"name"`
	Email string `dbq:"email" gorm:"column:email" db:"email"`
}

// Required by gorm
func (model) TableName() string {
	return "tests"
}

func Benchmark(b *testing.B) {
	setup()
	defer cleanup()

	limits := []int{
		5,
		50,
		100,
		500,
		10000,
	}

	for _, lim := range limits {
		lim := lim

		// Benchmark ozzo-dbx
		b.Run(fmt.Sprintf("ozzo-dbx limit:%d", lim), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var res []model
				err := ozzoDB.Select().From("tests").OrderBy("id").Limit(int64(lim)).All(&res)
				if err != nil {
					b.Fatal(err)
				}
				if len(res) != lim {
					panic("something is wrong")
				}
			}
		})

		// Benchmark sqlx
		b.Run(fmt.Sprintf("sqlx limit:%d", lim), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var res []model
				q := fmt.Sprintf("SELECT id, name, email FROM tests ORDER BY id LIMIT %d", lim)
				err := sqlxDB.Select(&res, q)
				if err != nil {
					b.Fatal(err)
				}
				if len(res) != lim {
					panic("something is wrong")
				}
			}
		})

		// Benchmark gorm v2
		b.Run(fmt.Sprintf("gorm v2 limit:%d", lim), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var res []model
				err := gormDB.Order("id").Limit(lim).Find(&res).Error
				if err != nil {
					b.Fatal(err)
				}
				if len(res) != lim {
					panic("something is wrong")
				}
			}
		})

		build := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
		// Benchmark pgx v4 with pgxscan
		b.Run(fmt.Sprintf("pgxpool limit:%d", lim), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var res []model

				q, args, err := build.Select("*").From("tests").OrderBy("id").Limit(uint64(lim)).ToSql()
				if err != nil {
					b.Fatal(err)
				}

				err = ps.Select(context.Background(), pgxPool, &res, q, args...)
				if err != nil {
					b.Fatal(err)
				}

				if len(res) != lim {
					panic("something is wrong")
				}
			}
		})

		fmt.Println("========================================================================")
	}

}

func setup() {
	// Create table
	createQ := `
	CREATE TABLE tests (
		id int NOT NULL,
		name varchar(50) NOT NULL DEFAULT '',
		email varchar(150) NOT NULL DEFAULT '',
		PRIMARY KEY (id)
	)`

	_, err := db.Exec(createQ)
	if err != nil {
		panic(err)
	}

	tx := gormDB.Begin()
	// Add 10,000 fake entries
	for i := 0; i < 10000; i++ {
		err = tx.Create(&model{i + 1,
			gofakeit.Name(),
			gofakeit.Email()}).Error
		if err != nil {
			tx.Rollback()
			panic(err)
		}
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		panic(err)
	}
}

func cleanup() {
	_, err := db.Exec(`DROP TABLE tests`)
	if err != nil {
		panic(err)
	}
}
