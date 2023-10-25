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

(`fastqSplit` can read from .gz compressed files natively but the .gz archive compression will be much faster if handled in a separate process and transmitted over stdin with a pipe )

# Installation



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

Speed increases are supposedly due to holding open output file handles, but could possibly be increased further with other methods as well.