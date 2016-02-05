package main

/* Tim Henderson (tadh@case.edu)
*
* Copyright (c) 2015, Tim Henderson, Case Western Reserve University
* Cleveland, Ohio 44106. All Rights Reserved.
*
* This library is free software; you can redistribute it and/or modify
* it under the terms of the GNU General Public License as published by
* the Free Software Foundation; either version 3 of the License, or (at
* your option) any later version.
*
* This library is distributed in the hope that it will be useful, but
* WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
* General Public License for more details.
*
* You should have received a copy of the GNU General Public License
* along with this library; if not, write to the Free Software
* Foundation, Inc.,
*   51 Franklin Street, Fifth Floor,
*   Boston, MA  02110-1301
*   USA
 */

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/getopt"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners"
	"github.com/timtadh/sfp/miners/absorbing"
	"github.com/timtadh/sfp/miners/musk"
	"github.com/timtadh/sfp/miners/ospace"
	"github.com/timtadh/sfp/miners/premusk"
	"github.com/timtadh/sfp/miners/reporters"
	"github.com/timtadh/sfp/miners/fastmax"
	"github.com/timtadh/sfp/miners/uniprox"
	"github.com/timtadh/sfp/miners/walker"
	"github.com/timtadh/sfp/types/graph"
	"github.com/timtadh/sfp/types/itemset"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if urandom, err := os.Open("/dev/urandom"); err != nil {
		panic(err)
	} else {
		seed := make([]byte, 8)
		if _, err := urandom.Read(seed); err == nil {
			rand.Seed(int64(binary.BigEndian.Uint64(seed)))
		}
		urandom.Close()
	}
}

var ErrorCodes map[string]int = map[string]int{
	"usage":   0,
	"version": 2,
	"opts":    3,
	"badfloat":  6,
	"badint":  5,
	"baddir":  6,
	"badfile": 7,
}

var UsageMessage string = "sfp --help"
var ExtendedMessage string = `
sfp - sample frequent patterns

$ sfp -o <path> --samples=<int> --support=<int> [Global Options] \
    <type> [Type Options]
    <mode> [Mode Options]
    <input-path>

Note: You must supply [Global Options] then [<type> [Type Options]] then
[<mode> [Mode Options]] and finally <input-path>. Changes in ordering are not
supported.

Note: You may either supply the <input-path> as a regular file or a gzipped
file. If supplying a gzip file the file extension must be '.gz'.

Global Options
    -h, --help                          view this message
    --modes                             show the available modes
    --types                             show the available types
    -o, --ouput=<path>                  path to output directory (required)
                                        NB: will overwrite contents of dir
    -c, --cache=<path>                  path to cache directory (optional)
                                        NB: will overwrite contents of dir
    --samples=<int>                     number of samples to collect (required)
    --support=<int>                     minimum support of patterns (required)

Modes
    absorbing                           uses absorbing markov chain
    musk                                uniform sampling of maximal patterns
    ospace                              uniform sampling of all patterns
    fastmax                             faster sampling of large max patterns
                                        than absorbing
    uniprox                             approximately uniform sampling of max
                                        patterns using an absorbing chain

    uniprox Options
        -w, walks=<int>                 (default 15) number of estimating
                                        walks

Type: itemset

$ sfp -o /tmp/sfp --support=1000 --samples=10 \
    itemset --min-items=4 --max-items=4 \
    absorbing \
    ./data/transactions.dat.gz

itemset Options
    -h, --help                          view this message
    -l, --loader=<loader-name>          the loader to use (default int)
    --min-items=<int>                   minimum items in a samplable set
    --max-items=<int>                   maximum items in a samplable set

itemset Loaders
    int                                 each line is a transaction
                                        the items are integers
                                        the items are space separated
       ex.
            10 1 5 7
            213 2 5 1
            23 1 4 5 7
            3 4 1


Type: graph

$ sfp -o /tmp/sfp --support=5 --samples=100 \
    graph --min-vertices=5 --max-vertices=8 --max-edges=15 \
    absorbing \
    ./data/graph.veg.gz

graph Options
    -h, --help                          view this message
    -l, --loader=<loader-name>          the loader to use (default veg)
    --min-edges=<int>                   minimum edges in a samplable graph
    --max-edges=<int>                   maximum edges in a samplable graph
    --min-vertices=<int>                minimum vertices in a samplable graph
    --max-vertices=<int>                maximum vertices in a samplable graph

graph Loaders
    veg File Format

        The veg file format is a line delimited format with vertex lines and
        edge lines. For example:

        vertex	{"id":136,"label":""}
        edge	{"src":23,"targ":25,"label":"ddg"}

    veg Grammar

        line -> vertex "\n"
              | edge "\n"

        vertex -> "vertex" "\t" vertex_json

        edge -> "edge" "\t" edge_json

        vertex_json -> {"id": int, "label": string, ...}
        // other items are optional

        edge_json -> {"src": int, "targ": int, "label": int, ...}
        // other items are  optional
`

func Usage(code int) {
	fmt.Fprintln(os.Stderr, UsageMessage)
	if code == 0 {
		fmt.Fprintln(os.Stdout, ExtendedMessage)
		code = ErrorCodes["usage"]
	} else {
		fmt.Fprintln(os.Stderr, "Try -h or --help for help")
	}
	os.Exit(code)
}

func Input(input_path string) (reader io.Reader, closeall func()) {
	stat, err := os.Stat(input_path)
	if err != nil {
		panic(err)
	}
	if stat.IsDir() {
		return InputDir(input_path)
	} else {
		return InputFile(input_path)
	}
}

func InputFile(input_path string) (reader io.Reader, closeall func()) {
	freader, err := os.Open(input_path)
	if err != nil {
		panic(err)
	}
	if strings.HasSuffix(input_path, ".gz") {
		greader, err := gzip.NewReader(freader)
		if err != nil {
			panic(err)
		}
		return greader, func() {
			greader.Close()
			freader.Close()
		}
	}
	return freader, func() {
		freader.Close()
	}
}

func InputDir(input_dir string) (reader io.Reader, closeall func()) {
	var readers []io.Reader
	var closers []func()
	dir, err := ioutil.ReadDir(input_dir)
	if err != nil {
		panic(err)
	}
	for _, info := range dir {
		if info.IsDir() {
			continue
		}
		creader, closer := InputFile(path.Join(input_dir, info.Name()))
		readers = append(readers, creader)
		closers = append(closers, closer)
	}
	reader = io.MultiReader(readers...)
	return reader, func() {
		for _, closer := range closers {
			closer()
		}
	}
}

func ParseInt(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing '%v' expected an int\n", str)
		Usage(ErrorCodes["badint"])
	}
	return i
}

func ParseFloat(str string) float64 {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing '%v' expected a float\n", str)
		Usage(ErrorCodes["badfloat"])
	}
	return f
}

func AssertDir(dir string) string {
	dir = path.Clean(dir)
	fi, err := os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0775)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			Usage(ErrorCodes["baddir"])
		}
		return dir
	} else if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		Usage(ErrorCodes["baddir"])
	}
	if !fi.IsDir() {
		fmt.Fprintf(os.Stderr, "Passed in file was not a directory, %s", dir)
		Usage(ErrorCodes["baddir"])
	}
	return dir
}

func EmptyDir(dir string) string {
	dir = path.Clean(dir)
	_, err := os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0775)
		if err != nil {
			log.Fatal(err)
		}
	} else if err != nil {
		log.Fatal(err)
	} else {
		// something already exists lets delete it
		err := os.RemoveAll(dir)
		if err != nil {
			log.Fatal(err)
		}
		err = os.MkdirAll(dir, 0775)
		if err != nil {
			log.Fatal(err)
		}
	}
	return dir
}

func AssertFile(fname string) string {
	fname = path.Clean(fname)
	fi, err := os.Stat(fname)
	if err != nil && os.IsNotExist(err) {
		return fname
	} else if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		Usage(ErrorCodes["badfile"])
	} else if fi.IsDir() {
		fmt.Fprintf(os.Stderr, "Passed in file was a directory, %s", fname)
		Usage(ErrorCodes["badfile"])
	}
	return fname
}

func itemsetType(argv []string, conf *config.Config) (lattice.Loader, func(lattice.DataType) lattice.Formatter, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"hl:", []string{"help", "loader=", "min-items=", "max-items="},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		Usage(ErrorCodes["opts"])
	}

	loaderType := "int"
	min := 0
	max := int(math.MaxInt32)
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			Usage(0)
		case "-l", "--loader":
			loaderType = oa.Arg()
		case "--min-items":
			min = ParseInt(oa.Arg())
		case "--max-items":
			max = ParseInt(oa.Arg())
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}

	var loader lattice.Loader
	switch loaderType {
	case "int":
		loader, err = itemset.NewIntLoader(conf, min, max)
	default:
		fmt.Fprintf(os.Stderr, "Unknown itemset loader '%v'\n", loaderType)
		Usage(ErrorCodes["opts"])
	}
	if err != nil {
		log.Panic(err)
	}
	fmtr := func(_ lattice.DataType) lattice.Formatter {
		return itemset.Formatter{}
	}
	return loader, fmtr, args
}

func graphType(argv []string, conf *config.Config) (lattice.Loader, func(lattice.DataType) lattice.Formatter, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"hl:", []string{"help", "loader=",
			"min-edges=",
			"max-edges=",
			"min-vertices=",
			"max-vertices=",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		Usage(ErrorCodes["opts"])
	}

	loaderType := "veg"
	minE := 0
	maxE := int(math.MaxInt32)
	minV := 0
	maxV := int(math.MaxInt32)
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			Usage(0)
		case "-l", "--loader":
			loaderType = oa.Arg()
		case "--min-edges":
			minE = ParseInt(oa.Arg())
		case "--max-edges":
			maxE = ParseInt(oa.Arg())
		case "--min-vertices":
			minV = ParseInt(oa.Arg())
		case "--max-vertices":
			maxV = ParseInt(oa.Arg())
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}

	var loader lattice.Loader
	switch loaderType {
	case "veg":
		loader, err = graph.NewVegLoader(conf, minE, maxE, minV, maxV)
	default:
		fmt.Fprintf(os.Stderr, "Unknown itemset loader '%v'\n", loaderType)
		Usage(ErrorCodes["opts"])
	}
	if err != nil {
		log.Panic(err)
	}
	fmtr := func(dt lattice.DataType) lattice.Formatter {
		g := dt.(*graph.Graph)
		return graph.NewFormatter(g)
	}
	return loader, fmtr, args
}

func absorbingMode(argv []string, conf *config.Config) (miners.Miner, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"hc",
		[]string{
			"help",
			"compute-pr-matrices",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		Usage(ErrorCodes["opts"])
	}
	computePrMatrices := false
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			Usage(0)
		case "-c", "--compute-pr-matrices":
			computePrMatrices = true
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}
	return absorbing.NewWalker(conf, computePrMatrices), args
}

func fastmaxMode(argv []string, conf *config.Config) (miners.Miner, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"h",
		[]string{
			"help",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		Usage(ErrorCodes["opts"])
	}
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			Usage(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}
	return fastmax.NewWalker(conf), args
}

func uniproxMode(argv []string, conf *config.Config) (miners.Miner, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"hw:",
		[]string{
			"help",
			"walks=",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		Usage(ErrorCodes["opts"])
	}
	walks := 15
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			Usage(0)
		case "-w", "--walks":
			walks = ParseInt(oa.Arg())
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}
	miner, err := uniprox.NewWalker(conf, walks)
	if err != nil {
		log.Fatal(err)
	}
	return miner, args
}

func muskMode(argv []string, conf *config.Config) (miners.Miner, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"h",
		[]string{
			"help",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		Usage(ErrorCodes["opts"])
	}
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			Usage(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}
	miner := walker.NewWalker(conf, musk.MakeMaxUniformWalk(musk.Next, nil))
	return miner, args
}

func premuskMode(argv []string, conf *config.Config) (miners.Miner, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"h",
		[]string{
			"help",
			"teleport=",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		Usage(ErrorCodes["opts"])
	}
	teleport := .01
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			Usage(0)
		case "--teleport":
			teleport = ParseFloat(oa.Arg())
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}
	miner := premusk.NewWalker(conf, teleport)
	return miner, args
}

func ospaceMode(argv []string, conf *config.Config) (miners.Miner, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"h",
		[]string{
			"help",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		Usage(ErrorCodes["opts"])
	}
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			Usage(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}
	miner := walker.NewWalker(conf, ospace.MakeUniformWalk(0, true))
	return miner, args
}

func types(argv []string, conf *config.Config) (lattice.Loader, func(lattice.DataType) lattice.Formatter, []string) {
	switch argv[0] {
	case "itemset":
		return itemsetType(argv[1:], conf)
	case "graph":
		return graphType(argv[1:], conf)
	default:
		fmt.Fprintf(os.Stderr, "Unknown data type '%v'\n", argv[0])
		Usage(ErrorCodes["opts"])
		panic("unreachable")
	}
}

func modes(argv []string, conf *config.Config) (miners.Miner, []string) {
	switch argv[0] {
	case "absorbing":
		return absorbingMode(argv[1:], conf)
	case "fastmax":
		return fastmaxMode(argv[1:], conf)
	case "musk":
		return muskMode(argv[1:], conf)
	case "ospace":
		return ospaceMode(argv[1:], conf)
	case "premusk":
		return premuskMode(argv[1:], conf)
	case "uniprox":
		return uniproxMode(argv[1:], conf)
	default:
		fmt.Fprintf(os.Stderr, "Unknown mining mode '%v'\n", argv[0])
		Usage(ErrorCodes["opts"])
		panic("unreachable")
	}
}

func main() {
	args, optargs, err := getopt.GetOpt(
		os.Args[1:],
		"ho:c:",
		[]string{
			"help", "output=", "cache=", "modes", "types",
			"support=",
			"samples=",
			"skip-log=",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "could not process your arguments (perhaps you forgot a mode?) try:")
		fmt.Fprintf(os.Stderr, "$ %v breadth %v\n", os.Args[0], strings.Join(os.Args[1:], " "))
		Usage(ErrorCodes["opts"])
	}

	output := ""
	cache := ""
	support := 0
	samples := 0
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			Usage(0)
		case "-o", "--output":
			output = EmptyDir(oa.Arg())
		case "-c", "--cache":
			cache = EmptyDir(oa.Arg())
		case "--support":
			support = ParseInt(oa.Arg())
		case "--samples":
			samples = ParseInt(oa.Arg())
		case "--types":
			fmt.Fprintln(os.Stderr, "Types: itemset, graph")
			os.Exit(0)
		case "--modes":
			fmt.Fprintln(os.Stderr, "Modes: absorbing, musk, ospace, fastmax")
			os.Exit(0)
		case "--skip-log":
			level := oa.Arg()
			errors.Logf("INFO", "not logging level %v", level)
			errors.SkipLogging[level] = true
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}

	if support <= 0 {
		fmt.Fprintf(os.Stderr, "Support <= 0, must be > 0\n")
		Usage(ErrorCodes["opts"])
	}

	if samples <= 0 {
		fmt.Fprintf(os.Stderr, "Samples <= 0, must be > 0\n")
		Usage(ErrorCodes["opts"])
	}

	if output == "" {
		fmt.Fprintf(os.Stderr, "You must supply an output dir (-o)\n")
		Usage(ErrorCodes["opts"])
	}

	conf := &config.Config{
		Cache:   cache,
		Output:  output,
		Support: support,
		Samples: samples,
	}

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "You must supply a type and a mode)\n")
		Usage(ErrorCodes["opts"])
	}
	loader, fmtr, args := types(args, conf)

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "You must supply a mode\n")
		Usage(ErrorCodes["opts"])
	}
	mode, args := modes(args, conf)

	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "You must supply exactly one input path\n")
		fmt.Fprintf(os.Stderr, "You gave: %v\n", args)
		Usage(ErrorCodes["opts"])
	}

	getInput := func() (io.Reader, func()) {
		return Input(args[0])
	}

	errors.Logf("INFO", "Got configuration about to load dataset")
	dt, err := loader.Load(getInput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "There was error during the loading process\n")
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fr, err := reporters.NewFile(conf, fmtr(dt))
	if err != nil {
		fmt.Fprintf(os.Stderr, "There was error creating output files\n")
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	rptr := &reporters.Chain{[]miners.Reporter{&reporters.Log{}, reporters.NewUnique(fr)}}

	errors.Logf("INFO", "loaded data, about to start mining")
	mineErr := mode.Mine(dt, rptr)

	code := 0
	if e := mode.Close(); e != nil {
		errors.Logf("ERROR", "error closing %v", e)
		code++
	}
	if mineErr != nil {
		fmt.Fprintf(os.Stderr, "There was error during the mining process\n")
		fmt.Fprintf(os.Stderr, "%v\n", mineErr)
		code++
	} else {
		errors.Logf("INFO", "Done!")
	}
	os.Exit(code)
}
