# GOAT: A Framework for Analysis and Testing of Concurrent Go Applications
​
\goat is a combined static and dynamic concurrency testing
and analysis tool that facilitates the process of debugging for real-world Go programs.
%
Key ideas in \goat include
1) automated dynamic tracing to capture the behavior of concurrency primitives,
2) systematic schedule space exploration to accelerate the bug occurrence
and 3) deadlock detection with supplementary visualizations and reports.
\goat also propose a set of coverage requirements that characterize the dynamic behavior of concurrency primitives and provide metrics to measure the quality of tests.

All of above are done through goatlib and tuning parameters such as global deadlock timeout, visualization grain, number of runtime processes, etc.

\goat is available at \texttt{https://github.com/staheri/goat.git} and (ZENODO).
We are working on a docker version of goat to make it available for schedule testing through test packages.
​
## Build GoAT Runtime

GoAT is working in a custom runtime based on version [1.15.6](https://github.com/golang/go/tree/go1.15.6) of Golang currently only for Linux. It should be extensible to other architectures of the same version but they have not been tested.

### Prerequisite 0: install Go
Let's say your original Go installation is under `/usr/local/go` and your `$PATH` includes `/usr/local/go/bin` which means that you are able to run Go programs.
For installing Go, please refer to [install Go](https://golang.org/doc/install) and follow its simple steps to set paths.
​
### Prerequisite 1: rename original Go
Assuming your original Go installation is under `/usr/local/go`. First we rename the original go to `go-orig` and create a virutal link under the name `/usr/local/go` to easily switch between builds later.
```
$> sudo -i
$> mv /usr/local/go /usr/local/go-orig
$> ln -nsf /usr/local/go-orig /usr/local/go
```
​
### Prerequisite 2: download new Go
Download version 1.15.6.
​
```
$> sudo -i
$> wget https://golang.org/dl/go1.15.6.linux-amd64.tar.gz
$> mkdir -p /usr/local/go-goat
$> tar -xzvf go1.15.6.linux-amd64.tar.gz -C /usr/local/go-goat
```

###  Prerequisite 3: Set environment variables
For developing in Go, you have to set your $GOPATH and $GOROOT vars. $GOROOT is where the go runtime resides and $GOPATH is where your projects are.
```
$> export GOPATH=$HOME/gopath
$> mkdir $GOPATH
$> export GOROOT=/usr/local/go
$> export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
```
GoAT requires the GOATWS (workstation), GOATTO (global deadlock timeout) GOATMAXPROCS to be set as well.
```
$> export GOATWS=$HOME/goatws
$> mkdir -p $GOATWS
$> export GOATTO=30
$> export GOATMAXPROCS=4
```
You should add above lines to your `./.bashrc` to have them set each time you log in to the server.
`GOATMAXPROCS` is the number of CPU cores that GoAT uses for experimenting. The max number of cores that you can set is the number of your machine CPU cores.

You can change the path for `$GOATWS` as you wish. All the GoAT's output will be generated in the work station directory.
​
###  Prerequisite 4: Download GoAT to the correct path
In order to make sure that paths are available, execute below command first:
```
$> go get github.com/staheri/goat
```
It downloads GoAT and its dependencies under the right path (`$GOPATH/src/github.com/staheri/goat`) and installs the GoAT binary under `$GOPATH/bin/goat` but it will not work since we have to re-build GoAT with the custom runtime.

### Prerequisite 5: Patch and build the custom GoAT runtime:
```
$> cd /usr/local/go-goat/go
$> patch -p2 -d src/  < $GOPATH/src/github.com/staheri/goat/go1.15.6_goat_june15.patch
$> cd src/
$> export GOROOT_BOOTSTRAP=/usr/local/go-orig
$> ./make.bash
```
It will take a while. Then you need to make this build as the main Go runtime:
```
$> ln -nsf /usr/local/go-goat/go /usr/local/go
```
You can always switch back to your default Go by:
```
$> ln -nsf /usr/local/go-orig /usr/local/go
```

### Make GoAT:
```
$> cd $GOPATH/src/github.com/staheri/goat
$> go build -o $GOPATH/bin/goat
```

## GoAT Workflow
To print the help message, run `goat -h`:
```
$> goat -h
Initializing GOAT V.0.1 ...
Usage of bin/goat-single:
  -cov
        Include coverage report in evaluation
  -d int
        Number of delays
  -eval_conf string
        Config file with benchmark paths in it
  -freq int
        Frequency of executions (default 1)
  -path string
        Target folder
  -race
        Enable race detection
```

### Examples

#### Simple
Collect traces from the execution of `CodeBenchmark/goBench_goat/defSel/goatDefSel_test.go`, analyze traces for deadlocks, measure coverage and generate visualization:
```
$> ls CodeBenchmark/goBench_goat/defSel/
goatDefSel_test.go

$> goat -path=CodeBenchmark/goBench_goat/defSel
```

Output:
```
$> tree $GOATWS/p8/single_defSel
goatws/p8/single_defSel
└── goat_trace
    ├── bin
    │   └── 939498256trace
    ├── concUsage.json
    ├── out
    │   └── goat_goat_d0.out
    ├── results
    │   └── p8_defSel_goat_d0_T1_hitBug.json
    ├── src
    │   └── goatDefSel_test.go
    ├── traces
    │   └── goat_d0
    │       └── defSel_B0_I0.trace
    ├── traceTimes
    │   └── defSel_B0_I0.time
    └── visual
        ├── SUCC_defSel_B0_I0_fullVis.dot
        ├── SUCC_defSel_B0_I0_fullVis.pdf
        ├── SUCC_defSel_B0_I0_gtree.dot
        ├── SUCC_defSel_B0_I0_gtree.pdf
        ├── SUCC_defSel_B0_I0_minVis.dot
        └── SUCC_defSel_B0_I0_minVis.pdf
```

#### Add delay
You can add delay(s) around concurrency usages of the target code:
```
$> goat -path=CodeBenchmark/goBench_goat/defSel -d=2
```
Above code adds 2 delays around the concurrency usage of defSel which is already stored in `concUsage.json`

#### Test more than once until it fails
You can run tests multiple times (e.g., 1000 times) until it fails:
```
$> goat -path=CodeBenchmark/goBench_goat/defSel -d=3 -freq=1000
```


## Evaluation
In the IISWC paper, we evaluated GoAT against three deadlock detectors. To reproduce the results (table IV), you should follow below steps:

### Obtain other detectors
```
$> ln -nsf /usr/local/go-orig /usr/local/go

$> go get -u go.uber.org/goleak
$> go get golang.org/x/tools/cmd/goimports
$> go get github.com/sasha-s/go-deadlock

$> ln -nsf /usr/local/go-goat/go /usr/local/go
```

### Obtain GoBench
```
git clone https://github.com/goodmorning-coder/gobench.git
```

### Run evaluation
```
$> goat -conf=<path-to-conf> -freq=1000
```

`<path-to-conf>` should be a text file (similar to files in `configs` folder) where you list the paths to the bugs that you want to evaluate. For example, `configs/conf_attn_blocking_all.txt` has all the paths of all blocking bugs in GoKer. You have to follow the same pattern for paths as GoAT's naming conventions depend on the directories and subdirectories of the benchmark.

You can include `-cov` in your execution to show the coverage report as well.
