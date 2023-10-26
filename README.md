`fastqSplit`

Program to split a fastq file into separate files per read group, where the read group is defined as `<flow cell ID>.<lane number>`. The flow cell ID and the lane number are extracted from each fastq record header line.

The default parameters assume that your fastq file has headers in the [Illumina format](https://en.wikipedia.org/wiki/FASTQ_format#Illumina_sequence_identifiers); example:

```
@EAS139:136:FC706VJ:2:2104:15343:197393 1:N:18:1
```

In this example, the flow cell ID is `FC706VJ` and the lane ID is `2`.

# Usage

Example usage;

```
fastqSplit data/test1.fastq
```

Will output files `XYZ12345.1.fastq` and `XYZ12345.2.fastq`

If you fastq file is .gz compressed, you should pipe it in with `zcat` or `gunzip`;

```
gunzip -c data/test1.fastq.gz | ./fastqSplit
```

## Options

Note that `fastqSplit` can read from .gz compressed files natively, but the .gz archive decompression, and total execution time, will be much faster if handled in a separate process and transmitted over stdin with a pipe as shown in the above examples. However if you wish to use `fastqSplit`'s built-in .gz decompression, you should add the `-p` arg in order to utilize background buffered file decompression & scanning for a performance boost.

Other command line options for `fastqSplit` include;

```
  -b int
    	read buffer size (number of lines) when using parallel read method (default 10000)
  -delim string
    	delimiter character for the fastq header fields (default ":")
  -fcIndexPos int
    	field number for the flowcell ID in the header (default 2)
  -laneIndexPos int
    	field number for the lane ID in the header (default 3)
  -p	read input on a separate thread (parallel)
  -rgJoinChar string
    	character used to join the flowcell and lane IDs to create the read group ID (default ".")
```

Note that the "field number" for flowcell ID and lane ID is 0-based, so in the header example `@EAS139:136:FC706VJ:2`, the value `FC706VJ` is position 2 and value `2` is in position 3.


# Installation

You can download a pre-built binary from the Releases page.

You can get it from Docker Hub here: (coming soon)

You can build it from source with Go version 1.20+

```
go build -o ./fastqSplit ./main.go
```

# Notes

This program is an alternative to the equivalent `awk` implementation that would look like this;

```bash
gunzip -c ${fastq} | awk 'BEGIN{FS="[@:]"}{out=$4"."$5".fastq"; print > out; for (i = 1; i <= 3; i++) {getline ; print > out}}'
```

`fastqSplit` is significantly faster than the `awk` implementation.

## Comparison

Using a 3.5GB size fastq.gz file with 268093288 lines (67023322 total reads), tested on a Linux server, you get roughly 40% speed up using `fastqSplit` over `awk`.

- `awk` method; 132s

```
time ( gunzip -c data.fastq.gz | awk 'BEGIN{FS="[@:]"}{out=$4"."$5".fastq"; print > out; for (i = 1; i <= 3; i++) {getline ; print > out}}' )

real    2m12.825s
user    3m7.649s
sys     0m33.943s
```

- `fastqSplit`; 92s

```
$ time ( gunzip -c data.fastq.gz | .fastqSplit )

real    1m32.150s
user    2m13.515s
sys     0m43.867s
```
