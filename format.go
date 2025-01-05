package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

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
	// 関数の数は決め打ちで1
	wasmFunction.functionSection.num = 1
	// 関数の数は1つなので参照するIndexは0
	wasmFunction.functionSection.index = 0
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
		case strings.EqualFold(inst.order, "i32.sub"):
			codeSection.code = append(codeSection.code, i32Sub)
		case strings.EqualFold(inst.order, "i32.mul"):
			codeSection.code = append(codeSection.code, i32Mul)
		case strings.EqualFold(inst.order, "i32.div_s"):
			codeSection.code = append(codeSection.code, i32Divs)
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

func createWasmBinary(wasmmodule WasmModule) []byte {
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
	bin = append(bin, wasmmodule.exportSection.toByte()...)
	// Codeセクションをセット
	bin = append(bin, wasmmodule.codeSection.toByte()...)
	return bin
}

func readULEB128(reader *bytes.Reader) (uint32, error) {
	var result uint32
	var shift uint
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return 0, err
		}
		result |= uint32(b&0x7F) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
	}
	return result, nil
}

func validateWASMHeader(data []byte) error {
	if len(data) < 8 {
		return errors.New("file too small to be a valid WASM")
	}
	if bytes.Equal(data[:4], magicNumber) == false {
		return fmt.Errorf("invalid magic number: 0x%x", data[:4])
	}
	if bytes.Equal(data[4:8], version) == false {
		return fmt.Errorf("unsupported version: 0x%x", data[4:8])
	}
	return nil
}

func parseWASMCodeSection(wasmData []byte) error {
	reader := bytes.NewReader(wasmData[8:]) // Skip magic number and version

	for reader.Len() > 0 {
		sectionID, err := reader.ReadByte()
		if err != nil {
			return fmt.Errorf("failed to read section ID: %w", err)
		}

		sectionSize, err := readULEB128(reader)
		if err != nil {
			return fmt.Errorf("failed to read section size: %w", err)
		}

		sectionData := make([]byte, sectionSize)
		if _, err := reader.Read(sectionData); err != nil {
			return fmt.Errorf("failed to read section data: %w", err)
		}

		if sectionID == codeSectionNum {
			return executeCodeSection(sectionData)
		}
	}

	return errors.New("code section not found")
}
