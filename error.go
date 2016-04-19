package errutil

import (
	"bytes"
	"errors"
	"os"
)

const (
	NoCallStack = "$nocs"
	OnlyFuncInfo = "$onlyfunc"
	FullCallStack = "$fullcs"

	MoreInfo = "$moreinfo"
)

var ErrAssert error

type Error struct {
	typ   error
	inner error
	where map[string]string
}

// ErrorPrintFn should prints type, callstack, info, and inner error.
var ErrorPrinter func(error, map[string]string, error) string
var DefaultCallStackLevel string = OnlyFuncInfo

func New(t error, s ...string) *Error {
	e := &Error{ typ: t, where: map[string]string{} }
	return e.addInfo(s...).addCallStack(DefaultCallStackLevel, 1)
}

func NewAssert(s ...string) *Error {
	e := &Error{ typ: ErrAssert, where: map[string]string{} }
	return e.addInfo(s...).addCallStack(FullCallStack, 1)
}

func Embed(t error, inner error, s ...string) *Error {
	e := &Error{ typ: t, where: map[string]string{}, inner: inner }
	return e.addInfo(s...).addCallStack(DefaultCallStackLevel, 1)
}

func AssertEmbed(inner error, s ...string) *Error {
	e := &Error{ typ: ErrAssert, where: map[string]string{}, inner: inner }
	return e.addInfo(s...).addCallStack(FullCallStack, 1)
}

func AddInfo(e error, s ...string) *Error {
	ee, ok := e.(*Error)
	if ok {
		ee.addInfo(s...).addCallStack(DefaultCallStackLevel, 1)
	} else {
		ee = New(e, s...)
	}
	return ee
}

func (e *Error) AddCallStack(lvl string) *Error {
	return e.addCallStack(lvl, 1)
}

func (e *Error) Error() string {
	return ErrorPrinter(e.typ, e.where, e.inner)
}

// CompareType checks whether error e is of type t.
func CompareType(e error, t error) bool {
	ee, ok := e.(*Error)
	if ok {
		return ee.typ == t
	} else {
		return e == t
	}
}

// IsNotExist corresponds to os.IsNotExist
func IsNotExist(e error) bool {
	ee, ok := e.(*Error)
	if ok {
		return os.IsNotExist(ee.typ)
	} else {
		return os.IsNotExist(e)
	}
}



////////////////////////////////////////////////////////////////////

func (e *Error) addInfo(s ...string) *Error {
	m := e.where
	for i := 0; i < len(s); i++ {
		k := s[i]
		vv, exists := keysWithoutValue[k]
		if exists {
			m[vv[0]]=vv[1]
			continue
		} else {
			var v string
			if i + 1 < len(s) {
				v = s[i + 1]
			}
			m[k] = v
			i++
		}
	}
	return e
}

// addCallStack adds callstack if it has not been added already
func (e *Error) addCallStack(lvl string, skip int) *Error {
	_, exists := e.where[callStackKey]
	if exists {
		return e
	}
	var callstack string
	if lvl == OnlyFuncInfo {
		callstack = CallerStr(skip+1)
	} else if lvl == FullCallStack {
		callstack = CallStack(skip+1)
	}
	e.where[callStackKey] = callstack
	return e
}

var keysWithoutValue map[string][2]string = map[string][2]string{
	NoCallStack: [2]string{ callStackLevelKey, NoCallStack },
	OnlyFuncInfo: [2]string{ callStackLevelKey, OnlyFuncInfo },
	FullCallStack: [2]string{ callStackLevelKey, FullCallStack },
}
const specialKeyMark = "$"
const callStackKey = "$callstack"
const callStackLevelKey = "$cslevel"

func defaultErrorPrinter(t error, m map[string]string, inner error) string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(t.Error())
	v, exists := m[MoreInfo]
	if exists {
		buf.WriteString(", ")
		buf.WriteString(v)
	}
	for k, v := range m {
		if len(k) > 0 && k[0:1] == specialKeyMark {
			continue
		}
		buf.WriteString(", ")
		buf.WriteString(k)
		buf.WriteString(": ")
		buf.WriteString(v)
	}
	callstack, exists := m[callStackKey]
	if exists && len(callstack) > 0 {
		buf.WriteString(", at: ")
		buf.WriteString(callstack)
	}
	if inner != nil {
		buf.WriteString(", inner error: ")
		buf.WriteString(inner.Error())
	}
	return string(buf.Bytes())
}

func init() {
	ErrorPrinter = defaultErrorPrinter
	ErrAssert = errors.New("Something wrong")
}
