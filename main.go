package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	// module type
	function = 0x60

	// https://webassembly.github.io/spec/core/binary/types.html#number-types
	// Number Types
	numberTypeI32 = 0x7f
	numberTypeI64 = 0x7e
	numberTypeF32 = 0x7d
	numberTypeF64 = 0x7b
)

const (
	// https://webassembly.github.io/spec/core/binary/modules.html#sections
	customSectionNum = iota
	typeSectionNum
	importSectionNum
	functionSectionNum
	tableSectionNum
	memorySectionNum
	globalSectionNum
	exportSectionNum
	startSectionNum
	elementSectionNum
	codeSectionNum
	dataSectionNum
	dataCountSectionNum
)

var magicNumber = []byte{0x00, 0x61, 0x73, 0x6D} // "\0asm"
var version = []byte{0x01, 0x00, 0x00, 0x00}     // version

type instruction struct {
	order string
	args  string
}

type typeSection struct {
	moduleSection uint8  // セクションID
	size          uint8  // サイズ
	moduleType    uint8  // タイプ
	argsNum       uint8  // 引数の数
	args          []byte // 引数
	result        []byte // 戻り値
}

type functionSection struct {
	sectionID uint8 // セクションID
	size      uint8 // サイズ
	num       uint8 // 関数の数
}

type exportSection struct{}

type codeSection struct{}

type WasmModule struct {
	typeSection     typeSection
	funcName        string
	functionSection functionSection

	instructionSet []instruction
}

func strToType(str string) int {
	switch str {
	case "i32":
		return numberTypeI32
	case "i64":
		return numberTypeI64
	case "f32":
		return numberTypeF32
	case "f64":
		return numberTypeF64
	default:
		return -1
	}
}

func printBytes(data []byte) {
	for i, b := range data {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Printf("%02x", b)
	}
	fmt.Println()
}

func setParams(lines []string) ([]byte, uint8) {
	var params []byte
	var cnt uint8
	for _, line := range lines {
		if strings.Contains(line, "32") || strings.Contains(line, "64") {
			params = append(params, byte(strToType(line)))
			cnt++
		}
	}
	return params, cnt
}

func setTypeSection(wasmFunction *WasmModule, tokens string) {
	splitTokens := strings.Split(tokens, ")")

	for _, token := range splitTokens {
		splitSpace := strings.Split(token, " ")
		switch {
		case strings.Contains(token, "export"):
			// 関数名をセットする
			wasmFunction.funcName = strings.Replace(splitSpace[(len(splitSpace)-1)], "\"", "", -1)
		case strings.Contains(token, "param"):
			// 引数をセットする
			wasmFunction.typeSection.args, wasmFunction.typeSection.argsNum = setParams(splitSpace)
		case strings.Contains(token, "result"):
			// 戻り値をセットする
			wasmFunction.typeSection.result, _ = setParams(splitSpace)
		}
	}
}

func setFunctionSection(wasmFunction *WasmModule) {
	switch wasmFunction.typeSection.moduleType {
	case function:
		wasmFunction.functionSection.sectionID = functionSectionNum
	}
	// サイズは決め打ちで2
	wasmFunction.functionSection.size = 2
	// Indexは決め打ちで0
	wasmFunction.functionSection.num = 0
}

func readInstructions(wasmFunction *WasmModule, tokens string) {
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

func readWatFile(path string) WasmModule {

	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	wasmmodule := WasmModule{}
	scanner := bufio.NewScanner(f)
	cnt := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 0行目にmodule宣言がなければエラーで終了
		if cnt == 0 && !strings.Contains(line, "module") {
			log.Fatalf("WASM text format error")
		}
		// 1行目を解析
		if cnt == 1 {
			tokens := strings.Split(line, "\n")
			for _, token := range tokens {
				// 関数名、引数、戻り値を解析する
				// 関数型以外のモジュールもあるが今はこれだけ対応
				switch {
				case strings.HasPrefix(token, "(func"):
					wasmmodule.typeSection.moduleSection = typeSectionNum
					wasmmodule.typeSection.moduleType = function
					setTypeSection(&wasmmodule, token)
					setFunctionSection(&wasmmodule)
				}
			}
		}

		// 2行目以後はInstructionを解析する
		if 2 <= cnt {
			//fmt.Println(line)
			readInstructions(&wasmmodule, line)
		}
		cnt++
	}

	return wasmmodule
}

func createWasmBinary(wasmmodule WasmModule) {
	var bin []byte
	bin = append(bin, magicNumber...)
	bin = append(bin, version...)
	// セクションID
	bin = append(bin, wasmmodule.typeSection.moduleSection)
	// サイズ
	bin = append(bin, 0x07)
	bin = append(bin, 0x01)
	// タイプ
	bin = append(bin, wasmmodule.typeSection.moduleType)
	// 引数の数
	bin = append(bin, wasmmodule.typeSection.argsNum)
	// 引数の型
	bin = append(bin, wasmmodule.typeSection.args...)
	// 戻り値の数は1つ
	bin = append(bin, 0x01)
	// 戻り値の型
	bin = append(bin, wasmmodule.typeSection.result...)

	bin = append(bin, wasmmodule.functionSection.sectionID)
	bin = append(bin, wasmmodule.functionSection.size)
	bin = append(bin, wasmmodule.functionSection.num)

	printBytes(bin)
}

func main() {
	createWasmBinary(readWatFile("./add.wat"))
}
