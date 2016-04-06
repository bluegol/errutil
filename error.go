package errutil

import (
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
var ErrorPrintFn func(error, map[string]string, error) string
var DefaultCallStackLevel string = OnlyFuncInfo

func AddInfo(e error, s ...string) *Error {
	ee, ok := e.(*Error)
	if ok {
		ee.addInfo(s...).addCallStack(2)
	} else {
		ee = New(e, s...)
	}
	return ee
}

func New(t error, s ...string) *Error {
	e := &Error{ typ: t, where: map[string]string{} }
	return e.addInfo(s...).addCallStack(2)
}

func Embed(t error, inner error, s ...string) *Error {
	e := &Error{ typ: t, where: map[string]string{}, inner: inner }
	return e.addInfo(s...).addCallStack(2)
}

func CompareType(e error, t error) bool {
	ee, ok := e.(*Error)
	if ok {
		return ee.typ == t
	} else {
		return e == t
	}
}

func IsNotExist(e error) bool {
	ee, ok := e.(*Error)
	if ok {
		return os.IsNotExist(ee.typ)
	} else {
		return os.IsNotExist(e)
	}
}

func (e *Error) Error() string {
	return ErrorPrintFn(e.typ, e.where, e.inner)
}

func (e *Error) AddCallStack() *Error {
	return e.addCallStack(2)
}



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

func init() {
	ErrorPrintFn = func(t error, m map[string]string, inner error) string {
		buf := make([]byte, 0, 1024)
		buf = append(buf, []byte(t.Error())...)
		callstack := m[callStackKey]
		if len(callstack) > 0 {
			buf = append(buf, []byte(" at: ")...)
			buf = append(buf, []byte(callstack)...)
		}
		v, exists := m[MoreInfo]
		if exists {
			buf = append(buf, ',', ' ')
			buf = append(buf, []byte(v)...)
		}
		if inner != nil {
			buf = append(buf, []byte(", inner error: ")...)
			buf = append(buf, []byte(inner.Error())...)
		}
		for k, v := range m {
			if len(k) > 0 && k[0:1] == specialKeyMark {
				continue
			}
			buf = append(buf, ',', ' ')
			buf = append(buf, []byte(k)...)
			buf = append(buf, ':', ' ')
			buf = append(buf, []byte(v)...)
		}
		return string(buf)
	}
}
