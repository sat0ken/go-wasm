package main

import (
	"bufio"
	"fmt"
	"go/types"
	"log"
	"os"
	"strings"
)

type wasmFunction struct {
	funcName     string
	param        []types.BasicKind
	result       []types.BasicKind
	instructions []string
}

func strToType(str string) types.BasicKind {
	switch str {
	case "i32":
		return types.Int32
	case "i64":
		return types.Int64
	case "f32":
		return types.Float32
	case "f64":
		return types.Float64
	}
	return types.Invalid
}

func setParams(lines []string) []types.BasicKind {
	var params []types.BasicKind
	for _, line := range lines {
		if strings.Contains(line, "32") || strings.Contains(line, "64") {
			params = append(params, strToType(line))
		}
	}
	return params
}

func readFuncDefinition(wasmFunction *wasmFunction, tokens string) {
	splitTokens := strings.Split(tokens, ")")
	for _, token := range splitTokens {
		splitSpace := strings.Split(token, " ")
		if strings.Contains(token, "export") {
			wasmFunction.funcName = strings.Replace(splitSpace[(len(splitSpace)-1)], "\"", "", -1)
		} else if strings.Contains(token, "param") {
			wasmFunction.param = setParams(splitSpace)
		} else if strings.Contains(token, "result") {
			wasmFunction.result = setParams(splitSpace)
		}
	}

}

func readWatFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	wasmfunc := wasmFunction{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		tokens := strings.Split(line, "\n")

		for _, token := range tokens {
			if strings.HasPrefix(token, "(func") {
				readFuncDefinition(&wasmfunc, token)
			}
		}
	}
	fmt.Println(wasmfunc)
}

func main() {
	readWatFile("./add.wat")
}
