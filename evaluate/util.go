// Utility functions for package evaluate
package evaluate

import(
  "time"
  "log"
  "os"
  "bufio"
)


func ReadLines(f string) (lines []string){
  file, err := os.Open(f)
  if err != nil {
      log.Fatal(err)
  }
  defer file.Close()

  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
      lines = append(lines,scanner.Text())
  }

  if err := scanner.Err(); err != nil {
      log.Fatal(err)
  }
  return lines
}



func MeasureTime(fn func()) (et time.Duration) {
  start := time.Now()
  fn()
  end := time.Now()
  et = end.Sub(start)
  return et
}

func fileExist(filename string) bool {
  _, err := os.Stat(filename)
  return !os.IsNotExist(err)
}

func checkFile(filename string) bool {
  fi, err := os.Stat(filename)
  return !os.IsNotExist(err) && fi.Size()!=0
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
