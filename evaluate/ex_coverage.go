// Implements Experiments and their methods
package evaluate

import(
  //"bufio"
  "github.com/staheri/goatlib/instrument"
  // "fmt"
  // "os"
  // "strconv"
  // "path/filepath"
  // _"time"
  // "os/exec"

)


type CoverageTable struct{
  ConcUsage            []*instrument.ConcurrencyUsage        `json:"concUsage"`
  CovTable              map[int]*CovReport                   `json:"-"`
  // we need a table here to update after each execution
}

type CovReport struct{
  Selects    bool
  Locks      bool
  Unlocks    bool
  Sends      bool
  Recvs      bool
  Waits      bool
  Adds       bool
  Closes     bool
}


func (cov *CoverageTable) Update (cr *CovReport){

}
