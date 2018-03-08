package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/timtadh/getopt"
	"github.com/timtadh/regrax/cmd"
	"github.com/timtadh/regrax/mine"
	"github.com/timtadh/regrax/sample"
)

func init() {
	cmd.UsageMessage = fmt.Sprintf("%v --help", os.Args[0])
	cmd.ExtendedMessage = `regrax - the REcurring GRAph eXtractor

Commands

    mine     extract (mine) all frequent subgraphs
    sample   randomly sample a frequent subgraphs


mine - find all frequent patterns

    $ regrax mine -o <path> [Global Options] \
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
        -p, --parallelism=<int>   Parallelism level to use. Defaults to
                                  the number of CPU cores you have. Set to
                                  0 to turn off parallelism.
        --support=<int>           minimum support of patterns (required)
        --skip-log=<level>        don't output the given log level.

    Developer Options
        --cpu-profile=<path>      write a cpu-profile to this location

        heap-profile Reporter

            $ afp ... <type> ... <mode> ... chain ... heap-profile [options]

            -p, profile=<path>    where you want the heap-profile written
            -e, every=<int>       collect every n samples collected (default 1)
            -a, after=<int>       collect after n samples collected (default 0)

    Modes
        dfs                       depth first search of the lattice
        vsigram                   dfs but only on the canonical edges


sample - sample frequent patterns

    $ regrax sample -o <path> --samples=<int> --support=<int> [Global Options] \
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
        -p, --parallelism=<int>   Parallelism level to use. Defaults to
                                  the number of CPU cores you have. Set to
                                  0 to turn off parallelism.
        --samples=<int>           number of samples to collect (required)
        --support=<int>           minimum support of patterns (required)
        --non-unique              by default, regrax collects only unique samples. This
                                  option allows non-unique samples.
        --skip-log=<level>        don't output the given log level.

    Developer Options
        --cpu-profile=<path>      write a cpu-profile to this location

        heap-profile Reporter

            $ regrax ... <type> ... <mode> ... chain ... heap-profile [options]

            -p, profile=<path>    where you want the heap-profile written
            -e, every=<int>       collect every n samples collected (default 1)
            -a, after=<int>       collect after n samples collected (default 0)

    Modes
        graple                    the GRAPLE (unweighted random walk) algorithm.
        musk                      uniform sampling of maximal patterns.
        ospace                    uniform sampling of all patterns.
        fastmax                   faster sampling of large max patterns than
                                  graple.
        premusk                   musk but with random teleports
        uniprox                   approximately uniform sampling of max patterns
                                  using an absorbing chain

        premusk Options
            -t, teleports=<float> the probability of teleporting (default: .01)

        uniprox Options
            -w, walks=<int>       number of estimating walks (default 15)
`
}

func main() {
	code := run()
	if code != 0 {
		os.Exit(code)
	}
}

func run() int {

	args, optargs, err := getopt.GetOpt(
		os.Args[1:],
		"h",
		[]string{
			"help",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "could not process your arguments (perhaps you forgot a mode?) try:")
		fmt.Fprintf(os.Stderr, "$ %v [mine|sample] %v\n", os.Args[0], strings.Join(os.Args[1:], " "))
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

	if len(args) <= 0 {
		fmt.Fprintln(os.Stderr, "Error: no args supplied")
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	switch args[0] {
	case "mine":
		return mine.Run(args[1:])
	case "sample":
		return sample.Run(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown mode %q, supported modes are \"mine\" and \"sample\"\n", args[0])
		cmd.Usage(cmd.ErrorCodes["opts"])
	}
	return 0
}
