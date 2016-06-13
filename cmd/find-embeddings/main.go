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
	"bufio"
	"fmt"
	"math"
	"io"
	"os"
	"strings"
	"strconv"
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


func loadProbabilities(path string) (prs []float64, patterns []string, count int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, 0, err
	}
	prs = make([]float64, 0, 10)
	patterns = make([]string, 0, 10)
	seen := make(map[string]bool)
	count = 0
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		fields := strings.SplitN(line, ",", 2)
		pr, err := strconv.ParseFloat(fields[0], 64)
		if err != nil {
			return nil, nil, 0, err
		}
		pattern := strings.TrimSpace(fields[1])
		count++
		if _, has := seen[pattern]; !has {
			seen[pattern] = true
			prs = append(prs, pr)
			patterns = append(patterns, pattern)
		}
	}
	return prs, patterns, count, nil
}

func loadNames(path string) (prs []float64, patterns []string, count int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, 0, err
	}
	prs = make([]float64, 0, 10)
	patterns = make([]string, 0, 10)
	seen := make(map[string]bool)
	count = 0
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		pattern := line
		count++
		if _, has := seen[pattern]; !has {
			seen[pattern] = true
			patterns = append(patterns, pattern)
		}
	}
	return prs, patterns, count, nil
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
			"probabilities=",
			"samples=",
			"names=",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	visual := ""
	patterns := make([]string, 0, 10)
	prPath := ""
	namesPath := ""
	cpuProfile := ""
	samples := -1
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			cmd.Usage(0)
		case "-p", "--pattern":
			patterns = append(patterns, oa.Arg())
		case "--probabilities":
			prPath = cmd.AssertFileExists(oa.Arg())
		case "--names":
			namesPath = cmd.AssertFileExists(oa.Arg())
		case "--samples":
			samples = cmd.ParseInt(oa.Arg())
		case "--cpu-profile":
			cpuProfile = cmd.AssertFile(oa.Arg())
		case "-v", "--visual":
			visual = cmd.AssertFile(oa.Arg())
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			cmd.Usage(cmd.ErrorCodes["opts"])
		}
	}

	if (prPath != "" || namesPath != "") && len(patterns) > 0 {
		fmt.Fprintf(os.Stderr, "You cannot supply patterns with both (-p) and (--probabilities)\n")
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	if len(patterns) == 0 && prPath == "" && namesPath == "" {
		fmt.Fprintf(os.Stderr, "You must supply a pattern (-p, --names, --probabilities)\n")
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	var patternCount int = len(patterns)
	var prs []float64 = nil
	if prPath != "" {
		var err error
		prs, patterns, patternCount, err = loadProbabilities(prPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "There was error loading the probability file\n")
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
	} else if namesPath != "" {
		var err error
		prs, patterns, patternCount, err = loadNames(namesPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "There was error loading the probability file\n")
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
	}

	if samples < len(prs) {
		errors.Errorf("INFO", "assuming # of samples is the total number of patterns supplied: %v", patternCount)
		samples = patternCount
	}

	conf := &config.Config{}

	loadDt, args := cmd.ParseType(args, conf)
	dt, _ := loadDt(nil)
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

	matches := make([]float64, 0, len(patterns))
	matched := make([]*subgraph.SubGraph, 0, len(patterns))
	sgEdges := make([]float64, 0, len(patterns))
	total := 0.0
	totalEdges := 0.0
	for i, pattern := range patterns {
		sg, err := subgraph.ParsePretty(pattern, &graph.G.Colors, graph.G.Labels)
		if err != nil {
			fmt.Fprintf(os.Stderr, "There was error during the parsing the pattern '%v'\n", pattern)
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
		match, csg, err := sg.EstimateMatch(graph.Indices)
		match = match * float64(len(sg.E))
		if err != nil {
			errors.Logf("ERROR", "%v", err)
			return 1
		}
		matches = append(matches, match)
		matched = append(matched, csg)
		sgEdges = append(sgEdges, float64(len(sg.E)))
		fmt.Printf("%v, %v, %v\n", i+1, match, pattern)
		total += match
		totalEdges += float64(len(sg.E))
		if visualize != nil {
			dotty, err := csg.VisualizeEmbedding(sg.AsIndices(graph.Indices))
			if err != nil {
				fmt.Fprintf(os.Stderr, "There was error visualizing the embedding '%v'\n", csg)
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return 1
			}
			fmt.Fprintln(visualize, dotty)
		}
	}
	errors.Logf("DEBUG", "prs %v",  sum(prs))
	fmt.Printf(", %v, sample total covered edges\n", total)
	fmt.Printf(", %v, sample total edges\n", totalEdges)
	fmt.Printf(", %v, sample covered/total\n", total/totalEdges)
	fmt.Printf(", %v, sample avg covered\n", total/float64(len(patterns)))
	fmt.Printf(", %v, sample avg edges\n", totalEdges/float64(len(patterns)))

	if len(prs) > 0 {
		pis := samplingPrs(samples, prs)
		jpis := jointSamplingPrs(samples, prs, pis)
		estN := estPopSize(pis)
		estTotalMatch := estPopTotal(pis, matches)
		estVarTotalMatch := estVarTotal(pis, jpis, matches)
		estTotalEdges := estPopTotal(pis, sgEdges)
		estVarTotalEdges := estVarTotal(pis, jpis,  sgEdges)

		fmt.Printf("\n")
		fmt.Printf(", %v, estimated population total of matched edges\n", estTotalMatch)
		fmt.Printf(", %v, estimated population total of total edges\n", estTotalEdges)
		fmt.Printf(", %v, estimated var population total of match edges\n", estVarTotalMatch)
		fmt.Printf(", %v, estimated var population total of total edges\n", estVarTotalEdges)
		fmt.Printf(", %v, estimated std population total of match edges\n", math.Sqrt(estVarTotalMatch))
		fmt.Printf(", %v, estimated std population total of total edges\n", math.Sqrt(estVarTotalEdges))
		fmt.Printf(", %v, estimated population mean\n", estTotalMatch/estTotalEdges)

		estMeanMatch := estPopMean(estTotalMatch, estN)
		estMeanEdges := estPopMean(estTotalEdges, estN)
		fmt.Printf("\n")
		fmt.Printf(", %v, est. mean matches\n", estMeanMatch)
		fmt.Printf(", %v, est. mean edges\n", estMeanEdges)
		fmt.Printf(", %v, est. cover\n", estMeanMatch/estMeanEdges)

		varMeanMatch := estVarMean(estN, estMeanMatch, pis, jpis, matches)
		varMeanEdges := estVarMean(estN, estMeanEdges, pis, jpis, sgEdges)
		stdMeanMatch := math.Sqrt(varMeanMatch)
		stdMeanEdges := math.Sqrt(varMeanEdges)
		fmt.Printf("\n")
		fmt.Printf(", %v, var. mean matches\n", varMeanMatch)
		fmt.Printf(", %v, var. mean edges\n", varMeanEdges)
		fmt.Printf(", %v, std. mean matches\n", stdMeanMatch)
		fmt.Printf(", %v, std. mean edges\n", stdMeanEdges)

		t := t_alpha_05[samples-1]
		fmt.Printf("\n")
		fmt.Printf(", %v - %v, interval. mean matches\n",
			estMeanMatch - t*stdMeanMatch,
			estMeanMatch + t*stdMeanMatch)
		fmt.Printf(", %v - %v, interval. mean edges\n",
			estMeanEdges - t*stdMeanEdges,
			estMeanEdges + t*stdMeanEdges)
		fmt.Printf(", %v - %v, interval. cover\n",
			math.Max((estMeanMatch - t*stdMeanMatch)/(estMeanEdges + t*stdMeanEdges), 0.0),
			math.Min((estMeanMatch + t*stdMeanMatch)/(estMeanEdges - t*stdMeanEdges), 1.0))
	}
	return 0
}

func sum(xs []float64) float64 {
	s := 0.0
	for _, x := range xs {
		s += x
	}
	return s
}

func estPopSize(pis []float64) float64 {
	estN := 0.0
	for _, pi := range pis {
		estN += 1.0/pi
	}
	return estN
}

func estPopTotal(pis, ys []float64) (estTau float64) {
	for i := range pis {
		estTau += ys[i]/pis[i]
	}
	return estTau
}

func estPopMean(estTau, estN float64) (estMu float64) {
	estMu = estTau/estN
	return estMu
}

func samplingPrs(n int, prs []float64) []float64 {
	pis := make([]float64, 0, len(prs))
	for _, pr := range prs {
		pis = append(pis, 1 - math.Pow(1 - pr, float64(n)))
	}
	return pis
}

func jointSamplingPrs(n int, prs, pis []float64) [][]float64 {
	cJpi := func(i, j int) float64 {
		return pis[i] + pis[j] - (1 - math.Pow(1 - prs[i] - prs[j], float64(n)))
	}
	jpis := make([][]float64, 0, len(pis))
	for i := range pis {
		jpi := make([]float64, 0, len(pis))
		for j := range pis {
			x := cJpi(i, j)
			if x == 0.0 {
				x = 1e-17 // fixup in case of extremely small joint pr
			}
			jpi = append(jpi, x)
		}
		jpis = append(jpis, jpi)
	}
	return jpis
}

func estVarTotal(pis []float64, jpis [][]float64, ys []float64) float64 {
	a := 0.0
	for i, pi := range pis {
		a += ((1 - pi)/math.Pow(pi, 2)) * math.Pow(ys[i], 2)
	}
	b := 0.0
	for i := range pis {
		for j := i + 1; j < len(pis); j++ {
			b = ((1/(pis[i]*pis[j])) - (1/jpis[i][j])) * ys[i] * ys[j]
		}
	}
	return a + 2*b
}

func estVarMean(estN, estMean float64, pis []float64, jpis [][]float64, ys []float64) float64 {
	a := 0.0
	for _, pi := range pis {
		a += (1 - pi)/(math.Pow(pi, 2))
	}
	b := 0.0
	for i := range pis {
		for j := range pis {
			if i == j {
				continue
			}
			lhs := ((jpis[i][j] - pis[i]*pis[j])/(pis[i]*pis[j]))
			rhs := (((ys[i] - estMean)*(ys[j] - estMean))/(jpis[i][j]))
			b += lhs * rhs
		}
	}
	scale := 1/math.Pow(estN, 2)
	return scale*(a + b)
}
