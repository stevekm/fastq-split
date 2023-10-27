`fastqSplit`

Program to split a fastq file into separate files based on selected fields in each read's header.

# Usage

You can use `fastqSplit` to chunk a .fastq or .fastq.gz file into new files based on fields in the header line of each FASTQ read. For example, you could split all reads into separate files based on read group.

Your FASTQ records may have header lines that look like this;

```
@ABCD12:36:XYZ12345:1:1101:1524:1000 1:N:0:GCTCGGTA+GACACGTT
...
...
...
@ABCD12:36:XYZ12345:2:1101:2230:1000 1:N:0:GCTCGGTA+GACACGTT
...
...
...
```

Where the flow cell ID is `XYZ12345`, and the lane ID's are `1` and `2`. By default, `fastqSplit` will attempt to use these to create read group ID's in the format `<flow cell ID>.<lane number>`, which would be `XYZ12345.1` and `XYZ12345.2` in this example. You can run `fastqSplit` like this

```bash
fastqSplit data/test1.fastq
```

and it will output files `XYZ12345.1.fastq` and `XYZ12345.2.fastq`, which contain only the reads with these read groups.

The default parameters assume that your fastq file has headers in the [Illumina format](https://en.wikipedia.org/wiki/FASTQ_format#Illumina_sequence_identifiers). You can select different field delimiter and field indexes yourself based on the command line args provided by the program.

If your fastq file is .gz compressed, you should pipe it in with `zcat` or `gunzip` or `pigz` for faster processing;

```
gunzip -c data/test1.fastq.gz | ./fastqSplit
```

## Options

While `fastqSplit` can read from .gz compressed files, archive decompression and total execution time will be much faster if handled in a separate process and transmitted over stdin with a pipe as shown in the above examples. If you wish to use `fastqSplit`'s built-in .gz decompression, you should add the `-p` arg in order to utilize background buffered file decompression & scanning for a small performance boost.

Other command line options for `fastqSplit` include;

```
  -b int
    	read buffer size (number of lines) when using parallel read method (default 10000)
  -d string
    	delimiter character for the fastq header fields (default ":")
  -j string
    	character used to Join the selected key values on to create the read group ID (default ".")
  -k string
    	comma delimited string of 0-based integer field keys to split the fastq header line on (default "2,3")
  -p	read input on a separate thread (parallel)
  -prefix string
    	prefix for all output file names
  -suffix string
    	suffix for all output file names (default ".fastq")
```

### `-k` Field Keys

The values for the `-k` field key index arg are integers using 0-based indexing, so in the header example `@EAS139:136:FC706VJ:2`, the value `FC706VJ` is position 2 and value `2` is in position 3. These values are passed as a single comma-delimited string; `-k 2,3`.


# Installation

You can download a pre-built binary from the Releases page [here](https://github.com/stevekm/fastq-split/releases).

You can get it from Docker Hub [here](https://hub.docker.com/repository/docker/stevekm/fastq-split/tags?page=1&ordering=last_updated).

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

Using a 3.5GB size fastq.gz file with 268093288 lines (67023322 total reads), tested on an Ubuntu 20.04 Linux server, you get roughly 40% speed up using `fastqSplit` over GNU `awk`. Tests with other files show a similar 40-50% speed increase. You can get further speed increases by using `fastqSplit` in combination with [`pigz`](https://github.com/madler/pigz) for decompression.

- `awk` method; 132s

```
time ( gunzip -c data.fastq.gz | awk 'BEGIN{FS="[@:]"}{out=$4"."$5".fastq"; print > out; for (i = 1; i <= 3; i++) {getline ; print > out}}' )

real    2m12.825s
user    3m7.649s
sys     0m33.943s
```

- `fastqSplit`; 92s

```
$ time ( gunzip -c data.fastq.gz | ./fastqSplit )

real    1m32.150s
user    2m13.515s
sys     0m43.867s
```

- `fastqSplit` + `pigz`; 82s

```
$ time ( pigz -c -d data.fastq.gz | ./fastqSplit )

real    1m22.855s
user    1m45.716s
sys     1m6.102s
```

Tested with GNU Awk 5.0.1 on x86_64 GNU/Linux, and `pigz` version 2.8.

Note that you may be able to get faster speeds with `mawk` instead of GNU `awk`.

- `mawk`; 93s

```
$ time ( gunzip -c data.fastq.gz | mawk 'BEGIN{FS="[@:]"}{out=$4"."$5".mawk.fastq"; print > out; for (i = 1; i <= 3; i++) {getline ; print > out}}' )

real    1m33.219s
user    2m7.609s
sys     0m30.396s
```

- `mawk` + `pigz`; 72s

```
$ time ( pigz -d -c data.fastq.gz | mawk 'BEGIN{FS="[@:]"}{out=$4"."$5".mawk.fastq"; print > out; for (i = 1; i <= 3; i++) {getline ; print > out}}' )

real    1m12.000s
user    1m41.085s
sys     0m47.892s
```

Test with `mawk` version 1.3.4 20200120
