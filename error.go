package errutil

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
)

const (
	NoCallStack = "$nocs"
	OnlyFuncInfo = "$onlyfunc"
	FullCallStack = "$fullcs"

	MoreInfo = "$moreinfo"
)

type Error struct {
	typ   error
	where map[string]string
	inner error
}

// ErrorPrintFn should prints type, callstack, info, and inner error.
var ErrorPrinter func(error, map[string]string, error) string
var DefaultCallStackLevel string = OnlyFuncInfo

func New(t error, s ...string) *Error {
	e := &Error{ typ: t, where: map[string]string{} }
	return e.addInfo(s...).addCallStack(2)
}

func Embed(t error, inner error, s ...string) *Error {
	e := &Error{ typ: t, where: map[string]string{}, inner: inner }
	return e.addInfo(s...).addCallStack(2)
}

func AddInfo(e error, s ...string) *Error {
	ee, ok := e.(*Error)
	if ok {
		ee.addInfo(s...).addCallStack(2)
	} else {
		ee = New(e, s...)
	}
	return ee
}

func (e *Error) AddCallStack() *Error {
	return e.addCallStack(2)
}

func (e *Error) Error() string {
	return ErrorPrinter(e.typ, e.where, e.inner)
}

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

func (e *Error) addCallStack(skip int) *Error {
	cslevel, exists := e.where[callStackLevelKey]
	if ! exists {
		cslevel = DefaultCallStackLevel
	}
	var callstack string
	if cslevel == OnlyFuncInfo {
		pc := make([]uintptr, 1)
		count := runtime.Callers(skip+1, pc)
		var fn, file string
		var line int
		if count >= 1 {
			f := runtime.FuncForPC(pc[0])
			fn = f.Name()
			fn = fn[strings.LastIndex(fn, "/")+1:]
			file, line = f.FileLine(pc[0])
		}
		callstack = fmt.Sprintf("%s in %s:%d", fn, file, line)
	} else if cslevel == FullCallStack {
		callstack = callStack(skip+1)
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

func callStack(skip int) string {
	buf := debug.Stack()
	i := 0
	for skipped := 0; skipped < (skip+1) * 2 + 1; skipped++ {
		for ; i < len(buf); i++ {
			if buf[i] == '\n' {
				i++
				break
			}
		}
		if i >= len(buf) {
			i = len(buf) - 1
			break
		}
	}
	return string(buf[i:])
}

func defaultErrorPrinter(t error, m map[string]string, inner error) string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(t.Error())
	callstack, exists := m[callStackKey]
	if exists {
		buf.WriteString(" at: ")
		buf.WriteString(callstack)
	}
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
	if inner != nil {
		buf.WriteString(", inner error: ")
		buf.WriteString(inner.Error())
	}

	return string(buf.Bytes())
}

func init() {
	ErrorPrinter = defaultErrorPrinter
}
