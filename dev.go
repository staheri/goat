package main

import "github.com/staheri/goatlib/instrument"
import "github.com/staheri/goatlib/db"
import "fmt"

func test1(path string){
  // iapp is the instrumented version of target program
  for i := 0 ; i < 10 ; i++{
    iapp := instrument.Instrument(flagPath)
    // execute
    events,err := iapp.ExecuteTrace()
    check(err)
    // store
    dbName := db.Store(events, iapp.Name)
    fmt.Println(dbName)
    db.Checker(dbName,false)
  }
  //
}


func check(err error){
	if err != nil{
		panic(err)
	}
}


// If s contains e
func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}
