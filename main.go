package main

import (
	"bufio"
	// "compress/gzip"
	"flag"
	"fmt"
	gzip "github.com/klauspost/pgzip"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

// holds an open file handle so it doesnt need to be re-opened repeatedly
type FileHolder struct {
	File   *os.File
	Writer *bufio.Writer
}

func CreateFileHolder(filename string) FileHolder {
	// method for creating a new FileHolder instance
	// make sure that the caller runs :
	// defer FileHolder.Writer.Flush()
	// defer FileHolder.File.Close()
	outputFile, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Error creating output file: %v\n", err)
	}
	fileHolderObj := FileHolder{File: outputFile, Writer: bufio.NewWriter(outputFile)}
	return fileHolderObj
}

func GetScanner(args []string) (*bufio.Scanner, *os.File, *os.File) {
	// parses the args list to determine the correct way to create a new Scanner instance
	// the caller should make sure to run
	// defer file.Close()
	// defer gzFile.Closer()
	var scanner *bufio.Scanner
	var file *os.File
	var gzFile *os.File

	// Check if there are command-line arguments
	if len(args) > 0 {
		// If arguments are provided, assume the first argument is a file path
		filePath := args[0]
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatalf("Error opening file: %v\n", err)
		}

		// check if gz file
		if strings.HasSuffix(filePath, ".gz") {
			// NOTE: could use custom buffered reader here but in tests it did not speed anything up
			// https://github.com/klauspost/pgzip/blob/17e8dac29df8ce00febbd08ee5d8ee922024a003/gunzip.go#L139
			gz, err := gzip.NewReaderN(file, 100000000, 2) // 1000000 : 1MB, 16 blocks
			// gz, err := gzip.NewReader(file)

			if err != nil {
				log.Fatalf("Error opening file: %v\n", err)
			}

			scanner = bufio.NewScanner(gz)
		} else {
			// Create a scanner to read from a regular file
			scanner = bufio.NewScanner(file)
		}
	} else {
		// If no command-line arguments are provided, read from stdin
		scanner = bufio.NewScanner(os.Stdin)
	}

	return scanner, file, gzFile
}

func GetReadGroup(line string, config Config) string {
	// extracts the read group from the line

	// break the line into parts based on the delimiter
	lineParts := strings.Split(line, config.HeaderDelim)

	// fill in the selected portions of the line here
	rgParts := []string{}

	// placeholder for the final ID
	var readGroupID string

	// NOTE: figure out if we need to re-instate this safety check
	// NOTE: change this to 1
	// make sure we have parts left
	// if len(parts) < 2 {
	// 	log.Fatalf("Error extracting read group ID from line:\n%v\n", line)
	// }
	// readGroupID := parts[config.FlowCellFieldIndex] + config.ReadGroupJoinChar + parts[config.LaneFieldIndex]

	// get each desired field from the line
	for _, index := range config.FieldKeys {
		rgParts = append(rgParts, lineParts[index])
	}

	// join all the desired fields with the output delimiter
	readGroupID = strings.Join(rgParts, config.ReadGroupJoinChar)

	return readGroupID
}

func CreateOutputFileEntry(outputFiles map[string]FileHolder, readGroupID string, config Config) {
	// Create a new output file if it doesn't exist

	// NOTE: maps are pass by reference; https://stackoverflow.com/questions/40680981/are-maps-passed-by-value-or-by-reference-in-go
	// so the outputFiles map will be updated in-place in the scope of the caller
	if _, exists := outputFiles[readGroupID]; !exists {
		outputFileName := fmt.Sprintf("%s%s%s", config.FilePrefix, readGroupID, config.FileSuffix)
		outputFiles[readGroupID] = CreateFileHolder(outputFileName)
	}
}

func WriteLine(outputFiles map[string]FileHolder, readGroupID string, line string) {
	// writes the line to the file while also checking the readGroupID and finding the correct file handle

	// make sure that we successfully extracted a readGroupID earlier
	if readGroupID == "" {
		log.Fatalf("Error extracting read group ID from line:\n%v\n", line)
	}

	// Write the line to the appropriate output file
	_, err := outputFiles[readGroupID].Writer.WriteString(line + "\n")
	if err != nil {
		log.Fatalf("Error writing to file: %v\n", err)
	}
}

func runMain(config Config) {
	// Initialize variables to keep track of output files
	outputFiles := make(map[string]FileHolder)
	// TODO: should we have an output directory to avoid filename conflict?

	// set blank readGroup var for later
	var readGroupID string

	// get input file scanner
	scanner, inputFile, inputGzFile := GetScanner(config.CliArgs)
	if inputFile != nil {
		defer inputFile.Close()
	}
	if inputGzFile != nil {
		defer inputGzFile.Close()
	}

	// Read and process each line from the file
	// // if you get this error;
	// // panic: runtime error: invalid memory address or nil pointer dereference
	// // it usually means the first line did not have @ character
	for scanner.Scan() {
		line := scanner.Text()
		// Extract the flowcell ID from the first line of each FASTQ record
		if strings.HasPrefix(line, "@") {
			readGroupID = GetReadGroup(line, config)
			CreateOutputFileEntry(outputFiles, readGroupID, config)
		}
		WriteLine(outputFiles, readGroupID, line)
	}

	// Check for errors that may have occured while reading the file
	err := scanner.Err()
	if err != nil {
		log.Fatalf("Error while trying to read file: %v\n", err)
	}

	// Close all output files
	for _, fileHolder := range outputFiles {
		fileHolder.Writer.Flush()
		fileHolder.File.Close()
	}
}

func runMainP(config Config) {
	// parallel read / write implementation that uses a separate go routine to run the read operations
	// this is only faster when used with .gz input file
	// but is actually slower that just using gunzip -c | input for the program so its not actually worth using

	// Create a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Create a channel for reader goroutine to send lines through
	lines := make(chan string, config.BufferSize)

	// Launch worker goroutine to read lines
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(lines)

		// get input file scanner
		scanner, inputFile, inputGzFile := GetScanner(config.CliArgs)
		if inputFile != nil {
			defer inputFile.Close()
		}
		if inputGzFile != nil {
			defer inputGzFile.Close()
		}

		// Read each line from the file
		for scanner.Scan() {
			line := scanner.Text()
			lines <- line
		}
		// Check for errors that may have occured while reading the file
		err := scanner.Err()
		if err != nil {
			log.Fatalf("Error while trying to read file: %v\n", err)
		}
	}()

	// set blank readGroup var for later
	var readGroupID string
	// Initialize variables to keep track of output files
	outputFiles := make(map[string]FileHolder)

	// process the lines and write to file
	for line := range lines {
		if strings.HasPrefix(line, "@") {
			readGroupID = GetReadGroup(line, config)
			CreateOutputFileEntry(outputFiles, readGroupID, config)
		}
		WriteLine(outputFiles, readGroupID, line)
	}

	// signal workers to exit
	wg.Wait()

	// Close all output files
	for _, fileHolder := range outputFiles {
		fileHolder.Writer.Flush()
		fileHolder.File.Close()
	}
}

type Config struct {
	HeaderDelim       string
	ReadGroupJoinChar string
	RunParallel       bool
	BufferSize        int
	FileSuffix        string
	FilePrefix        string
	FieldKeys         []int
	CliArgs           []string
}

func main() {
	// runtime.GOMAXPROCS(2) // NOTE: dont use this because it defaults to the number of CPUs on recent Go versions
	headerDelim := flag.String("d", ":", "delimiter character for the fastq header fields")
	readGroupJoinChar := flag.String("j", ".", "character used to Join the selected key values on to create the read group ID")
	runParallel := flag.Bool("p", false, "read input on a separate thread (parallel)")
	bufferSize := flag.Int("b", 10000, "read buffer size (number of lines) when using parallel read method")
	fileSuffix := flag.String("suffix", ".fastq", "suffix for all output file names")
	filePrefix := flag.String("prefix", "", "prefix for all output file names")
	fieldKeys := flag.String("k", "2,3", "comma delimited string of 0-based integer field keys to split the fastq header line on")
	flag.Parse()
	cliArgs := flag.Args() // all positional args passed

	// parse the field keys to use for creating the Read Group
	fieldKeysParts := strings.Split(*fieldKeys, ",")
	fieldKeysInts := []int{}
	for _, key := range fieldKeysParts {
		keyInt, err := strconv.Atoi(key)
		if err != nil {
			log.Fatalf("Error while trying to parse key value: %v\n", err)
		}
		fieldKeysInts = append(fieldKeysInts, keyInt)
	}

	config := Config{
		*headerDelim,
		*readGroupJoinChar,
		*runParallel,
		*bufferSize,
		*fileSuffix,
		*filePrefix,
		fieldKeysInts,
		cliArgs,
	}

	if *runParallel {
		runMainP(config)
	} else {
		runMain(config)
	}
}
