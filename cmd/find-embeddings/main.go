package main

/* Tim Henderson (tadh@case.edu)
*
* Copyright (c) 2016, Tim Henderson, Case Western Reserve University
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
	"math"
	"io"
	"os"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/data-structures/exc"
	"github.com/timtadh/getopt"
)

import (
	"github.com/timtadh/sfp/cmd"
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/types/digraph"
	"github.com/timtadh/sfp/types/digraph/subgraph"
)

func init() {
	cmd.UsageMessage = "find-embeddings --help"
	cmd.ExtendedMessage = `
find-embeddings -p <pattern> <graph>
`
}

func main() {
	os.Exit(run())
}

func run() int {
	args, optargs, err := getopt.GetOpt(
		os.Args[1:],
		"h:p:",
		[]string{
			"help",
			"pattern=",
			"cpu-profile=",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	pattern := ""
	cpuProfile := ""
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			cmd.Usage(0)
		case "-p", "--pattern":
			pattern = oa.Arg()
		case "--cpu-profile":
			cpuProfile = cmd.AssertFile(oa.Arg())
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			cmd.Usage(cmd.ErrorCodes["opts"])
		}
	}

	if pattern == "" {
		fmt.Fprintf(os.Stderr, "You must supply a pattern (-p)\n")
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	conf := &config.Config{}

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "You must supply exactly an input path\n")
		fmt.Fprintf(os.Stderr, "You gave: %v\n", args)
		cmd.Usage(cmd.ErrorCodes["opts"])
	}
	inputPath := cmd.AssertFileOrDirExists(args[0])
	getInput := func() (io.Reader, func()) {
		return cmd.Input(inputPath)
	}

	loader, err := digraph.NewDotLoader(conf, digraph.OptimisticPruning, 0, int(math.MaxInt32), 0, int(math.MaxInt32))
	if err != nil {
		fmt.Fprintf(os.Stderr, "There was error constructing the loader\n")
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// errors.Logf("INFO", "Got configuration about to load dataset")
	dt, err := loader.Load(getInput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "There was error during the loading process\n")
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	graph := dt.(*digraph.Digraph)

	// errors.Logf("INFO", "input pattern %v", pattern)
	sg, err := subgraph.ParsePretty(pattern, graph.G.Labels)
	if err != nil {
		fmt.Fprintf(os.Stderr, "There was error during the parsing the pattern '%v'\n", pattern)
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	// errors.Logf("INFO", "loaded pattern %v", sg.Pretty(graph.G.Colors))

	if cpuProfile != "" {
		defer cmd.CPUProfile(cpuProfile)()
	}


	err = exc.Try(func(){
		csg := sg
		for len(csg.E) > 1 {
			found, chain, maxEid, ids := csg.Embedded(graph.Indices)
			if false {
				errors.Logf("INFO", "found: %v %v %v %v", found, chain, maxEid, ids)
			}
			if found {
				fmt.Printf("%v\n", float64(maxEid)/float64(len(sg.E)))
				break
			}
			//FIX
			var b *subgraph.Builder
			connected := false
			eid := maxEid
			for !connected && eid < len(chain) {
				b = csg.Builder().Ctx(func(b *subgraph.Builder) {
					exc.ThrowOnError(b.RemoveEdge(chain[eid]))
				})
				connected = b.Connected()
				eid++
			}
			csg = b.Build()
		}
	}).Error()
	if err != nil {
		errors.Logf("ERROR", "%v", err)
		return 1
	}

	/*
	sup, exts, embs, overlap, err := digraph.ExtsAndEmbs(graph, sg, nil, set.NewSortedSet(0), graph.Mode, true)
	if err != nil {
		errors.Logf("ERROR", "err: %v", err)
		return 2
	}
	errors.Logf("INFO", "pat %v sup %v exts %v embs %v overlap %v", sg, sup, len(exts), len(embs), overlap)
	for _, emb := range embs {
		errors.Logf("EMBEDDING", "emb %v", emb.Pretty(graph.G.Colors))
	}
	*/

	return 0
}
