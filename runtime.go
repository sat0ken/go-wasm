package main

import (
	"bytes"
	"fmt"
)

func executeCodeSection(sectionData []byte) error {
	reader := bytes.NewReader(sectionData)

	funcCount, err := readULEB128(reader)
	if err != nil {
		return fmt.Errorf("failed to read function count: %w", err)
	}

	for i := uint32(0); i < funcCount; i++ {
		bodySize, err := readULEB128(reader)
		if err != nil {
			return fmt.Errorf("failed to read function body size: %w", err)
		}

		body := make([]byte, bodySize)
		if _, err := reader.Read(body); err != nil {
			return fmt.Errorf("failed to read function body: %w", err)
		}

		err = executeFunctionBody(body)
		if err != nil {
			return fmt.Errorf("failed to execute function body: %w", err)
		}
	}

	return nil
}

func applyBinaryOperation(stack *Stack, operation func(a, b int32) int32) error {
	b, err := stack.Pop()
	if err != nil {
		return err
	}
	a, err := stack.Pop()
	if err != nil {
		return err
	}
	result := operation(a, b)
	stack.Push(result)
	return nil
}

func executeFunctionBody(body []byte) error {
	stack := &Stack{}
	stack.Push(20)
	stack.Push(10)

	reader := bytes.NewReader(body)

	for reader.Len() > 0 {
		opcode, err := reader.ReadByte()
		if err != nil {
			return fmt.Errorf("failed to read opcode: %w", err)
		}

		switch opcode {
		case i32Add:
			err := applyBinaryOperation(stack, func(a, b int32) int32 {
				fmt.Printf("Executed i32.add: %d + %d = %d\n", a, b, a+b)
				return a + b
			})
			if err != nil {
				return err
			}
		case i32Sub:
			err := applyBinaryOperation(stack, func(a, b int32) int32 {
				fmt.Printf("Executed i32.sub: %d - %d = %d\n", a, b, a-b)
				return a - b
			})
			if err != nil {
				return err
			}
		case i32Mul:
			err := applyBinaryOperation(stack, func(a, b int32) int32 {
				fmt.Printf("Executed i32.mul: %d * %d = %d\n", a, b, a*b)
				return a * b
			})
			if err != nil {
				return err
			}
		case i32Divs:
			err := applyBinaryOperation(stack, func(a, b int32) int32 {
				fmt.Printf("Executed i32.mul: %d * %d = %d\n", a, b, a/b)
				return a / b
			})
			if err != nil {
				return err
			}
			//default:
			//	fmt.Printf("Unsupported opcode: 0x%x\n", opcode)
		}
	}
	fmt.Println("After execution, stack:", stack.data)

	return nil
}
