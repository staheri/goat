package main

import "github.com/staheri/goatlib/instrument"
//import "fmt"

func test1(path string){
  instrument.Instrument(flagPath)
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
