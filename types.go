package main

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
	index       uint8 // 関数が参照する型インデックス（0番目の型を参照）
}

func (f *functionSection) toByte() []byte {
	var b []byte
	b = append(b, f.sectionID)
	b = append(b, f.sectionSize)
	b = append(b, f.num)
	b = append(b, f.index)
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
