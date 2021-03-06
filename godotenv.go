/*
A go port of the ruby dotenv library (https://github.com/bkeepers/dotenv)

Examples/readme can be found on the github page at https://github.com/joho/godotenv

The TL;DR is that you make a .env file that looks something like

		SOME_ENV_VAR=somevalue

and then in your go code you can call

		godotenv.Load()

and all the env vars declared in .env will be avaiable through os.Getenv("SOME_ENV_VAR")
*/
package godotenv

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
)

/*
	Call this function as close as possible to the start of your program (ideally in main)

	If you call Load without any args it will default to loading .env in the current path

	You can otherwise tell it which files to load (there can be more than one) like

		godotenv.Load("fileone", "filetwo")

	It's important to note that it WILL NOT OVERRIDE an env variable that already exists - consider the .env file to set dev vars or sensible defaults
*/
func Load(filenames ...string) (err error) {
	filenames = filenamesOrDefault(filenames)

	for _, filename := range filenames {
		err = loadFile(filename)
		if err != nil {
			return // return early on a spazout
		}
	}
	return
}

func Read(filenames ...string) (envMap map[string]string, err error) {
	filenames = filenamesOrDefault(filenames)
	envMap = make(map[string]string)

	for _, filename := range filenames {
		individualEnvMap, individualErr := readFile(filename)

		if individualErr != nil {
			err = individualErr
			return // return early on a spazout
		}

		for key, value := range individualEnvMap {
			envMap[key] = value
		}
	}

	return
}

func filenamesOrDefault(filenames []string) []string {
	if len(filenames) == 0 {
		return []string{".env"}
	} else {
		return filenames
	}
}

func loadFile(filename string) (err error) {
	envMap, err := readFile(filename)
	if err != nil {
		return
	}

	for key, value := range envMap {
		os.Setenv(key, value)
	}

	return
}

func readFile(filename string) (envMap map[string]string, err error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	envMap = make(map[string]string)

	lines := strings.Split(string(content), "\n")

	for _, fullLine := range lines {
		if !isIgnoredLine(fullLine) {
			key, value, err := parseLine(fullLine)

			if err == nil && os.Getenv(key) == "" {
				envMap[key] = value
			}
		}
	}
	return
}

func parseLine(line string) (key string, value string, err error) {
	if len(line) == 0 {
		err = errors.New("zero length string")
		return
	}

	// ditch the comments (but keep quoted hashes)
	if strings.Contains(line, "#") {
		segmentsBetweenHashes := strings.Split(line, "#")
		quotesAreOpen := false
		segmentsToKeep := make([]string, 0)
		for _, segment := range segmentsBetweenHashes {
			if strings.Count(segment, "\"") == 1 || strings.Count(segment, "'") == 1 {
				if quotesAreOpen {
					quotesAreOpen = false
					segmentsToKeep = append(segmentsToKeep, segment)
				} else {
					quotesAreOpen = true
				}
			}

			if len(segmentsToKeep) == 0 || quotesAreOpen {
				segmentsToKeep = append(segmentsToKeep, segment)
			}
		}

		line = strings.Join(segmentsToKeep, "#")
	}

	// now split key from value
	splitString := strings.Split(line, "=")

	if len(splitString) != 2 {
		// try yaml mode!
		splitString = strings.Split(line, ":")
	}

	if len(splitString) != 2 {
		err = errors.New("Can't separate key from value")
		return
	}

	// Parse the key
	key = splitString[0]
	if strings.HasPrefix(key, "export") {
		key = strings.TrimPrefix(key, "export")
	}
	key = strings.Trim(key, " ")

	// Parse the value
	value = splitString[1]
	// trim
	value = strings.Trim(value, " ")

	// check if we've got quoted values
	if strings.Count(value, "\"") == 2 || strings.Count(value, "'") == 2 {
		// pull the quotes off the edges
		value = strings.Trim(value, "\"'")

		// expand quotes
		value = strings.Replace(value, "\\\"", "\"", -1)
		// expand newlines
		value = strings.Replace(value, "\\n", "\n", -1)
	}

	return
}

func isIgnoredLine(line string) bool {
	trimmedLine := strings.Trim(line, " \n\t")
	return len(trimmedLine) == 0 || strings.HasPrefix(trimmedLine, "#")
}
