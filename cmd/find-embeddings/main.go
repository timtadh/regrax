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
		"h:p:v:",
		[]string{
			"help",
			"pattern=",
			"cpu-profile=",
			"visualize=",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	visual := ""
	patterns := make([]string, 0, 10)
	cpuProfile := ""
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			cmd.Usage(0)
		case "-p", "--pattern":
			patterns = append(patterns, oa.Arg())
		case "--cpu-profile":
			cpuProfile = cmd.AssertFile(oa.Arg())
		case "-v", "--visual":
			visual = cmd.AssertFile(oa.Arg())
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			cmd.Usage(cmd.ErrorCodes["opts"])
		}
	}

	if len(patterns) == 0 {
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

	if cpuProfile != "" {
		defer cmd.CPUProfile(cpuProfile)()
	}

	var visualize io.Writer = nil
	if visual != "" {
		f, err := os.Create(visual)
		if err != nil {
			fmt.Fprintf(os.Stderr, "There was error opening the visualization output file\n")
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
		defer f.Close()
		visualize = f
	}

	total := 0.0
	for i, pattern := range patterns {
		sg, err := subgraph.ParsePretty(pattern, &graph.G.Colors, graph.G.Labels)
		if err != nil {
			fmt.Fprintf(os.Stderr, "There was error during the parsing the pattern '%v'\n", pattern)
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
		match, csg, err := sg.EstimateMatch(graph.Indices)
		if err != nil {
			errors.Logf("ERROR", "%v", err)
			return 1
		}
		fmt.Printf("%v, %v, %v\n", i+1, match, pattern)
		total += match
		if visualize != nil {
			edge := func(eid int, vids map[int]int) int {
				e := &csg.E[eid]
				for eid := range sg.E {
					if vids[e.Src] == sg.E[eid].Src && vids[e.Targ] == sg.E[eid].Targ && e.Color == sg.E[eid].Color {
						return eid
					}
				}
				panic("unreachable")
			}
			found, edgeChain, eid, ids := csg.Embedded(sg.AsIndices(graph.Indices))
			if !found {
				panic("not found!")
			}
			vids := make(map[int]int)
			vidSet := make(map[int]bool)
			eidSet := make(map[int]bool)
			hv := make(map[int]bool)
			he := make(map[int]bool)
			for c := ids; c != nil; c = c.Prev {
				vids[c.Idx] = c.Id
				vidSet[c.Id] = true
			}
			for vidx := range sg.V {
				if !vidSet[vidx] {
					hv[vidx] = true
				}
			}
			for ceid, eidx := range edgeChain {
				if ceid >= eid {
					continue
				}
				eidSet[edge(eidx, vids)] = true
			}
			for eidx := range sg.E {
				if !eidSet[eidx] {
					he[eidx] = true
				}
			}
			fmt.Fprintln(visualize, sg.Dotty(graph.G.Colors, hv, he))
		}
	}
	fmt.Printf("%v, %v, total\n", "-", total/float64(len(patterns)))
	return 0
}
