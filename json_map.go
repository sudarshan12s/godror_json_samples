package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/godror/godror"
)

/*
 * Set connectString env
 * [bash]$setenv DB_URL "oracle://demo:demo@ip:1521/XEPDB1"
 */

func main() {

	// Get db pool object
	connectString := os.Getenv("DB_URL")
	db, err := sql.Open("godror", connectString)
	if err != nil { // nil means no error
		log.Fatal(err)
	}
	defer db.Close()

	// Cleanup
	db.Exec("DROP TABLE test")

	// set deadline
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// create a table test with JSON column type
	_, err = db.ExecContext(ctx,
		"CREATE TABLE test (id NUMBER(6), jdoc JSON)", //nolint:gas
	)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Exec("DROP TABLE test")

	// construct a sample JSON document with value as map
	birthdate, _ := time.Parse(time.UnixDate, "Wed Feb 25 11:06:39 PST 1990")
	var travelTime time.Duration = 5*time.Hour + 21*time.Minute + 10*time.Millisecond + 20*time.Nanosecond
	jsmap := map[string]interface{}{
		"person": map[string]interface{}{
			"ID":        godror.Number("12"),
			"FirstName": "Mary",
			"LastName":  "John",
			"creditScore": []interface{}{
				godror.Number("700"),
				godror.Number("250"),
				godror.Number("340"),
			},
			"age":              godror.Number("25"),
			"BirthDate":        birthdate,
			"salary":           godror.Number("45.23"),
			"Local":            true,
			"BinData":          []byte{0, 1, 2, 3, 4},
			"TravelTimePerDay": travelTime,
		},
	}
	fmt.Printf("Input: \n %v \n\n", jsmap)
	jsonval := godror.JSONValue{Value: jsmap}
	db.ExecContext(ctx, "INSERT INTO test(id, jdoc) VALUES(1 , :1)", jsonval)

	rows, err := db.QueryContext(ctx, "SELECT id, jdoc FROM test")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Convert the JSON document retrieved into a Go type
	var id interface{}
	var jsondoc godror.JSON

	for rows.Next() {
		// Read JSON document in OSON format from DB
		if err = rows.Scan(&id, &jsondoc); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Fetch Document as JSON string: \n %s \n\n", jsondoc)

		// Get Go native  map [string]interface{}
		v, _ := jsondoc.GetValue(godror.JSONOptNumberAsString)
		// type assert to verify the concrete type stored
		gotmap, _ := v.(map[string]interface{})
		fmt.Printf("Fetch Document back to Go type: \n %v \n\n", gotmap)
	}
}
