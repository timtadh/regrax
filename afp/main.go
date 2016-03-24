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
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/getopt"
)

import (
	"github.com/timtadh/sfp/afp/miners/dfs"
	"github.com/timtadh/sfp/afp/miners/vsigram"
	"github.com/timtadh/sfp/cmd"
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/miners"
)

func init() {
	cmd.UsageMessage = "afp --help"
	cmd.ExtendedMessage = `
afp - sample frequent patterns

$ afp -o <path> [Global Options] \
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

Developer Options
    --cpu-profile=<path>      write a cpu-profile to this location

    heap-profile Reporter

        $ afp ... <type> ... <mode> ... chain ... heap-profile [options]

        -p, profile=<path>    where you want the heap-profile written
        -e, every=<int>       collect every n samples collected (default 1)
        -a, after=<int>       collect after n samples collected (default 0)


Types
    itemset                   sets of items, treated as sets of integers
    digraph                   large directed graphs

    itemset Exmaple
        $ afp -o /tmp/afp --support=1000 --samples=10 \
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
        $ afp -o /tmp/afp --support=5 --samples=100 \
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

    unique Options
        --histogram=<name>    if set unique will write the histogram of how many
                              times each node is sampled.

    Examples

        $ afp -o <path> --samples=5 --support=5 \
            digraph ./digraph.veg.gz \
            graple \
            chain log file

        $ afp -o <path> --samples=5 --support=5 \
            digraph ./digraph.veg.gz \
            graple \
            chain log chain log log endchain file

        $ afp --non-unique --skip-log=DEBUG -o /tmp/afp --samples=5 --support=5 \
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
		cmd.Usage(cmd.ErrorCodes["opts"])
	}
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			cmd.Usage(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			cmd.Usage(cmd.ErrorCodes["opts"])
		}
	}
	return vsigram.NewMiner(conf), args
}

func dfsMode(argv []string, conf *config.Config) (miners.Miner, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"hc",
		[]string{
			"help",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		cmd.Usage(cmd.ErrorCodes["opts"])
	}
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			cmd.Usage(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			cmd.Usage(cmd.ErrorCodes["opts"])
		}
	}
	return dfs.NewMiner(conf), args
}

func main() {
	os.Exit(run())
}

func run() int {
	modes := map[string]cmd.Mode{
		"dfs":     dfsMode,
		"vsigram": vsigramMode,
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
			"cpu-profile=",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "could not process your arguments (perhaps you forgot a mode?) try:")
		fmt.Fprintf(os.Stderr, "$ %v breadth %v\n", os.Args[0], strings.Join(os.Args[1:], " "))
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	output := ""
	cache := ""
	support := 0
	cpuProfile := ""
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			cmd.Usage(0)
		case "-o", "--output":
			output = cmd.EmptyDir(oa.Arg())
		case "-c", "--cache":
			cache = cmd.EmptyDir(oa.Arg())
		case "--support":
			support = cmd.ParseInt(oa.Arg())
		case "--types":
			fmt.Fprintln(os.Stderr, "Types:")
			for k := range cmd.Types {
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
			for k := range cmd.Reporters {
				fmt.Fprintln(os.Stderr, "  ", k)
			}
			os.Exit(0)
		case "--skip-log":
			level := oa.Arg()
			errors.Logf("INFO", "not logging level %v", level)
			errors.SkipLogging[level] = true
		case "--cpu-profile":
			cpuProfile = cmd.AssertFile(oa.Arg())
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			cmd.Usage(cmd.ErrorCodes["opts"])
		}
	}

	if support <= 0 {
		fmt.Fprintf(os.Stderr, "Support <= 0, must be > 0\n")
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	if output == "" {
		fmt.Fprintf(os.Stderr, "You must supply an output dir (-o)\n")
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	if cpuProfile != "" {
		errors.Logf("DEBUG", "starting cpu profile: %v", cpuProfile)
		f, err := os.Create(cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			errors.Logf("DEBUG", "closing cpu profile")
			pprof.StopCPUProfile()
			err := f.Close()
			errors.Logf("DEBUG", "closed cpu profile, err: %v", err)
		}()
	}

	conf := &config.Config{
		Cache:   cache,
		Output:  output,
		Support: support,
	}

	return cmd.Main(args, conf, modes)
}
