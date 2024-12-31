package main

import (
	"bufio"
	"fmt"
	"go/types"
	"log"
	"os"
	"strings"
)

var magicNumber = []byte{0x00, 0x61, 0x73, 0x6D} // "\0asm"
var version = []byte{0x01, 0x00, 0x00, 0x00}     // version

type instruction struct {
	order string
	args  string
}

type wasmFunction struct {
	funcName       string
	param          []types.BasicKind
	result         []types.BasicKind
	instructionSet []instruction
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
			// 関数名をセットする
			wasmFunction.funcName = strings.Replace(splitSpace[(len(splitSpace)-1)], "\"", "", -1)
		} else if strings.Contains(token, "param") {
			// 引数をセットする
			wasmFunction.param = setParams(splitSpace)
		} else if strings.Contains(token, "result") {
			// 戻り値をセットする
			wasmFunction.result = setParams(splitSpace)
		}
	}
}

func readInstructions(wasmFunction *wasmFunction, tokens string) {
	splitTokens := strings.Split(tokens, " ")

	if len(splitTokens) == 1 {
		// 引数がない命令なら命令のみをセットする
		wasmFunction.instructionSet = append(wasmFunction.instructionSet, instruction{order: strings.Replace(splitTokens[0], ")", "", -1)})
	} else {
		// 引数があれば命令と引数をセットする
		wasmFunction.instructionSet = append(wasmFunction.instructionSet, instruction{
			order: splitTokens[0],
			args:  splitTokens[1]})
	}
}

func readWatFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	wasmfunc := wasmFunction{}
	scanner := bufio.NewScanner(f)
	cnt := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		tokens := strings.Split(line, "\n")

		for _, token := range tokens {
			// 関数名、引数、戻り値を解析する
			if strings.HasPrefix(token, "(func") {
				readFuncDefinition(&wasmfunc, token)
			}
		}
		// 2行目以後はInstructionを解析する
		if 2 <= cnt {
			//fmt.Println(line)
			readInstructions(&wasmfunc, line)
		}
		cnt++
	}
	fmt.Println(wasmfunc)
}

func main() {
	readWatFile("./add.wat")
}
