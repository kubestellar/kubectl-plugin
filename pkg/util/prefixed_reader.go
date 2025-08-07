package util

import (
	"bufio"
	"fmt"
	"io"
)

// PrefixedReader reads from a reader and writes to a writer with a prefix on each line
func PrefixedReader(reader io.Reader, prefix string, writer io.Writer) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Fprintf(writer, "[%s] %s\n", prefix, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(writer, "[%s] Error reading: %v\n", prefix, err)
	}
}