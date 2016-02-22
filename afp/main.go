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
	"github.com/timtadh/sfp/afp/miners/vsigram"
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners"
	"github.com/timtadh/sfp/miners/reporters"
	"github.com/timtadh/sfp/types/digraph"
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

$ sfp -o <path> [Global Options] \
    <type> [Type Options] <input-path> \
    <mode> [Mode Options] \
    [<reporter> [Reporter Options]]

Note: You must supply [Global Options] then [<type> [Type Options]] then
      [<mode> [Mode Options]] and finally <input-path>. Changes in ordering are
      not supported.

Note: You may either supply the <input-path> as a regular file or a gzipped
      file. If supplying a gzip file the file extension must be '.gz'.

Note: If you don't supply a reporter by default it will use 'chain log file'.
      See the the documentations for Reporters for details.


Global Options
    -h, --help                view this message
    --types                   show the available types
    --modes                   show the available modes
    -o, --ouput=<path>        path to output directory (required)
                              NB: will overwrite contents of dir
    -c, --cache=<path>        path to cache directory (optional)
                              NB: will overwrite contents of dir
    --skip-log=<level>        don't output the given log level.


Types
    itemset                   sets of items, treated as sets of integers
    digraph                   large directed graphs

    itemset Exmaple
        $ sfp -o /tmp/sfp --support=1000 --samples=10 \
            itemset --min-items=4 --max-items=4  ./data/transactions.dat.gz \
            graple

    itemset Options
        -h, help                 view this message
        -l, loader=<loader-name> the loader to use (default int)
        --min-items=<int>        minimum items in a samplable set
        --max-items=<int>        maximum items in a samplable set

    itemset Loaders
       int                         each line is a transaction
                                   the items are integers
                                   the items are space separated

       int Example file:
            10 1 5 7
            213 2 5 1
            23 1 4 5 7
            3 4 1

    digraph Example
        $ sfp -o /tmp/sfp --support=5 --samples=100 \
            digraph --min-vertices=5 --max-vertices=8 --max-edges=15 \
                ./data/digraph.veg.gz \
            graple

    digraph Options
        -h, help                 view this message
        -l, loader=<loader-name> the loader to use (default veg)
        --min-edges=<int>        minimum edges in a samplable digraph
        --max-edges=<int>        maximum edges in a samplable digraph
        --min-vertices=<int>     minimum vertices in a samplable digraph
        --max-vertices=<int>     maximum vertices in a samplable digraph

    digraph Loaders
        veg File Format
            The veg file format is a line delimited format with vertex lines and
            edge lines. For example:

            vertex	{"id":136,"label":""}
            edge	{"src":23,"targ":25,"label":"ddg"}

            Note: the spaces between vertex and {...} are tabs
            Note: the spaces between edge and {...} are tabs

        veg Grammar
            line -> vertex "\n"
                  | edge "\n"

            vertex -> "vertex" "\t" vertex_json

            edge -> "edge" "\t" edge_json

            vertex_json -> {"id": int, "label": string, ...}
            // other items are optional

            edge_json -> {"src": int, "targ": int, "label": int, ...}
            // other items are  optional


Modes


Reporters
    chain                     chain several reporters together (end the chain
                              with endchain)
    log                       log the samples
    file                      write the samples to a file in the output dir
    dir                       write samples to a nested dir format
    unique                    takes an "inner reporter" but only passes the
                              unique samples to inner reporter. (useful in
                              conjunction with --non-unique)

    log Options
        -l, level=<string>    log level the logger should use
        -p, prefix=<string>   a prefix to put before the log line
        --show-pr             show the selection probability (when applicable)
                              NB: may cause extra (and excessive computation)

    file Options
        -e, embeddings=<name>  the prefix of the name of the file in the output
                               directory to write the embeddings
        -p, patterns=<name>    the prefix of the name of the file in the output
                               directory to write the patterns
        --show-pr              show the selection probability (when applicable)
                               NB: may cause extra (and excessive computation)
        --matrices=<name>      when --show-pr (and the current <mode> supports
                               probabilities) this the name of the file where
                               the pr-matrices will be written. For some modes
                               nothing will be written to this file even when
                               probabilities are computed
        --probabilities=<name> when --show-pr (with <mode> support) the
                               probabilities computed will be written to this
                               file.

        Note: the file extension is chosen by the formatter for the datatype.
              Some data types may provide multiple formatters to choose from
              however that is configured (at this time) from the <type> Options.

        Note: all options are optional. There are default values setup.

    dir Options
        -d, dir-name=<name>   name of the directory.
        --show-pr             show the selection probability (when applicable)
                              NB: may cause extra (and excessive computation)

    unique Options
        --histogram=<name>    if set unique will write the histogram of how many
                              times each node is sampled.

    Examples

        $ sfp -o <path> --samples=5 --support=5 \
            digraph ./digraph.veg.gz \
            graple \
            chain log file

        $ sfp -o <path> --samples=5 --support=5 \
            digraph ./digraph.veg.gz \
            graple \
            chain log chain log log endchain file

        $ sfp --non-unique --skip-log=DEBUG -o /tmp/sfp --samples=5 --support=5 \
            digraph --min-vertices=3 ../fsm/data/expr.gz \
            graple \
            chain \
                log -p non-unique \
                unique \
                    chain \
                        log -p unique \
                        file -e unique-embeddings -p unique-patterns \
                    endchain \
                file -e non-unique-embeddings -p non-unique-patterns
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

func AssertFileOrDirExists(fname string) string {
	fname = path.Clean(fname)
	_, err := os.Stat(fname)
	if err != nil && os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "File '%s' does not exist!\n", fname)
		Usage(ErrorCodes["badfile"])
	} else if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		Usage(ErrorCodes["badfile"])
	}
	return fname
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

func itemsetType(argv []string, conf *config.Config) (lattice.Loader, func(lattice.DataType, lattice.PrFormatter) lattice.Formatter, []string) {
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
	fmtr := func(_ lattice.DataType, prfmt lattice.PrFormatter) lattice.Formatter {
		return &itemset.Formatter{prfmt}
	}
	return loader, fmtr, args
}

func digraphType(argv []string, conf *config.Config) (lattice.Loader, func(lattice.DataType, lattice.PrFormatter) lattice.Formatter, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"hl:", []string{"help", "loader=",
			"support=",
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
	supportedFunc := "min-image"
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
		case "-s", "--support":
			supportedFunc = oa.Arg()
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

	var supported digraph.Supported
	switch supportedFunc {
	case "min-image":
		supported = digraph.MinImgSupported
	case "max-indep":
		supported = digraph.MaxIndepSupported
	default:
		fmt.Fprintf(os.Stderr, "Unknown support function '%v'\n", supportedFunc)
		fmt.Fprintf(os.Stderr, "funcs: min-image, max-indep\n")
		Usage(ErrorCodes["opts"])
	}

	var loader lattice.Loader
	switch loaderType {
	case "veg":
		loader, err = digraph.NewVegLoader(conf, supported, minE, maxE, minV, maxV)
	default:
		fmt.Fprintf(os.Stderr, "Unknown itemset loader '%v'\n", loaderType)
		Usage(ErrorCodes["opts"])
	}
	if err != nil {
		log.Panic(err)
	}
	fmtr := func(dt lattice.DataType, prfmt lattice.PrFormatter) lattice.Formatter {
		g := dt.(*digraph.Graph)
		return digraph.NewFormatter(g, prfmt)
	}
	return loader, fmtr, args
}

func vsigramMode(argv []string, conf *config.Config) (miners.Miner, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"hc",
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
	return vsigram.NewMiner(conf), args
}

type Reporter func(map[string]Reporter, []string, lattice.Formatter, *config.Config)(miners.Reporter, []string)

func logReporter(rptrs map[string]Reporter, argv []string, fmtr lattice.Formatter, conf *config.Config) (miners.Reporter, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"hl:p:",
		[]string{
			"help",
			"level=",
			"prefix=",
			"show-pr",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		Usage(ErrorCodes["opts"])
	}
	level := "INFO"
	prefix := ""
	showPr := false
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			Usage(0)
		case "-l", "--level":
			level = oa.Arg()
		case "-p", "--prefix":
			prefix = oa.Arg()
		case "--show-pr":
			showPr = true
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}
	return reporters.NewLog(fmtr, showPr, level, prefix), args
}

func fileReporter(rptrs map[string]Reporter, argv []string, fmtr lattice.Formatter, conf *config.Config) (miners.Reporter, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"hp:e:",
		[]string{
			"help",
			"patterns=",
			"embeddings=",
			"matrices=",
			"probabilities=",
			"show-pr",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		Usage(ErrorCodes["opts"])
	}
	patterns := "patterns"
	embeddings := "embeddigns"
	matrices := "matrices.json"
	probabilities := "probabilities.prs"
	showPr := false
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			Usage(0)
		case "-p", "--patterns":
			patterns = oa.Arg()
		case "-e", "--embeddings":
			embeddings = oa.Arg()
		case "--matrices":
			matrices = oa.Arg()
		case "--probabilites":
			probabilities = oa.Arg()
		case "--show-pr":
			showPr = true
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}
	fr, err := reporters.NewFile(conf, fmtr, showPr, patterns, embeddings, matrices, probabilities)
	if err != nil {
		fmt.Fprintf(os.Stderr, "There was error creating output files\n")
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
    return fr, args
}

func dirReporter(rptrs map[string]Reporter, argv []string, fmtr lattice.Formatter, conf *config.Config) (miners.Reporter, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"hd:",
		[]string{
			"help",
			"dir-name=",
			"show-pr",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		Usage(ErrorCodes["opts"])
	}
	dir := "samples"
	showPr := false
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			Usage(0)
		case "-d", "--dir-name":
			dir = oa.Arg()
		case "--show-pr":
			showPr = true
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}
	fr, err := reporters.NewDir(conf, fmtr, showPr, dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "There was error creating output files\n")
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
    return fr, args
}

func chainReporter(reports map[string]Reporter, argv []string, fmtr lattice.Formatter, conf *config.Config) (miners.Reporter, []string) {
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
	rptrs := make([]miners.Reporter, 0, 10)
	for len(args) >=1 {
		if args[0] == "endchain" {
			args = args[1:]
			break
		}
		if _, has := reports[args[0]]; !has {
			fmt.Fprintf(os.Stderr, "Unknown reporter '%v'\n", args[0])
			fmt.Fprintln(os.Stderr, "Reporters:")
			for k := range reports {
				fmt.Fprintln(os.Stderr, "  ", k)
			}
			Usage(ErrorCodes["opts"])
		}
		var rptr miners.Reporter
		rptr, args = reports[args[0]](reports, args[1:], fmtr, conf)
		rptrs = append(rptrs, rptr)
	}
	if len(rptrs) == 0 {
		fmt.Fprintln(os.Stderr, "Empty chain")
		fmt.Fprintln(os.Stderr, "try: chain log file")
			Usage(ErrorCodes["opts"])
	}
	return &reporters.Chain{rptrs}, args
}

func uniqueReporter(reports map[string]Reporter, argv []string, fmtr lattice.Formatter, conf *config.Config) (miners.Reporter, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"h",
		[]string{
			"help",
			"histogram=",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		Usage(ErrorCodes["opts"])
	}
	histogram := ""
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			Usage(0)
		case "--histogram":
			histogram = oa.Arg()
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}
	var rptr miners.Reporter
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "You must supply an inner reporter to unique")
		fmt.Fprintln(os.Stderr, "try: unique file")
        Usage(ErrorCodes["opts"])
	} else if _, has := reports[args[0]]; !has {
		fmt.Fprintf(os.Stderr, "Unknown reporter '%v'\n", args[0])
		fmt.Fprintln(os.Stderr, "Reporters:")
		for k := range reports {
			fmt.Fprintln(os.Stderr, "  ", k)
		}
		Usage(ErrorCodes["opts"])
	} else {
		rptr, args = reports[args[0]](reports, args[1:], fmtr, conf)
	}
	uniq, err := reporters.NewUnique(conf, fmtr, rptr, histogram)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating unique reporter '%v'\n", err)
		Usage(ErrorCodes["opts"])
	}
	return uniq, args
}

func main() {

	modes := map[string]func([]string, *config.Config)(miners.Miner, []string) {
		"vsigram": vsigramMode,
	}

	types := map[string]func([]string, *config.Config) (lattice.Loader, func(lattice.DataType,lattice.PrFormatter) lattice.Formatter, []string) {
		"itemset": itemsetType,
		"digraph": digraphType,
	}

	reports := map[string]Reporter {
		"log": logReporter,
		"file": fileReporter,
		"dir": dirReporter,
		"chain": chainReporter,
		"unique": uniqueReporter,
	}

	args, optargs, err := getopt.GetOpt(
		os.Args[1:],
		"ho:c:",
		[]string{
			"help",
			"output=", "cache=",
			"support=",
			"modes", "types", "reporters",
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
		case "--types":
			fmt.Fprintln(os.Stderr, "Types:")
			for k := range types {
				fmt.Fprintln(os.Stderr, "  ", k)
			}
			os.Exit(0)
		case "--modes":
			fmt.Fprintln(os.Stderr, "Modes:")
			for k := range modes {
				fmt.Fprintln(os.Stderr, "  ", k)
			}
			os.Exit(0)
		case "--reporters":
			fmt.Fprintln(os.Stderr, "Reporters:")
			for k := range reports {
				fmt.Fprintln(os.Stderr, "  ", k)
			}
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

	if output == "" {
		fmt.Fprintf(os.Stderr, "You must supply an output dir (-o)\n")
		Usage(ErrorCodes["opts"])
	}

	conf := &config.Config{
		Cache:   cache,
		Output:  output,
		Support: support,
	}

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "You must supply a type and a mode\n")
		Usage(ErrorCodes["opts"])
	} else if _, has := types[args[0]]; !has {
		fmt.Fprintf(os.Stderr, "Unknown data type '%v'\n", args[0])
		fmt.Fprintln(os.Stderr, "Types:")
		for k := range types {
			fmt.Fprintln(os.Stderr, "  ", k)
		}
		Usage(ErrorCodes["opts"])
	}
	loader, makeFmtr, args := types[args[0]](args[1:], conf)

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "You must supply exactly an input path\n")
		fmt.Fprintf(os.Stderr, "You gave: %v\n", args)
		Usage(ErrorCodes["opts"])
	}
	inputPath := AssertFileOrDirExists(args[0])
	args = args[1:]

	getInput := func() (io.Reader, func()) {
		return Input(inputPath)
	}

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "You must supply a mode\n")
		Usage(ErrorCodes["opts"])
	} else if _, has := modes[args[0]]; !has {
		fmt.Fprintf(os.Stderr, "Unknown mining mode '%v'\n", args[0])
		fmt.Fprintln(os.Stderr, "Modes:")
		for k := range modes {
			fmt.Fprintln(os.Stderr, "  ", k)
		}
		Usage(ErrorCodes["opts"])
	}
	mode, args := modes[args[0]](args[1:], conf)

	errors.Logf("INFO", "Got configuration about to load dataset")
	dt, err := loader.Load(getInput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "There was error during the loading process\n")
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fmtr := makeFmtr(dt, mode.PrFormatter())

	var rptr miners.Reporter
	if len(args) == 0 {
		rptr, _ = reports["chain"](reports, []string{"log", "file"}, fmtr, conf)
	} else if _, has := reports[args[0]]; !has {
		fmt.Fprintf(os.Stderr, "Unknown reporter '%v'\n", args[0])
		fmt.Fprintln(os.Stderr, "Reporters:")
		for k := range reports {
			fmt.Fprintln(os.Stderr, "  ", k)
		}
		Usage(ErrorCodes["opts"])
	} else {
		rptr, args = reports[args[0]](reports, args[1:], fmtr, conf)
	}

	if len(args) != 0 {
		fmt.Fprintf(os.Stderr, "unconsumed commandline options: '%v'\n", strings.Join(args, " "))
		Usage(ErrorCodes["opts"])
	}

	errors.Logf("INFO", "loaded data, about to start mining")
	mineErr := mode.Mine(dt, rptr, fmtr)

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

