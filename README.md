# REGRAX - REcurring GRAph eXtractor

by Tim A. D. Henderson (tadh@case.edu)

Licensed under the GPLv3

This is a in progress research project. Parts of this work are discribed in

> **Tim A. D. Henderson**. [Frequent Subgraph Analysis and its Software
> Engineering
> Applications](http://hackthology.com/frequent-subgraph-analysis-and-its-software-engineering-applications.html).
> [Case Western Reserve University](http://case.edu/). Doctoral Dissertation.
> 2017.

However, this work is a combined platform which contains experiments which are
not fully explained in the above dissertation. If you use REGRAX in any way
please

1.  Cite the paper and this repository
2.  Contact me via (tadh@case.edu) to understand what bits you are using and if
    they require special consideration during publication.

# What this is

REGRAX is a platform for high performance frequent subgraph analysis of large
connected graphs. It supports both frequent subgraph mining and probablistic
markov based sampling methods for extracting frequent subgraphs. It does not
(and never will) support discriminative or importance mining.

## What is a frequent subgraph

A frequent subgraph is a graph fragment which recurs at least `n` times in
either a large graph or a database of small graphs. When considering computing
frequency or recurrence in large graphs there are subtle considerations as a
fragment overlap itself through automorphism. See my dissertation for a detailed
discussion.

# Install

1.  Install go <https://golang.org>

2.  Install dep <https://github.com/golang/dep>

3.  Set up the directory structure

        $ mkdir -p regrax/src/github.com/timtadh/
        $ cd regrax

4.  Create a python virtualenv

        $ virtualenv --no-site-packages env

5.  Clone the repo

        $ src/github.com/timtadh
        $ git clone https://github.com/timtadh/regrax

6.  Activate the virtualenv and set the go path (you have to do this everytime
    you want to use regrax):

        $ cd src/github.com/timtadh
        $ source .activate

7.  Run dep (in the repo) to install the go dependencies:

        $ dep ensure

8.  Install the python deps:

        $ pip install -r requirements.txt

9.  Compile the regrax

        $ go install github.com/timtadh/regrax

10. Check to make sure it works

        $ regrax
        Support <= 0, must be > 0
        regrax --help
        Try -h or --help for help

# Usage

```
regrax - the REcurring GRAph eXtractor

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


Types

    itemset                   sets of items, treated as sets of integers
    digraph                   large directed graphs

    itemset Exmaple

        $ regrax sample -o /tmp/sfp --support=1000 --samples=10 \
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

        $ regrax sample -o /tmp/sfp --support=5 --samples=100 \
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

            $ digraph -i '^github\.com/timtadh' -i '^$' -e regrax -e fs2

          Would result in the following regular expressions

            include: (^github\.com/timtadh)|(^$)
            exclude: (regrax)|(fs2)

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

            vertex  {"id":136,"label":""}
            edge    {"src":23,"targ":25,"label":"ddg"}

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

        $ regrax sample -o <path> --samples=5 --support=5 \
            digraph ./digraph.veg.gz \
            graple \
            chain log file

        $ regrax sample -o <path> --samples=5 --support=5 \
            digraph ./digraph.veg.gz \
            graple \
            chain log chain log log endchain file

        $ regrax sample -o <path> --samples=5 --support=5 \
            digraph ./digraph.veg.gz \
            graple \
            chain log -p all max chain log -p max file

        $ regrax sample --non-unique --skip-log=DEBUG -o /tmp/sfp --samples=5 --support=5 \
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
```
