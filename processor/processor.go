/*
package main of vcpu
*/
package main

import (
	"errors"
	"fmt"
	"github.com/jfitz/virtual-processor/vputils"
	"os"
	"strings"
)

type stack []byte

func (s stack) push(v byte) stack {
	return append(s, v)
}

func (s stack) top() (byte, error) {
	if len(s) == 0 {
		return 0, errors.New("Stack underflow")
	}

	last := len(s) - 1
	return s[last], nil
}

func (s stack) pop() (stack, error) {
	if len(s) == 0 {
		return s, errors.New("Stack underflow")
	}

	last := len(s) - 1
	return s[:last], nil
}

type vector []byte

func (v vector) get(offset int) (byte, error) {
	if offset < 0 || offset >= len(v) {
		return 0, errors.New("Index out of range")
	}

	return v[offset], nil
}

func executeCode(code vector, data vector) {
	pc := 0
	vStack := make(stack, 0)
	// bytesPerCodeAddress := 1
	bytesPerDataAddress := 1

	fmt.Printf("Execution started at %04x\n", pc)
	halt := false
	for !halt {
		opcode, err := code.get(pc)
		vputils.CheckAndPanic(err)

		switch opcode {
		case 0x00:
			// EXIT
			halt = true
			pc += 1 // the opcode
		case 0x40:
			// PUSH.B Value
			codeAddress := pc + 1
			value, err := code.get(codeAddress)
			vputils.CheckAndPanic(err)

			vStack = vStack.push(value)
			pc += 1 // the opcode
			pc += 1 // the target value
		case 0x41:
			// PUSH.B Address
			codeAddress := pc + 1
			dataAddr, err := code.get(codeAddress)
			dataAddress := int(dataAddr)
			vputils.CheckAndPanic(err)

			value, err := data.get(dataAddress)
			vputils.CheckAndPanic(err)

			vStack = vStack.push(value)
			pc += 1                   // the opcode
			pc += bytesPerDataAddress // the target address
		case 0x51:
			// POP.B Address
		case 0x08:
			// OUT.B (implied stack)
			c, err := vStack.top()
			vputils.CheckAndPanic(err)

			vStack, err = vStack.pop()
			vputils.CheckAndPanic(err)

			fmt.Print(string(c))
			pc += 1 // the opcode
		default:
			// invalid opcode
			fmt.Printf("Invalid opcode %02x at %04x\n", opcode, pc)
			return
		}
	}

	fmt.Printf("Execution halted at %04x\n", pc)
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("No module file specified")
		os.Exit(1)
	}

	moduleFile := args[0]
	fmt.Printf("Opening file '%s'\n", moduleFile)

	f, err := os.Open(moduleFile)
	vputils.CheckAndPanic(err)

	defer f.Close()

	header := vputils.ReadString(f)
	if header != "module" {
		fmt.Println("Did not find module header")
		return
	}

	header = vputils.ReadString(f)
	if header != "properties" {
		fmt.Println("Did not find properties header")
		return
	}

	properties := vputils.ReadTextTable(f)

	codeWidth := 0
	dataWidth := 0
	for _, nameValue := range properties {
		shortName := strings.Replace(nameValue.Name, " ", "", -1)
		if shortName == "CODEADDRESSWIDTH" {
			codeWidth = 1
		}
		if shortName == "DATAADDRESSWIDTH" {
			dataWidth = 1
		}
	}

	header = vputils.ReadString(f)
	if header != "code" {
		fmt.Println("Did not find code header")
		return
	}

	code := vputils.ReadBinaryBlock(f, codeWidth)

	header = vputils.ReadString(f)
	if header != "data" {
		fmt.Println("Did not find data header")
		return
	}

	data := vputils.ReadBinaryBlock(f, dataWidth)

	executeCode(code, data)
}
