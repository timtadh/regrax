package cmd

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
	"os/signal"
	"path"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"syscall"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/getopt"
)

import (
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/lattice"
	"github.com/timtadh/sfp/miners"
	"github.com/timtadh/sfp/reporters"
	"github.com/timtadh/sfp/types/digraph"
	"github.com/timtadh/sfp/types/digraph/subgraph"
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
	"usage":    0,
	"version":  2,
	"opts":     3,
	"badfloat": 6,
	"badint":   5,
	"baddir":   6,
	"badfile":  7,
}

var UsageMessage string
var ExtendedMessage string

var CommonUsage string = TypesUsage + ReportersUsage

var TypesUsage string = `
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
        -l, loader=<loader-name> the loader to use (default: veg)
        -c, count-mode=<cmode>   strategy for support counting
                                 (default: MNI minimum image support)
        --extend-from-freq-edges (see below)
        --extend-from-embeddings (see below) (the default)
        --unsup-embs-pruning     (see below)
        --overlap-pruning        (see below)
        --extension-pruning      (see below)
        --no-caching             do not cache any lattice nodes.
        --min-edges=<int>        minimum edges in a samplable digraph
        --max-edges=<int>        maximum edges in a samplable digraph
        --min-vertices=<int>     minimum vertices in a samplable digraph
        --max-vertices=<int>     maximum vertices in a samplable digraph
        -i, --include=<regex>    regex specifying what nodes and edges should
                                 be included based on their label.
        -e, --exclude=<regex>    regex specifying what nodes and edges should
                                 be excluded based on their label.

        Note on inclusion and exclusion of nodes/edges by regexs:

          The include directives are processed before exclude directives. If
          no includes are specified then all labels are included by default.
          If no excludes are specified then no labels are excluded by default.

          You can specify both (-i,--include) and (-e,--exclude) multiple
          times. For example:

            $ digraph -i '^github\.com/timtadh' -i '^$' -e sfp -e fs2

          Would result in the following regular expressions

            include: (^github\.com/timtadh)|(^$)
            exclude: (sfp)|(fs2)

    digraph Support Counting Modes

        Digraph support is usually counted using the Minimum Image Support (MNI)
        [1] which satisifies the Downward Closure Property (DCP). Support
        counting modes which satisfy DCP are called sound those that do not are
        unsound. If a counting mode satisfies DCP on some but not all mining
        sequences it is partially unsound.

        [1] B. Bringmann and S. Nijssen, “What is frequent in a single graph?,”
            in Lecture Notes in Computer Science (including subseries Lecture
            Notes in Artificial Intelligence and Lecture Notes in
            Bioinformatics), 2008, vol.  5012 LNAI, pp. 858–863.

        MNI (Minimum Image)      For the full definition see the Bringmann
                                 paper. Intuitively, the support of a subgraph
                                 is the minimum number of embeddings a
                                 particular vertex of the subgraph has. This
                                 allows fully automorphic to rotations to
                                 count towards the support of the subgraph.

        FIS (Fully Indep.)       Fully independent subgraphs requires that each
                                 disconnected component of the *embedding graph*
                                 is counted once towards support. FIS is a
                                 partially unsound method of counting support.
                                 It is sound when mining using every extension
                                 path (e.g. for: DFS, QSPLOR, GRAPLE, and
                                 FASTMAX) but unsound when only the canonical
                                 paths are used (e.g. for: VSIGRAM, UNIPROX).

        GIS (Greedy Indep.)      Greedy independent subgraphs is a greedy
                                 approximation of FIS. It optimistically prunes
                                 parts of the embedding search tree if only of
                                 the vertex emeddings in the current search
                                 branch has been seen previously. For long
                                 overlapping embedding chains it will return a
                                 higher support number than FIS but is otherwise
                                 equivalent. GIS is an unsound counting mode.

        Notes on support:

            Most of the time the best support option to use is MNI and it is the
            default. However, MNI is an inefficient choice when mining graphs
            which contain frequent subgraphs with many automorphisms. In those
            cases it is more appropriate to use FIS. However, depending on the
            number of automorphic rotations FIS may be too slow as it still
            needs to find all of them. If this is the case, one should use GIS.
            GIS will skip sections of the search tree which FIS must explore at
            the cost of reporting higher support for long embedding chains such
            as following chain:

                pattern:    x -- o

                graph:      o -- x -- o -- x -- o -- x -- o -- x -- o

                FIS support: 1
                GIS support: 4
                MNI support: 4

            FIS is a partially unsound support counting metric. Here is an
            example where it will violate downward closure. Downward closure
            states that subgraphs of the of a frequent subgraph must have
            support greater or equal to the frequent subgraph.

                             1    2    3    4    5    6    7
                graph:       z -- o -- x -- o -- x -- o -- z

                pattern 1:   o -- x
                pattern 2:   z -- o -- x

                embs of 1:   o -- x
                  MNI: 2     2    3
                  GIS: 2     4    3
                  FIS: 1     4    5
                             6    5

                embs of 2:   z -- o -- x
                  MNI: 2     1    2    3
                  GIS: 2     7    6    5
                  FIS: 2
                       ^
                       violation of DCP for FIS


    digraph Candidate Extention Generation Options

        Candidate extentions are potential subgraphs for the graph being mined.
        These options control the method for generating candidates. There is no
        "one-size-fits-all" method.

        --extend-from-embeddings Compute candidate extensions from the
                                 embeddings of the current subgraph. This
                                 extension method is best for mining with low
                                 support values. When using minimum support is
                                 higher than the number of frequent edges in the
                                 mined graph using --extend-from-freq-edges is
                                 better.

        --extend-from-freq-edges Compute candidate extensions from frequent
                                 edges in the graph being mined. This may
                                 compute extensions which are not subgraphs of
                                 the mined graph (spurious candidates). However,
                                 if the number of frequent labels is very low
                                 (in comparison to the embedding frequency) it
                                 may be more efficient than extending from
                                 embeddings.


    digraph Pruning Options

        --unsup-embs-pruning Prune the embedding search by excluding embedding
                             points for subgraph vertices which were proven by a
                             parent subgraph to be invalid.  It is safe to use
                             with all support counting options. It is a much
                             more conservative pruning strategy than overlap
                             pruning (below).  It is unhelpful to use both this
                             option and overlap-pruning as overlap pruning will
                             prune everything that unsupported embedding points
                             pruning will prune.

        --overlap-pruning    Prune the embedding search by only looking for
                             embeddings when fully overlap the parent subgraph
                             of the currently being explored supergraph.  It is
                             safe to use with sound support counting options
                             (such as MNI) when candidate extensions are
                             computed from the embeddings.  However, for other
                             support counting modes it may cause some embeddings
                             to not be discovered as it prunes potenial
                             embeddings of the current node based on the overlap
                             of the embeddings of the parent node.  Since not
                             all rotations of the parent are included in the
                             overlap for FIS and GIS some nodes may be
                             spuriously unsupported. For some datasets, with
                             high amounts of automorphism you may want to uses
                             this flag in conjuction with "optimistic-pruning"
                             to get the best performance (at the cost of
                             completeness).

                             NOTE: Overlap pruning is unsuitable for use with
                                --extend-from-freq-edges as this will mode will
                                terminate embedding search early when sufficient
                                support has been found.

        --extension-pruning  Prune potential extensions by removing extensions
                             which had no support in ancestor nodes. This is a
                             safe mode to use with sound support counting
                             options. With unsound counting modes it may cause
                             the miner to miss frequent subgraphs which have
                             subgraphs with less support (this can only happen
                             when DCP is violated). It may cause a high amount
                             of file IO depending on the mining mode used.  You
                             can use --no-caching to turn off the caching layer.
                             Turning off caching is only recommended when mining
                             all subgraphs (and then it is encouraged).

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

`

var ReportersUsage string = `
Reporters
    chain                     chain several reporters together (end the chain
                                with endchain)
    max                       only write maximal patterns
    canon-max                 only write patterns that are leaf nodes of the
                                canonical-edge frequent pattern tree
    skip                      skip a specified (-s) number of patterns between
                                each reported pattern
    log                       log the samples
    file                      write the samples to a file in the output dir
    dir                       write samples to a nested dir format
    count                     write the count of samples to a file
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
        -n, names=<name>       the name of the file in the output directory to
                               write the pattern names
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

    count Options
        -f, --filename=<name> name of the file to write the count.
                              (default: count)

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

        $ sfp -o <path> --samples=5 --support=5 \
            digraph ./digraph.veg.gz \
            graple \
            chain log -p all max chain log -p max file

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
		fmt.Fprintln(os.Stdout, CommonUsage)
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

func AssertFileExists(fname string) string {
	fname = path.Clean(fname)
	fi, err := os.Stat(fname)
	if err != nil && os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "File '%s' does not exist!\n", fname)
		Usage(ErrorCodes["badfile"])
	} else if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		Usage(ErrorCodes["badfile"])
	} else if fi.IsDir() {
		fmt.Fprintf(os.Stderr, "Passed in file was a directory, %s\n", fname)
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
		fmt.Fprintf(os.Stderr, "Passed in file was a directory, %s\n", fname)
		Usage(ErrorCodes["badfile"])
	}
	return fname
}

func AssertRegex(pat string) string {
	_, err := regexp.Compile(pat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "String '%v' is not a valid regex\n", pat)
		fmt.Fprintf(os.Stderr, "Compile error: %v\n", err)
		Usage(ErrorCodes["opts"])
	}
	return pat
}

func CPUProfile(cpuProfile string) func() {
	errors.Logf("DEBUG", "starting cpu profile: %v", cpuProfile)
	f, err := os.Create(cpuProfile)
	if err != nil {
		log.Fatal(err)
	}
	err = pprof.StartCPUProfile(f)
	if err != nil {
		log.Fatal(err)
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig:=<-sigs
		errors.Logf("DEBUG", "closing cpu profile")
		pprof.StopCPUProfile()
		err := f.Close()
		errors.Logf("DEBUG", "closed cpu profile, err: %v", err)
		panic(errors.Errorf("caught signal: %v", sig))
	}()
	return func() {
		errors.Logf("DEBUG", "closing cpu profile")
		pprof.StopCPUProfile()
		err := f.Close()
		errors.Logf("DEBUG", "closed cpu profile, err: %v", err)
	}
}

type Type func([]string, *config.Config) (lattice.Loader, func(lattice.DataType, lattice.PrFormatter) lattice.Formatter, []string)

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
		"hl:c:i:e:",
		[]string{"help",
			"loader=",
			"count-mode=",
			"fully-optimistic",
			"overlap-pruning",
			"extension-pruning",
			"unsup-embs-pruning",
			"extend-from-embeddings",
			"extend-from-freq-edges",
			"no-caching",
			"emb-search-starting-point=",
			"min-edges=",
			"max-edges=",
			"min-vertices=",
			"max-vertices=",
			"include=",
			"exclude=",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		Usage(ErrorCodes["opts"])
	}

	loaderType := "veg"
	modeStr := "MNI"
	overlapPruning := false
	extensionPruning := false
	unsupEmbsPruning := false
	extendFromEmb := false
	extendFromEdges := false
	embSearchStartingPoint := subgraph.MostConnected
	caching := true
	minE := 0
	maxE := int(math.MaxInt32)
	minV := 0
	maxV := int(math.MaxInt32)
	includes := make([]string, 0, 10)
	excludes := make([]string, 0, 10)
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			Usage(0)
		case "-l", "--loader":
			loaderType = oa.Arg()
		case "-c", "--count-mode":
			modeStr = oa.Arg()
		case "--overlap-pruning":
			overlapPruning = true
		case "--extension-pruning":
			extensionPruning = true
		case "--unsup-embs-pruning":
			unsupEmbsPruning = true
		case "--emb-search-starting-point":
			switch oa.Arg() {
			case "random-start":
				embSearchStartingPoint = subgraph.RandomStart
			case "most-connected":
				embSearchStartingPoint = subgraph.MostConnected
			case "least-connected":
				embSearchStartingPoint = subgraph.LeastConnected
			case "most-frequent":
				embSearchStartingPoint = subgraph.MostFrequent
			case "least-frequent":
				embSearchStartingPoint = subgraph.LeastFrequent
			case "most-extensions":
				embSearchStartingPoint = subgraph.MostExtensions
			case "fewest-extensions":
				embSearchStartingPoint = subgraph.FewestExtensions
			case "lowest-cardinality":
				embSearchStartingPoint = subgraph.LowestCardinality
			case "highest-cardinality":
				embSearchStartingPoint = subgraph.HighestCardinality
			default:
				fmt.Fprintf(os.Stderr, "unknown mode for --emb-search-starting-point %v", oa.Arg())
				fmt.Fprintln(os.Stderr, "valid modes: random-start, (most|least)-connected, (most|least)-frequent")
				fmt.Fprintln(os.Stderr, "             (most|fewest)-extensions, (lowest|highest)-cardinality")
				Usage(ErrorCodes["opts"])
			}
		case "--no-caching":
			caching = false
		case "--min-edges":
			minE = ParseInt(oa.Arg())
		case "--max-edges":
			maxE = ParseInt(oa.Arg())
		case "--min-vertices":
			minV = ParseInt(oa.Arg())
		case "--max-vertices":
			maxV = ParseInt(oa.Arg())
		case "--extend-from-embeddings":
			extendFromEmb = true
		case "--extend-from-freq-edges":
			extendFromEdges = true
		case "-i", "--include":
			includes = append(includes, "(" + AssertRegex(oa.Arg()) + ")")
		case "-e", "--exclude":
			excludes = append(excludes, "(" + AssertRegex(oa.Arg()) + ")")
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}

	var mode digraph.Mode
	if extendFromEmb && extendFromEdges {
		fmt.Fprintf(os.Stderr, "Cannot have both --extend-from-embeddings and --extend-from-freq-edges\n")
		Usage(ErrorCodes["opts"])
	} else if extendFromEmb {
		mode |= digraph.ExtFromEmb
	} else if extendFromEdges {
		mode |= digraph.ExtFromFreqEdges
	} else {
		mode |= digraph.ExtFromEmb
	}

	switch modeStr {
	case "MNI":
		mode |= digraph.MNI
	case "FIS":
		mode |= digraph.FIS
	case "GIS":
		mode |= digraph.GIS
	default:
		fmt.Fprintf(os.Stderr, "Unknown support mode '%v'\n", modeStr)
		fmt.Fprintf(os.Stderr, "support modes: MNI (min-image support), FIS (fully independent subgraphs)\n")
		fmt.Fprintf(os.Stderr, "               GIS (greedy independent subgraphs)\n")
		Usage(ErrorCodes["opts"])
	}
	if overlapPruning {
		mode |= digraph.OverlapPruning
	}
	if extensionPruning {
		mode |= digraph.ExtensionPruning
	}
	if unsupEmbsPruning {
		mode |= digraph.EmbeddingPruning
	}
	if caching {
		mode |= digraph.Caching
	}

	var include *regexp.Regexp = nil
	var exclude *regexp.Regexp = nil
	if len(includes) > 0 {
		include = regexp.MustCompile(strings.Join(includes, "|"))
		errors.Logf("INFO", "including labels matching '%v'", include)
	}
	if len(excludes) > 0 {
		exclude = regexp.MustCompile(strings.Join(excludes, "|"))
		errors.Logf("INFO", "excluding labels matching '%v'", exclude)
	}

	dc := &digraph.Config{
		MinEdges: minE,
		MaxEdges: maxE,
		MinVertices: minV,
		MaxVertices: maxV,
		Mode: mode,
		Include: include,
		Exclude: exclude,
		EmbSearchStartPoint: embSearchStartingPoint,
	}

	var loader lattice.Loader
	switch loaderType {
	case "veg":
		loader, err = digraph.NewVegLoader(conf, dc)
	case "dot":
		loader, err = digraph.NewDotLoader(conf, dc)
	default:
		fmt.Fprintf(os.Stderr, "Unknown itemset loader '%v'\n", loaderType)
		Usage(ErrorCodes["opts"])
	}
	if err != nil {
		log.Panic(err)
	}
	fmtr := func(dt lattice.DataType, prfmt lattice.PrFormatter) lattice.Formatter {
		g := dt.(*digraph.Digraph)
		return digraph.NewFormatter(g, prfmt)
	}
	return loader, fmtr, args
}

type Reporter func(map[string]Reporter, []string, lattice.Formatter, *config.Config) (miners.Reporter, []string)

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
		"hp:e:n:",
		[]string{
			"help",
			"patterns=",
			"embeddings=",
			"names=",
			"matrices=",
			"probabilities=",
			"show-pr",
		},
	)
	if err != nil {
		errors.Logf("ERROR", "%v", err)
		Usage(ErrorCodes["opts"])
	}
	patterns := "patterns"
	embeddings := "embeddings"
	names := "names.txt"
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
		case "-n", "--names":
			names = oa.Arg()
		case "--matrices":
			matrices = oa.Arg()
		case "--probabilites":
			probabilities = oa.Arg()
		case "--show-pr":
			showPr = true
		default:
			errors.Logf("ERROR", "Unknown flag '%v'\n", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}
	fr, err := reporters.NewFile(conf, fmtr, showPr, patterns, embeddings, names, matrices, probabilities)
	if err != nil {
		errors.Logf("ERROR", "There was error creating output files\n")
		errors.Logf("ERROR", "%v\n", err)
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
		errors.Logf("ERROR", "%v", err)
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
			errors.Logf("ERROR", "Unknown flag '%v'\n", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}
	fr, err := reporters.NewDir(conf, fmtr, showPr, dir)
	if err != nil {
		errors.Logf("ERROR", "There was error creating output files\n")
		errors.Logf("ERROR", "%v", err)
		os.Exit(1)
	}
	return fr, args
}

func countReporter(rptrs map[string]Reporter, argv []string, fmtr lattice.Formatter, conf *config.Config) (miners.Reporter, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"hf:",
		[]string{
			"help",
            "filename=",
		},
	)
	if err != nil {
		errors.Logf("ERROR", "%v", err)
		Usage(ErrorCodes["opts"])
	}
	filename := "count"
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			Usage(0)
        case "-f", "--filename":
            filename = oa.Arg()
		default:
			errors.Logf("ERROR", "Unknown flag '%v'\n", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}
	r, err := reporters.NewCount(conf, filename)
	if err != nil {
		errors.Logf("ERROR", "There was error creating output files\n")
		errors.Logf("ERROR", "%v", err)
		os.Exit(1)
	}
	return r, args
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
		errors.Logf("ERROR", "%v", err)
		Usage(ErrorCodes["opts"])
	}
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			Usage(0)
		default:
			errors.Logf("ERROR", "Unknown flag '%v'", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}
	rptrs := make([]miners.Reporter, 0, 10)
	for len(args) >= 1 {
		if args[0] == "endchain" {
			args = args[1:]
			break
		}
		if _, has := reports[args[0]]; !has {
			errors.Logf("ERROR", "Unknown reporter '%v'\n", args[0])
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
		errors.Logf("ERROR", "Empty chain")
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
		errors.Logf("ERROR", "%v", err)
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
			errors.Logf("ERROR", "Unknown flag '%v'", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}
	var rptr miners.Reporter
	if len(args) == 0 {
		errors.Logf("ERROR", "You must supply an inner reporter to unique")
		fmt.Fprintln(os.Stderr, "try: unique file")
		Usage(ErrorCodes["opts"])
	} else if _, has := reports[args[0]]; !has {
		errors.Logf("ERROR", "Unknown reporter '%v'", args[0])
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
		errors.Logf("ERROR", "Error creating unique reporter '%v'\n", err)
		Usage(ErrorCodes["opts"])
	}
	return uniq, args
}

func maxReporter(reports map[string]Reporter, argv []string, fmtr lattice.Formatter, conf *config.Config) (miners.Reporter, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"h",
		[]string{
			"help",
		},
	)
	if err != nil {
		errors.Logf("ERROR", "%v", err)
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
	var rptr miners.Reporter
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "You must supply an inner reporter to max")
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
	m, err := reporters.NewMax(rptr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating max reporter '%v'\n", err)
		Usage(ErrorCodes["opts"])
	}
	return m, args
}

func canonMaxReporter(reports map[string]Reporter, argv []string, fmtr lattice.Formatter, conf *config.Config) (miners.Reporter, []string) {
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
	var rptr miners.Reporter
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "You must supply an inner reporter to canon-max")
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
	m, err := reporters.NewCanonMax(rptr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating canon-max reporter '%v'\n", err)
		Usage(ErrorCodes["opts"])
	}
	return m, args
}

func skipReporter(reports map[string]Reporter, argv []string, fmtr lattice.Formatter, conf *config.Config) (miners.Reporter, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"hs:",
		[]string{
			"help",
			"skip=",
		},
	)
	if err != nil {
		errors.Logf("ERROR", "%v", err)
		Usage(ErrorCodes["opts"])
	}
	skip := 0
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			Usage(0)
		case "-s", "--skip":
			skip = ParseInt(oa.Arg())
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}
	var rptr miners.Reporter
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "You must supply an inner reporter to skip")
		fmt.Fprintln(os.Stderr, "try: skip log")
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
	r := reporters.NewSkip(skip, rptr)
	return r, args
}

func dbscanReporter(rptrs map[string]Reporter, argv []string, fmtr lattice.Formatter, conf *config.Config) (miners.Reporter, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"hf:e:a:",
		[]string{
			"help",
			"filename=",
			"epsilon=",
			"attr=",
		},
	)
	if err != nil {
		errors.Logf("ERROR", "%v", err)
		Usage(ErrorCodes["opts"])
	}
	filename := "clusters"
	attr := ""
	epsilon := 0.2
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			Usage(0)
		case "-f", "--filename":
			filename = oa.Arg()
		case "-a", "--attr":
			attr = oa.Arg()
		case "-e", "--epsilon":
			epsilon = ParseFloat(oa.Arg())
		default:
			errors.Logf("ERROR", "Unknown flag '%v'\n", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}
	if attr == "" {
		errors.Logf("ERROR", "You must supply --attr=<attr> to dbscan")
		Usage(ErrorCodes["opts"])
	}
	r, err := reporters.NewDbScan(conf, fmtr, filename, attr, epsilon)
	if err != nil {
		errors.Logf("ERROR", "There was error creating output files\n")
		errors.Logf("ERROR", "%v", err)
		os.Exit(1)
	}
	return r, args
}

func heapProfileReporter(rptrs map[string]Reporter, argv []string, fmtr lattice.Formatter, conf *config.Config) (miners.Reporter, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"hp:a:e:",
		[]string{
			"help",
			"profile=",
			"after=",
			"every=",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		Usage(ErrorCodes["opts"])
	}
	after := 0
	every := 1
	profile := ""
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			Usage(0)
		case "-p", "--patterns":
			profile = oa.Arg()
		case "-a", "--after":
			after = ParseInt(oa.Arg())
		case "-e", "--every":
			every = ParseInt(oa.Arg())
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			Usage(ErrorCodes["opts"])
		}
	}
	if profile == "" {
		fmt.Fprintf(os.Stderr, "You must supply a location to write the profile (-p) in heap-profile.\n")
		os.Exit(1)
	}
	r, err := reporters.NewHeapProfile(profile, after, every)
	if err != nil {
		fmt.Fprintf(os.Stderr, "There was error creating output files\n")
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	return r, args
}

var Types map[string]Type = map[string]Type{
	"itemset": itemsetType,
	"digraph": digraphType,
}

var Reporters map[string]Reporter = map[string]Reporter{
	"log":          logReporter,
	"file":         fileReporter,
	"dir":          dirReporter,
	"count":        countReporter,
	"chain":        chainReporter,
	"unique":       uniqueReporter,
	"max":          maxReporter,
	"canon-max":    canonMaxReporter,
	"skip":         skipReporter,
	"dbscan":       dbscanReporter,
	"heap-profile": heapProfileReporter,
}

type Mode func(argv []string, conf *config.Config) (miners.Miner, []string)

func ParseType(args []string, conf *config.Config) (func(lattice.PrFormatter) (lattice.DataType, lattice.Formatter), []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "You must supply a type\n")
		Usage(ErrorCodes["opts"])
	} else if _, has := Types[args[0]]; !has {
		fmt.Fprintf(os.Stderr, "Unknown data type '%v'\n", args[0])
		fmt.Fprintln(os.Stderr, "Types:")
		for k := range Types {
			fmt.Fprintln(os.Stderr, "  ", k)
		}
		Usage(ErrorCodes["opts"])
	}
	loader, makeFmtr, args := Types[args[0]](args[1:], conf)
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
	f := func(prfmtr lattice.PrFormatter) (lattice.DataType, lattice.Formatter) {
		errors.Logf("INFO", "Got configuration about to load dataset")
		dt, err := loader.Load(getInput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "There was error during the loading process\n")
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		fmtr := makeFmtr(dt, prfmtr)
		return dt, fmtr
	}
	return f, args
}

func ParseMode(args []string, conf *config.Config, modes map[string]Mode) (miners.Miner, []string) {
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
	return modes[args[0]](args[1:], conf)
}

func ParseReporter(args []string, conf *config.Config, fmtr lattice.Formatter) (miners.Reporter, []string) {
	var rptr miners.Reporter
	if len(args) == 0 {
		rptr, _ = Reporters["chain"](Reporters, []string{"log", "file"}, fmtr, conf)
	} else if _, has := Reporters[args[0]]; !has {
		fmt.Fprintf(os.Stderr, "Unknown reporter '%v'\n", args[0])
		fmt.Fprintln(os.Stderr, "Reporters:")
		for k := range Reporters {
			fmt.Fprintln(os.Stderr, "  ", k)
		}
		Usage(ErrorCodes["opts"])
	} else {
		rptr, args = Reporters[args[0]](Reporters, args[1:], fmtr, conf)
	}
	return rptr, args
}

func Run(dt lattice.DataType, fmtr lattice.Formatter, mode miners.Miner, rptr miners.Reporter) int {
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
	return code
}

func Main(args []string, conf *config.Config, modes map[string]Mode) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "You must supply a type and a mode\n")
		Usage(ErrorCodes["opts"])
	}

	errors.Logf("INFO", "args: %v", os.Args)
	loadDt, args := ParseType(args, conf)
	mode, args := ParseMode(args, conf, modes)
	dt, fmtr := loadDt(mode.PrFormatter())
	rptr, args := ParseReporter(args, conf, fmtr)

	if len(args) != 0 {
		fmt.Fprintf(os.Stderr, "unconsumed commandline options: '%v'\n", strings.Join(args, " "))
		Usage(ErrorCodes["opts"])
	}


	return Run(dt, fmtr, mode, rptr)
}
