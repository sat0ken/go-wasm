package main

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
