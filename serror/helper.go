package serror

import (
	"fmt"
	nativeRuntime "runtime"
	"strings"

	"github.com/DeniesKresna/gohelper/utstring"
)

func getErrorFlow() (rowsJoin []string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error occured", r)
		}
	}()

	pc := make([]uintptr, 10)
	n := nativeRuntime.Callers(3, pc)
	frames := nativeRuntime.CallersFrames(pc[:n])

	serrorTraceDirs := utstring.GetEnv("SERROR_TRACE_DIRS", "")
	rootMandatoryDirs := strings.Split(serrorTraceDirs, ",")
	if len(rootMandatoryDirs) <= 0 {
		rootMandatoryDirs = []string{"services"}
	}

	getIfStringContainsMandatoryDir := func(str string) string {
		for _, v := range rootMandatoryDirs {
			if strings.Contains(str, v) {
				return v
			}
		}
		return ""
	}

	for {
		frame, more := frames.Next()

		rootDir := getIfStringContainsMandatoryDir(frame.Function)
		if rootDir == "" {
			break
		}

		// get latest string based on function as shortFunction
		var functionName string
		fFunctions := strings.Split(frame.Function, "/")
		functionNames := fFunctions[len(fFunctions)-1]
		functionNameSegments := strings.Split(functionNames, ".")
		functionName = functionNameSegments[len(functionNameSegments)-1]

		// get latest string basen on file path
		mandStrIndex := strings.Index(frame.File, rootDir)
		if mandStrIndex < 0 {
			break
		}
		fileName := frame.File[mandStrIndex:]

		// get the lineNumb only in first layer
		lineNumb := frame.Line

		// generating row and join the string
		rj := fmt.Sprintf("%s (line: %d) on function %s", fileName, lineNumb, functionName)

		if !more {
			break
		}
		rowsJoin = append(rowsJoin, rj)
	}

	return
}
