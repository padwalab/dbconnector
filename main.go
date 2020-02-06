package main

import (
	"C"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/padwalab/gojs/gosrc"
)
import "os"

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	// Serve / with the index.html file.
	fs := http.FileServer(http.Dir("./lib"))
	http.Handle("/", fs)
	fmt.Print("server started at:")
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir)

	var db gosrc.Conn
	var currentStmtContext gosrc.ODBCStmt

	// Serve /callme with a text response.
	http.HandleFunc("/queryBuilder", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		a, err := db.FetchTables()
		if err != nil {
			fmt.Errorf("error %v", err)
		}
		data := `{"tables": ` + a + `}`
		// fmt.Println(strings.ReplaceAll(data, "null", "[]"))
		w.Header().Set("Content-Type", "application/json")
		// fmt.Print(data)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(data)
		fmt.Print("call recieved")
	})

	http.HandleFunc("/connection", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		DSN := r.FormValue("DSN")
		fmt.Println(DSN)
		db1, err := gosrc.Drv.Connect(DSN)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			fmt.Fprintf(w, "Connection failed %s", err)
		} else {
			db = *db1
			fmt.Println(gosrc.Drv.Stats)
			fmt.Fprintf(w, "Connection string is %s", DSN)
		}

	})

	http.HandleFunc("/disconnect", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		err := db.Close()
		if err != nil {
			log.Println(err)
		}
		fmt.Println(gosrc.Drv.Stats)
		fmt.Fprint(w, "Disconnected")
	})

	http.HandleFunc("/prepare", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		query := r.FormValue("queryPost")
		fmt.Println(query)
		currentStmtContext1, datum, err := db.PrepareODBCStmt(query)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			fmt.Fprintf(w, "Failed: %s", err)
		} else {
			currentStmtContext = *currentStmtContext1
			w.Header().Set("Content-Type", "application/json")
			// fmt.Print(data)
			w.WriteHeader(http.StatusCreated)
			data, _ := json.Marshal(datum)
			w.Write(data)
		}
		fmt.Println(datum)
		//fmt.Fprint(w, "Query Received: ", query)
	})

	http.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		paramData := r.FormValue("execPost")
		// if paramData != "{}" {
		fmt.Println(paramData)
		// args := []int{233, 22}
		// args = append(args, "ad")
		// currentStmtContext.Exec(args...)
		c := make(map[string]interface{})
		e := json.Unmarshal([]byte(paramData), &c)
		check(e)
		var a []interface{}
		for _, val := range c {
			// fmt.Println("key: ", key, " Value: ", val)
			for _, h := range val.(map[string]interface{}) {
				// fmt.Print("g: ", g, " h:", h)
				a = append(a, h)
			}
		}
		mrs, err := currentStmtContext.Exec(a)
		// err := currentStmtContext.Query(a)
		if err != nil {
			// fmt.Errorf("error %v", err)
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			fmt.Fprintf(w, "Query Execution failed %s", err)
		} else {
			w.Header().Set("Content-Type", "application/json")
			// fmt.Print(data)
			w.WriteHeader(http.StatusCreated)
			data, _ := json.Marshal(mrs)
			w.Write(data)
			// fmt.Println(data)
			// fmt.Fprint(w, "Data Recieved")
		}

		// fmt.Println(currentStmtContext)

		// }

	})

	http.HandleFunc("/columns", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		tableName := r.FormValue("tableName")
		fmt.Println(tableName)
		if !(tableName == "" || tableName == " ") {
			cols, err := db.FetchColumns(tableName)
			if err != nil {
				// fmt.Errorf("error %v", err)
				w.WriteHeader(http.StatusBadRequest)
				log.Println(err)
				fmt.Fprintf(w, "Connection failed %s", err)
			} else {
				data := `{ "columns": ` + cols + `, "text":"` + tableName + `"}`
				fmt.Println(data)
				w.Header().Set("Content-Type", "application/json")
				// fmt.Print(data)
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(data)
			}
		}
		fmt.Println("/table call recieved")
	})

	// Start the server at http://localhost:9000
	log.Print("starting server...")
	log.Fatal(http.ListenAndServe(":8000", nil))

}
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}
