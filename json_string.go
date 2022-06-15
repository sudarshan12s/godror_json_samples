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

	// construct a sample JSON document represented as JSON string
	jsonstring := "{\"person\":{\"BirthDate\":\"1999-02-03T00:00:00\",\"ID\":\"12\",\"JoinDate\":\"2020-11-24T12:34:56.123000Z\",\"Name\":\"Alex\",\"RandomString\":\"APKZYKSv2\",\"age\":\"25\",\"creditScore\":[\"700\",\"250\",\"340\"],\"salary\":\"45.23\"}}"
	fmt.Printf("Input: \n %v \n\n", jsonstring)

	// Get the JSONValue type from Go string
	jsonval := godror.JSONString{Value: jsonstring, Flags: 0}
	if _, err = db.ExecContext(ctx, "INSERT INTO test(id, jdoc) VALUES(1 , :1)",
		jsonval); err != nil {
		log.Fatal(err)
	}

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

		// Get Go native map[string]interface{}
		v, err := jsondoc.GetValue(godror.JSONOptNumberAsString)
		if err != nil {
			log.Fatal(err)
		}
		// type assert to verify the concrete type stored
		gotmap, _ := v.(map[string]interface{})
		fmt.Printf("Fetch Document back to Go type:\n %v \n\n", gotmap)
	}
}

/* 
 * Output
 * 
 
 Input: 
 {"person":{"BirthDate":"1999-02-03T00:00:00","ID":"12","JoinDate":"2020-11-24T12:34:56.123000Z","Name":"Alex","RandomString":"APKZYKSv2","age":"25","creditScore":["700","250","340"],"salary":"45.23"}} 

Fetch Document as JSON string: 
 {"person":{"BirthDate":"1999-02-03T00:00:00","ID":"12","JoinDate":"2020-11-24T12:34:56.123000Z","Name":"Alex","RandomString":"APKZYKSv2","age":"25","creditScore":["700","250","340"],"salary":"45.23"}} 

Fetch Document back to Go type:
 map[person:map[BirthDate:1999-02-03T00:00:00 ID:12 JoinDate:2020-11-24T12:34:56.123000Z Name:Alex RandomString:APKZYKSv2 age:25 creditScore:[700 250 340] salary:45.23]] 
  
 */
