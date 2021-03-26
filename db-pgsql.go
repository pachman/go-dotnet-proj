package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"strconv"
	"strings"
)

type PgsqlStorage struct {
	db *sql.DB
}

func (storage *PgsqlStorage) Init(connection string) (Storage, error) {

	db, err := sql.Open("postgres", connection)
	if err != nil {
		panic(err)
	}

	// https://github.com/theory/pg-semver/blob/main/doc/semver.mmd

	db.Exec(`CREATE TABLE packages(
 package varchar(500) NOT NULL,
 version SEMVER NOT NULL,
 original_version varchar(255) NOT NULL,
 file varchar(1000) NOT NULL,
 owner varchar(255) NOT NULL,
 repository varchar(255) NOT NULL,
 modify_date timestamp NOT NULL,
 PRIMARY KEY(package, file, owner, repository))`)

	db.Exec(`CREATE TABLE repositories(
 url varchar(500) NOT NULL,
 enable bool NOT NULL,
 PRIMARY KEY(url))`)

	db.Exec(`CREATE TABLE frameworks(
framework varchar(50) NOT NULL,
file varchar(1000) NOT NULL,
repository varchar(255) NOT NULL,
modify_date timestamp NOT NULL,
PRIMARY KEY(framework, file, repository))`)

	db.Exec(`CREATE TABLE framework_rate(
framework varchar(50) NOT NULL,
rate int NOT NULL,
PRIMARY KEY(framework))`)

	db.Exec(`CREATE TABLE package_threshold(
package varchar(500) NOT NULL,
lower SEMVER,
higher SEMVER,
rate int NOT NULL,
PRIMARY KEY(package))`)

	var storageDb Storage
	storageDb = &PgsqlStorage{db: db}

	return storageDb, nil
}

func (storage *PgsqlStorage) InsertPackages(packages []DotnetPackage) error {
	chunkList := chunk(unique(packages), 100)
	for _, chunk := range chunkList {
		var valueStrings []string
		var valueArgs []interface{}

		i := 1
		const fieldCount = 6

		for _, dp := range chunk {
			valueStrings = append(valueStrings, storage.ConcatPlaceholder(&i, fieldCount))

			valueArgs = append(valueArgs, dp.Package, dp.Version, dp.OriginalVersion, dp.File, dp.Owner, dp.Repository)
		}

		sql := "INSERT INTO packages (package, version, original_version, file, owner, repository, modify_date) VALUES " +
			strings.Join(valueStrings, ",") +
			"ON CONFLICT (package, file, owner, repository) DO UPDATE SET modify_date = excluded.modify_date, version = excluded.version, original_version = excluded.original_version"

		stmt, err := storage.db.Prepare(sql)
		if err != nil {
			return fmt.Errorf("InsertPackages | Prepare | %v", err)
		}

		_, err = stmt.Exec(valueArgs...)
		if err != nil {
			return fmt.Errorf("InsertPackages | Exec | %v", err)
		}
	}

	return nil
}

func (storage *PgsqlStorage) InsertFrameworks(frameworks []DotnetProjectFramework) error {

	var valueStrings []string
	var valueArgs []interface{}

	i := 1
	const fieldCount = 3

	for _, dp := range frameworks {
		valueStrings = append(valueStrings, storage.ConcatPlaceholder(&i, fieldCount))

		valueArgs = append(valueArgs, dp.Framework, dp.File, dp.Repository)
	}

	sql := "INSERT INTO frameworks (framework, file, repository, modify_date) VALUES " +
		strings.Join(valueStrings, ",") +
		"ON CONFLICT (framework, file, repository) DO UPDATE SET modify_date = excluded.modify_date"

	stmt, err := storage.db.Prepare(sql)
	if err != nil {
		return fmt.Errorf("InsertFrameworks | Prepare | %v", err)
	}

	_, err = stmt.Exec(valueArgs...)
	if err != nil {
		return fmt.Errorf("InsertFrameworks | Exec | %v", err)
	}

	return nil
}

func (storage *PgsqlStorage) ConcatPlaceholder(counter *int, fieldCount int) string {
	// "($n, $n+1, $n+2, $n+3, $n+4)"

	i := *counter
	values := "("
	for j := i; j < i+fieldCount; j++ {
		values += "$" + strconv.Itoa(j) + ","
	}
	values += "now())"
	*counter = i + fieldCount

	return values
}

func (storage *PgsqlStorage) SelectPackages() ([]DotnetPackage, error) {
	sql := "SELECT package, version, file, owner, repository FROM packages"

	var args []interface{}

	rows, err := storage.db.Query(sql, args...)
	if err != nil {
		return nil, fmt.Errorf("SelectPackages | Query | %v", err)
	}

	defer rows.Close()

	var packages []DotnetPackage

	for rows.Next() {
		p := DotnetPackage{}
		err := rows.Scan(&p.Package, &p.Version, &p.File, &p.Owner, &p.Repository)
		if err != nil {
			log.Warnf("SelectPackages | Next | %v", err)
			continue
		}
		packages = append(packages, p)
	}

	return packages, nil
}

func (storage *PgsqlStorage) SelectRepositories() ([]string, error) {
	sql := "SELECT url FROM repositories WHERE enable=true"

	rows, err := storage.db.Query(sql)
	if err != nil {
		return nil, fmt.Errorf("SelectRepositories | Query | %v", err)
	}

	defer rows.Close()

	var repositories []string

	for rows.Next() {
		var url string
		err := rows.Scan(&url)
		if err != nil {
			log.Warnf("SelectRepositories | Next | %v", err)
			continue
		}
		repositories = append(repositories, url)
	}

	return repositories, nil
}

func (storage *PgsqlStorage) Close() {
	storage.db.Close()
}
