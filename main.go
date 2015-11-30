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
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
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
	"github.com/timtadh/sfp/miners/reporters"
	"github.com/timtadh/sfp/miners/unisorb"
	"github.com/timtadh/sfp/miners/walker"
	"github.com/timtadh/sfp/types/itemset"
	"github.com/timtadh/sfp/types/graph"
)


func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

var ErrorCodes map[string]int = map[string]int{
	"usage":   0,
	"version": 2,
	"opts":    3,
	"badint":  5,
	"baddir":  6,
	"badfile": 7,
}

var UsageMessage string = "sfp --help"
var ExtendedMessage string = `
sfp - sample frequent patterns

sfp <type> [Type Options] <mode> [Mode Options] -o <output-path> <input-path>

Global Options
    -h, --help                          view this message
    --modes                             show the available modes
    --types                             show the available types
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
		"hl:", []string{ "help", "loader=", "min-items=", "max-items="},
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
	case "int": loader, err = itemset.NewIntLoader(conf, min, max)
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
		"hl:", []string{ "help", "loader=",
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
	case "veg": loader, err = graph.NewVegLoader(conf, minE, maxE, minV, maxV)
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
	miner := walker.NewWalker(conf, absorbing.RejectingWalk)
	return miner, args
}

func unisorbMode(argv []string, conf *config.Config) (miners.Miner, []string) {
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
	miner := walker.NewWalker(conf, unisorb.MaxUniformWalk)
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
	miner := walker.NewWalker(conf, musk.MaxUniformWalk)
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
	miner := walker.NewWalker(conf, ospace.UniformWalk)
	return miner, args
}

func types(argv []string, conf *config.Config) (lattice.Loader, func(lattice.DataType) lattice.Formatter, []string) {
	switch argv[0] {
	case "itemset": return itemsetType(argv[1:], conf)
	case "graph": return graphType(argv[1:], conf)
	default:
		fmt.Fprintf(os.Stderr, "Unknown data type '%v'\n", argv[0])
		Usage(ErrorCodes["opts"])
		panic("unreachable")
	}
}

func modes(argv []string, conf *config.Config) (miners.Miner, []string) {
	switch argv[0] {
	case "absorbing": return absorbingMode(argv[1:], conf)
	case "unisorb": return unisorbMode(argv[1:], conf)
	case "musk": return muskMode(argv[1:], conf)
	case "ospace": return ospaceMode(argv[1:], conf)
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
			fmt.Fprintln(os.Stderr, "Types")
			os.Exit(0)
		case "--modes":
			fmt.Fprintln(os.Stderr, "Modes")
			os.Exit(0)
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
		Cache: cache,
		Output: output,
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
	defer dt.Close()

	fr, err := reporters.NewFile(conf, fmtr(dt))
	if err != nil {
		fmt.Fprintf(os.Stderr, "There was error creating output files\n")
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	rptr := &reporters.Chain{[]miners.Reporter{&reporters.Log{}, fr}}
	defer rptr.Close()

	errors.Logf("INFO", "loaded data, about to start mining")
	err = mode.Mine(dt, rptr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "There was error during the mining process\n")
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	} else {
		log.Println("Done!")
	}
}

