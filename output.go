package vt_index

import (
	"os"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	m "github.com/murphy214/mercantile"
	"log"
	"fmt"
	"strconv"
)

// creating the slice that will be used to create the metadata table
func Make_Metadata_Slice(zoom int,filename string) [][]string {
	// getting the json blob metadata
	// creating values 
	values := [][]string{{"name",filename},{"type","overlay"},{"version","2"},{"description",filename},{"format","pbf"},{"zoom",strconv.Itoa(zoom)}}

	return values
}

func Create_Metadata(db *sql.DB) (*sql.Stmt,*sql.Tx) {

	sqlStmt := `
	CREATE TABLE metadata (name text, value text);
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
	}
	fmt.Print("Created metadata table.\n")

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("insert into metadata(value, name) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	return stmt,tx 
}

// creates the sqllite database and inserts metadata 
func Create_Database_Meta(filename string,zoom int) *sql.DB {
	os.Remove(filename)

	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatal(err)
	}
	//defer db.Close()
	var stmt *sql.Stmt
	var tx *sql.Tx
	fmt.Printf("Creating and opening %s.\n",filename)
	stmt,tx = Create_Metadata(db)

	// creating metadata slice string
	values := Make_Metadata_Slice(zoom,filename)


	defer stmt.Close()
	for _,i := range values {
		_, err = stmt.Exec(i[1],i[0])
		if err != nil {
			log.Fatal(err)
		}
	}
	tx.Commit()


	sqlStmt := `
	CREATE TABLE tiles (zoom_level integer, tile_column integer, tile_row integer, tile_data blob);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
	}

	tx, err = db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Inserted appropriate metadata: %s.\n",filename)


	return db
}


// creating an index for sqllite db
func Make_Index(db *sql.DB) {
	defer db.Close()

	sqlStmt := `
	CREATE UNIQUE INDEX IF NOT EXISTS tile_index on tiles (zoom_level, tile_column, tile_row)
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		//db.Close()

		fmt.Print(err,"\n")

		//_, err = db.Exec(sqlStmt)
		//return
	}

}


// inserting data into shit
func Insert_Data(newmap map[m.TileID][]byte,db *sql.DB) *sql.DB {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	
	stmt, err := tx.Prepare("insert into tiles(zoom_level, tile_column,tile_row,tile_data) values(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()
	count := 0
	total := 0
	count3 := 0


	sizenewmap := len(newmap)

	for k,v := range newmap {



		_, err = stmt.Exec(int(k.Z),int(k.X),int(k.Y),v)
		if err != nil {
			fmt.Print(err,"\n")
		}

		count += 1
		if count == 1000 {
			count = 0
			total += 1000
			fmt.Printf("\r[%d/%d] Compressing tiles and inserting into db.",total,sizenewmap)
		}

		count3 += 1
		//fmt.Print(count,"\n")
		//count += 1
	}



	tx.Commit()


	return db
	
}

