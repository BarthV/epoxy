// Copyright 2015 Netflix, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package binprot

import (
	"errors"

	"github.com/netflix/rend/common"
)

var ErrBadMagic = errors.New("Bad magic value")

const (
	MagicRequest  = uint8(0x80)
	MagicResponse = uint8(0x81)

	// All opcodes as defined in memcached
	// Minus SASL and range ops
	OpcodeGet        = uint8(0x00)
	OpcodeSet        = uint8(0x01)
	OpcodeAdd        = uint8(0x02)
	OpcodeReplace    = uint8(0x03)
	OpcodeDelete     = uint8(0x04)
	OpcodeIncrement  = uint8(0x05)
	OpcodeDecrement  = uint8(0x06)
	OpcodeQuit       = uint8(0x07)
	OpcodeFlush      = uint8(0x08)
	OpcodeGetQ       = uint8(0x09)
	OpcodeNoop       = uint8(0x0a)
	OpcodeVersion    = uint8(0x0b)
	OpcodeGetK       = uint8(0x0c)
	OpcodeGetKQ      = uint8(0x0d)
	OpcodeAppend     = uint8(0x0e)
	OpcodePrepend    = uint8(0x0f)
	OpcodeStat       = uint8(0x10)
	OpcodeSetQ       = uint8(0x11)
	OpcodeAddQ       = uint8(0x12)
	OpcodeReplaceQ   = uint8(0x13)
	OpcodeDeleteQ    = uint8(0x14)
	OpcodeIncrementQ = uint8(0x15)
	OpcodeDecrementQ = uint8(0x16)
	OpcodeQuitQ      = uint8(0x17)
	OpcodeFlushQ     = uint8(0x18)
	OpcodeAppendQ    = uint8(0x19)
	OpcodePrependQ   = uint8(0x1a)
	OpcodeTouch      = uint8(0x1c)
	OpcodeGat        = uint8(0x1d)
	OpcodeGatQ       = uint8(0x1e)
	OpcodeGatK       = uint8(0x23)
	OpcodeGatKQ      = uint8(0x24)
	OpcodeInvalid    = uint8(0xFF)

	OpcodeGetE  = uint8(0x40)
	OpcodeGetEQ = uint8(0x41)

	StatusSuccess        = uint16(0x00)
	StatusKeyEnoent      = uint16(0x01)
	StatusKeyExists      = uint16(0x02)
	StatusE2big          = uint16(0x03)
	StatusEinval         = uint16(0x04)
	StatusNotStored      = uint16(0x05)
	StatusDeltaBadval    = uint16(0x06)
	StatusAuthError      = uint16(0x20)
	StatusAuthContinue   = uint16(0x21)
	StatusUnknownCommand = uint16(0x81)
	StatusEnomem         = uint16(0x82)
	StatusNotSupported   = uint16(0x83)
	StatusInternalError  = uint16(0x84)
	StatusBusy           = uint16(0x85)
	StatusTempFailure    = uint16(0x86)
	StatusInvalid        = uint16(0xFFFF)
)

func DecodeError(header ResponseHeader) error {
	switch header.Status {
	case StatusKeyEnoent:
		return common.ErrKeyNotFound
	case StatusKeyExists:
		return common.ErrKeyExists
	case StatusE2big:
		return common.ErrValueTooBig
	case StatusEinval:
		return common.ErrInvalidArgs
	case StatusNotStored:
		return common.ErrItemNotStored
	case StatusDeltaBadval:
		return common.ErrBadIncDecValue
	case StatusAuthError:
		return common.ErrAuth
	case StatusUnknownCommand:
		return common.ErrUnknownCmd
	case StatusEnomem:
		return common.ErrNoMem
	case StatusNotSupported:
		return common.ErrNotSupported
	case StatusInternalError:
		return common.ErrInternal
	case StatusBusy:
		return common.ErrBusy
	case StatusTempFailure:
		return common.ErrTempFailure
	}
	return nil
}

func errorToCode(err error) uint16 {
	switch err {
	case common.ErrKeyNotFound:
		return StatusKeyEnoent
	case common.ErrKeyExists:
		return StatusKeyExists
	case common.ErrValueTooBig:
		return StatusE2big
	case common.ErrInvalidArgs:
		return StatusEinval
	case common.ErrItemNotStored:
		return StatusNotStored
	case common.ErrBadIncDecValue:
		return StatusDeltaBadval
	case common.ErrAuth:
		return StatusAuthError
	case common.ErrUnknownCmd:
		return StatusUnknownCommand
	case common.ErrNoMem:
		return StatusEnomem
	case common.ErrNotSupported:
		return StatusNotSupported
	case common.ErrInternal:
		return StatusInternalError
	case common.ErrBusy:
		return StatusBusy
	case common.ErrTempFailure:
		return StatusTempFailure
	}
	return StatusInvalid
}
