package main
// This is the old simple version of the program before I added go routine handling
import (
	"bufio"
	// "compress/gzip"
	"flag"
	"fmt"
	gzip "github.com/klauspost/pgzip"
	"log"
	"os"
	"strings"
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
			// gz, err := gzip.NewReaderN(file, 512, 16)
			gz, err := gzip.NewReader(file)

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

func main() {
	headerDelim := flag.String("delim", ":", "delimiter character for the fastq header fields")
	flowCellFieldIndex := flag.Int("fcIndexPos", 2, "field number for the flowcell ID in the header")
	laneFieldIndex := flag.Int("laneIndexPos", 3, "field number for the lane ID in the header")
	readGroupJoinChar := flag.String("rgJoinChar", ".", "character used to join the flowcell and lane IDs to create the read group ID")
	flag.Parse()
	cliArgs := flag.Args() // all positional args passed

	// Initialize variables to keep track of output files
	outputFiles := make(map[string]FileHolder)
	// TODO: should we have an output directory to avoid filename conflict?

	// set blank nil readGroup var for later
	var readGroupID string

	// get input file scanner
	scanner, inputFile, inputGzFile := GetScanner(cliArgs)
	if inputFile != nil {
		defer inputFile.Close()
	}
	if inputGzFile != nil {
		defer inputGzFile.Close()
	}

	// Read and process each line from the file
	// if you get this error;
	// panic: runtime error: invalid memory address or nil pointer dereference
	// it means the first line did not have @ character
	for scanner.Scan() {
		line := scanner.Text()

		// Extract the flowcell ID from the first line of each FASTQ record
		if strings.HasPrefix(line, "@") {
			parts := strings.Split(line, *headerDelim)

			// make sure we have parts left
			if len(parts) < 2 {
				log.Fatalf("Error extracting read group ID from line:\n%v\n", line)
			}

			flowcellID := parts[*flowCellFieldIndex]
			laneID := parts[*laneFieldIndex]
			readGroupID = flowcellID + *readGroupJoinChar + laneID

			// Create a new output file if it doesn't exist
			if _, exists := outputFiles[readGroupID]; !exists {
				outputFileName := fmt.Sprintf("%s.fastq", readGroupID)
				outputFiles[readGroupID] = CreateFileHolder(outputFileName)
			}
		}

		// make sure that we successfully extracted a readGroupID
		if readGroupID == "" {
			log.Fatalf("Error extracting read group ID from line:\n%v\n", line)
		}

		// Write the line to the appropriate output file
		_, err := outputFiles[readGroupID].Writer.WriteString(line + "\n")
		if err != nil {
			log.Fatalf("Error writing to file: %v\n", err)
		}
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
