package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

var version = "0.1.0"

// borrowed from https://gist.github.com/slimsag/14c66b88633bd52b7fa710349e4c6749
func replaceAllSubmatchFunc(re *regexp.Regexp, src []byte, repl func([][]byte) [][]byte, n int) []byte {
	var (
		result  = make([]byte, 0, len(src))
		matches = re.FindAllSubmatchIndex(src, n)
		last    = 0
	)
	for _, match := range matches {
		// Append bytes between our last match and this one (i.e. non-matched bytes).
		matchStart := match[0]
		matchEnd := match[1]
		result = append(result, src[last:matchStart]...)
		last = matchEnd

		// Determine the groups / submatch bytes and indices.
		groups := [][]byte{}
		groupIndices := [][2]int{}
		for i := 2; i < len(match); i += 2 {
			start := match[i]
			end := match[i+1]
			groups = append(groups, src[start:end])
			groupIndices = append(groupIndices, [2]int{start, end})
		}

		// Replace the groups as desired.
		groups = repl(groups)

		// Append match data.
		lastGroup := matchStart
		for i, newValue := range groups {
			// Append bytes between our last group match and this one (i.e. non-group-matched bytes)
			groupStart := groupIndices[i][0]
			groupEnd := groupIndices[i][1]
			result = append(result, src[lastGroup:groupStart]...)
			lastGroup = groupEnd

			// Append the new group value.
			result = append(result, newValue...)
		}
		result = append(result, src[lastGroup:matchEnd]...) // remaining
	}
	result = append(result, src[last:]...) // remaining
	return result
}

var placeholderRegexp = regexp.MustCompile(`(?m)[^\\]{1}(\$\{[a-zA-Z\.0-9-_:]+\})`)

// substitute ${propname} but ignore \${propname}
func substitute(props map[string]string, text string, notFoundErr bool, foundProps map[string]string) (string, error) {
	var errorFromReplace error
	return string(replaceAllSubmatchFunc(placeholderRegexp, []byte(text), func(groups [][]byte) [][]byte {
		for idx, prop := range groups {
			filtered := strings.TrimSpace(string(prop))
			propName := strings.TrimSpace(filtered[2 : len(filtered)-1])
			propVal, ok := props[propName]
			if !ok {
				if notFoundErr {
					errorFromReplace = errors.New("Property '" + propName + "' not found")
				}

			} else {
				groups[idx] = []byte(propVal)
				foundProps[filtered] = propVal
			}

		}
		return groups
	}, 10000)), errorFromReplace
}

func substituteStream(props map[string]string, in io.Reader, out io.Writer, failIfNotFound bool, foundProps map[string]string) error {
	inStr, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	res, err := substitute(props, string(inStr), failIfNotFound, foundProps)

	if err != nil {
		return err
	}

	_, err = io.WriteString(out, res)
	if err != nil {
		return err
	}
	return nil
}

func execute(propertyFiles []string, inlineProperties []string, files []string, inPlace bool, failIfNotFound bool) error {
	props := make(map[string]string)
	for _, propFile := range propertyFiles {
		err := ReadPropertiesFile(propFile, props)
		if err != nil {
			return err
		}
	}

	for _, inlineProps := range inlineProperties {
		err := ReadProperties(inlineProps, props)
		if err != nil {
			return err
		}
	}

	for _, v := range files {
		f, err := os.Open(v)
		if err != nil {
			return err
		}

		var out io.Writer
		if inPlace {
			out = &bytes.Buffer{}
		} else {
			out = os.Stdout
		}
		foundProps := make(map[string]string)

		err = substituteStream(props, f, out, failIfNotFound, foundProps)
		f.Close()
		if err != nil {
			return err
		}

		if inPlace {
			err = ioutil.WriteFile(v, out.(*bytes.Buffer).Bytes(), 0644)
			fmt.Printf("%s: %b placeholder(s) replaced\n", v, len(foundProps))
			for key, value := range foundProps {
				fmt.Printf("%s: '%s' => '%s'\n", v, key, value)
			}
		}

		if err != nil {
			return err
		}
	}
	return nil
}

type stringsArrayFlag []string

func (i *stringsArrayFlag) String() string {
	return "my string representation"
}

func (i *stringsArrayFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	var propertyFiles stringsArrayFlag
	flag.Var(&propertyFiles, "f", "Property files. Can be used multiple times")
	var propertiesSources stringsArrayFlag
	flag.Var(&propertiesSources, "p", "Inline properties declaration. Can be used multiple times. Provide result of `env` command to use env variable")

	inPlace := flag.Bool("i", false, "Do substitution in place")
	failIfNotFound := flag.Bool("fail-not-found", false, "Fail if no property required by placehold found in source")

	flag.Parse()
	files := flag.Args()

	err := execute(propertyFiles, propertiesSources, files, *inPlace, *failIfNotFound)
	if err != nil {
		log.Fatalln(err)
	}

}
