package helper

import (
	"bufio"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// ParseInts : takes a comma separated ints in a string a returns a slice of ints
func ParseInts(ints string) []int {
	res := make([]int, 0)
	if ints == "" {
		return res
	}

	intsArray := strings.Split(ints, ",")
	for _, i := range intsArray {
		parsedInt, _ := strconv.Atoi(i)
		res = append(res, parsedInt)
	}

	return res
}

// FileToStrings : reads a file and returns its lines in a slice of strings
func FileToStrings(filename string) []string {
	res := make([]string, 0)
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		res = append(res, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return res
}

// FileToInts : reads a file and returns its lines in a slice of ints
func FileToInts(filename string) []int {
	res := make([]int, 0)
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(filename)
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parsedInt, _ := strconv.Atoi(scanner.Text())
		res = append(res, parsedInt)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
		return nil
	}
	return res
}

// Contains : checks if a slice contains an int
func Contains(array []int, element int) bool {
	for _, e := range array {
		if e == element {
			return true
		}
	}
	return false
}

// ContainsStr : checks if a slice contains a string
func ContainsStr(array []string, element string) bool {
	for _, e := range array {
		if e == element {
			return true
		}
	}
	return false
}

// GetWordlists : returns a slice of available wordlists in the resources folder
func GetDNSlists() []string {
	files, err := os.ReadDir("ressources/subdomains")
	res := make([]string, 0)
	if err != nil {
		log.Fatal(err)
		return res
	}
	for _, file := range files {
		res = append(res, file.Name())
	}
	return res
}

// GetWordlists : returns a slice of available wordlists in the resources folder
func GetWordlists() []string {
	files, err := os.ReadDir("ressources/dirs")
	res := make([]string, 0)
	if err != nil {
		log.Fatal(err)
		return res
	}
	for _, file := range files {
		res = append(res, file.Name())
	}
	return res
}

// GetPortlists : returns a slice of available portlists in the resources folder
func GetPortlists() []string {
	files, err := os.ReadDir("ressources/ports")
	res := make([]string, 0)
	if err != nil {
		log.Fatal(err)
		return res
	}
	for _, file := range files {
		res = append(res, file.Name())
	}
	return res
}

// Reverse : reverses a slice
func Reverse(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

// RemoveFromSlice : removes a string from slice
func RemoveFromSlice(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// ChunkSlice : chunks slice into sub-silces
func ChunkSlice(slice []string, chunkSize int) [][]string {
	var chunks [][]string
	for {
		if len(slice) == 0 {
			break
		}

		// necessary check to avoid slicing beyond
		// slice capacity
		if len(slice) < chunkSize {
			chunkSize = len(slice)
		}

		chunks = append(chunks, slice[0:chunkSize])
		slice = slice[chunkSize:]
	}

	return chunks
}

// MergeSlices : merges slices into one slice
func MergeSlice(chunks [][]string) []string {
	var mergedSlice []string
	for _, slice := range chunks {
		mergedSlice = append(mergedSlice, slice...)
	}

	return mergedSlice
}

// Tokenize : uses a regex to tokenize the search string
func Tokenize(strsearch string) map[string]string {
	rmain := regexp.MustCompile(`(^[A-Za-z0-9.]+\s)`)
	rtokens := regexp.MustCompile(`(\S+:"[^"]+")`)
	matches := rtokens.FindAllStringSubmatch(strsearch, -1)
	mp := make(map[string]string)
	for _, v := range matches {
		splitted := strings.Split(v[1], ":")
		mp[splitted[0]] = strings.Replace(splitted[1], "\"", "", -1)
	}
	matches = rmain.FindAllStringSubmatch(strsearch, -1)
	for _, v := range matches {
		mp["default"] = strings.Replace(v[1], " ", "", -1)
	}

	firstWord := ""
	for i, c := range strsearch {
		if c == ':' {
			break
		}
		firstWord += string(strsearch[i])
	}
	if mp["default"] == firstWord {
		mp["default"] = ""
	}
	return mp
}

// GetTags : uses a regex to split comment into tags
func GetTags(strtags string) []string {
	res := make([]string, 0)
	rmain := regexp.MustCompile(`(#[A-Za-z0-9_-]+\s?)`)
	matches := rmain.FindAllStringSubmatch(strtags, -1)

	for _, v := range matches {
		res = append(res, strings.Replace(v[1], " ", "", -1))
	}

	return res
}
