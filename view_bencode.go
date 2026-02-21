package main

import (
	"fmt"
	"os"
)

func main() {
	file, err := os.Open("test.torrent")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Read the entire file
	data := make([]byte, 5000) // Read first 5000 bytes
	n, err := file.Read(data)
	if err != nil && err.Error() != "EOF" {
		fmt.Println("Error reading file:", err)
		return
	}

	data = data[:n]

	fmt.Println("=== RAW BENCODE FORMAT (First 5000 bytes) ===\n")

	// Display as text with special characters shown
	for i := 0; i < len(data); i++ {
		char := data[i]

		// Show printable characters directly
		if char >= 32 && char <= 126 {
			fmt.Printf("%c", char)
		} else if char == '\n' {
			fmt.Printf("\\n")
		} else if char == '\r' {
			fmt.Printf("\\r")
		} else if char == '\t' {
			fmt.Printf("\\t")
		} else {
			// Show control characters as hex
			fmt.Printf("[%02X]", char)
		}

		// Add newline every 150 chars for readability
		if (i+1)%150 == 0 {
			fmt.Printf("\n")
		}
	}

	fmt.Printf("\n\n=== BENCODE STRUCTURE BREAKDOWN ===\n\n")
	printBencodeStructure(data[:min(1000, len(data))], 0)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func printBencodeStructure(data []byte, indent int) {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	i := 0
	for i < len(data) {
		char := data[i]

		if char == 'd' {
			fmt.Printf("%s[DICT] Dictionary start\n", prefix)
			i++
		} else if char == 'l' {
			fmt.Printf("%s[LIST] List start\n", prefix)
			i++
		} else if char == 'e' {
			fmt.Printf("%s[END] End marker\n", prefix)
			i++
		} else if char == 'i' {
			// Integer
			j := i + 1
			for j < len(data) && data[j] != 'e' {
				j++
			}
			fmt.Printf("%s[INT] %s\n", prefix, string(data[i:j+1]))
			i = j + 1
		} else if char >= '0' && char <= '9' {
			// String with length prefix
			j := i
			for j < len(data) && data[j] != ':' {
				j++
			}
			lengthStr := string(data[i:j])
			j++ // skip ':'

			// Find the actual string
			var strEnd int
			var count int
			for strEnd = j; strEnd < len(data) && count < 100; strEnd++ {
				if data[strEnd] < 32 || data[strEnd] > 126 {
					break
				}
				count++
			}

			stringVal := string(data[j:strEnd])
			fmt.Printf("%s[STR] %s:%s\n", prefix, lengthStr, stringVal)
			i = strEnd + 1
		} else {
			i++
		}

		// Limit output
		if i > 500 {
			fmt.Printf("%s... (truncated, more data below)\n", prefix)
			break
		}
	}
}
