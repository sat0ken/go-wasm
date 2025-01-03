package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	// バイナリ識別子のまとめ
	// https://webassembly.github.io/spec/core/binary/types.html#types

	// Number Types
	numberTypeI32 = 0x7f
	numberTypeI64 = 0x7e
	numberTypeF32 = 0x7d
	numberTypeF64 = 0x7b

	// Vector Types
	vecType = 0x7b

	// Reference Type
	refTypeFuncRef   = 0x70
	refTypeExternRef = 0x6f

	// Global Type
	globalTypeConst = 0x00
	globalTypeVar   = 0x01

	// Function Types
	function = 0x60
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

const (
	// https://webassembly.github.io/spec/core/binary/instructions.html#instructions
	// Instructionsのまとめ

	// Variable Instructions
	localGet  = 0x20
	localSet  = 0x21
	localTee  = 0x22
	globalGet = 0x23
	globalSet = 0x24

	// Numeric Instructionsはもっといっぱいあるけど省略
	i32Add = 0x6a
)

var magicNumber = []byte{0x00, 0x61, 0x73, 0x6D} // "\0asm"
var version = []byte{0x01, 0x00, 0x00, 0x00}     // version

type WasmModule struct {
	typeSection     typeSection
	funcName        []byte
	functionSection functionSection
	exportSection   exportSection
	instructionSet  []instruction
	codeSection     codeSection
}

type instruction struct {
	order string
	args  string
}

type typeSection struct {
	moduleSection uint8  // セクションID
	sectionSize   uint8  // サイズ
	moduleType    uint8  // タイプ
	argsNum       uint8  // 引数の数
	args          []byte // 引数の型
	result        []byte // 戻り値
}

func (t *typeSection) toByte() []byte {
	var b []byte
	b = append(b, t.moduleSection)
	// セクションサイズ
	b = append(b, 0x07)
	// 型の数（1つの関数型を定義）
	b = append(b, 0x01)
	b = append(b, t.moduleType)
	b = append(b, t.argsNum)
	b = append(b, t.args...)
	// 戻り値の数は1つ
	b = append(b, 0x01)
	b = append(b, t.result...)
	return b
}

type functionSection struct {
	sectionID   uint8 // セクションID
	sectionSize uint8 // サイズ
	num         uint8 // 関数の数
}

func (f *functionSection) toByte() []byte {
	var b []byte
	b = append(b, f.sectionID)
	b = append(b, f.sectionSize)
	b = append(b, f.num)
	return b
}

type exportSection struct {
	sectionID   uint8  // セクションID
	sectionSize uint8  // サイズ
	exportNum   uint8  // エクスポートの数
	nameLength  uint8  // 名前の長さ
	name        []byte // 名前
	exportType  uint8  // エクスポートの種類
	index       uint8  // 関数のインデックス(0番目なら0が入る)
}

func (e *exportSection) toByte() []byte {
	var b []byte
	b = append(b, e.sectionID)
	b = append(b, e.sectionSize)
	b = append(b, e.exportNum)
	b = append(b, e.nameLength)
	b = append(b, e.name...)
	b = append(b, e.exportType)
	b = append(b, e.index)
	return b
}

type codeSection struct {
	sectionID   uint8
	sectionSize uint8
	funcNum     uint8
	codeSize    uint8
	code        []byte
}

func (c *codeSection) toByte() []byte {
	var b []byte
	b = append(b, c.sectionID)
	b = append(b, c.sectionSize)
	b = append(b, c.funcNum)
	b = append(b, c.codeSize)
	b = append(b, c.code...)
	return b
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
			wasmFunction.funcName = []byte(strings.Replace(splitSpace[(len(splitSpace)-1)], "\"", "", -1))
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
	wasmFunction.functionSection.sectionSize = 2
	// Indexは決め打ちで0
	wasmFunction.functionSection.num = 0
}

func setExportSection(modtype uint8, export *exportSection, funcName []byte) {
	export.sectionID = exportSectionNum
	export.sectionSize = uint8(4 + len(funcName))
	export.exportNum = 1
	export.nameLength = uint8(len(funcName))
	export.name = funcName

	// functionだけ対応,本当は後tableとmemoryとglobalもある
	// https://webassembly.github.io/spec/core/binary/modules.html#export-section
	switch modtype {
	case function:
		export.exportType = 0
	}
	// 0番目の関数をエクスポートで決め打ち
	export.index = 0
}

func setCodeSection(instructions []instruction, codeSection *codeSection) {
	// codeを生成する
	codeSection.code = append(codeSection.code, 0) // ローカル変数の数は0
	for _, inst := range instructions {
		switch {
		case strings.EqualFold(inst.order, "local.get"):
			codeSection.code = append(codeSection.code, localGet)
			args, _ := strconv.ParseInt(inst.args, 10, 64)
			codeSection.code = append(codeSection.code, byte(args))
		case strings.EqualFold(inst.order, "i32.add"):
			codeSection.code = append(codeSection.code, i32Add)
		}
	}
	// 関数の終了を入れる
	codeSection.code = append(codeSection.code, 0x0b)
	// 関数コードのサイズをセット
	codeSection.codeSize = uint8(len(codeSection.code))
	codeSection.sectionID = 0x0a
	codeSection.sectionSize = 2 + codeSection.codeSize
	codeSection.funcNum = 1
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
					setExportSection(wasmmodule.typeSection.moduleType, &wasmmodule.exportSection, wasmmodule.funcName)
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
	setCodeSection(wasmmodule.instructionSet, &wasmmodule.codeSection)
	return wasmmodule
}

func createWasmBinary(wasmmodule WasmModule) {
	var bin []byte
	// Magic Numberをセット
	bin = append(bin, magicNumber...)
	// Versionをセット
	bin = append(bin, version...)
	// Typeセクションをセット
	bin = append(bin, wasmmodule.typeSection.toByte()...)
	// Functionセクションをセット
	bin = append(bin, wasmmodule.functionSection.toByte()...)
	// Exportセクションをセット
	//bin = append(bin, wasmmodule.exportSection.toByte()...)
	printBytes(bin)
	printBytes(wasmmodule.exportSection.toByte())
	printBytes(wasmmodule.codeSection.toByte())
}

func main() {
	createWasmBinary(readWatFile("./add.wat"))
}
