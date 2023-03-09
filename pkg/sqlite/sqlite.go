package sqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type (
	TSQLite3DB struct {
		db *sql.DB
	}

	TSQLite3ResultRows struct {
		rows *sql.Rows
	}

	TResultDataRow = map[string]string
)

func ConnectToDB(SQL3File string) (TSQLite3DB, error) {
	newdb, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?cache=shared&mode=rwc", SQL3File))
	return TSQLite3DB{db: newdb}, err
}

func (sqldb TSQLite3DB) Close() {
	sqldb.db.Close()
}

func (sqldb TSQLite3DB) QuerySingle(query string, args ...any) (sql.Result, error) {
	stmt, perr := sqldb.db.Prepare(query)
	if perr == nil {
		defer stmt.Close()
		res, eerr := stmt.Exec(args...)
		if eerr == nil {
			return res, eerr
		}
		return nil, eerr
	}
	return nil, perr
}

func (sqldb TSQLite3DB) QueryData(query string, args ...any) (TSQLite3ResultRows, error) {
	rawRows, err := sqldb.db.Query(query, args...)
	if err == nil {
		return TSQLite3ResultRows{rows: rawRows}, nil
	}
	return TSQLite3ResultRows{}, err
}

func (r TSQLite3ResultRows) UnloadNextRow() TResultDataRow {
	if r.rows.Next() {
		columns, cerr := r.rows.Columns()
		if cerr == nil {
			pointers := make([]any, len(columns))
			values := make([]string, len(columns))
			for i := range pointers {
				pointers[i] = &values[i]
			}
			scanerr := r.rows.Scan(pointers...)
			if scanerr == nil {
				result := make(TResultDataRow)
				for i := range values {
					result[columns[i]] = values[i]
				}
				return result
			}
		}
	}
	return nil
}
