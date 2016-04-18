package errutil

import (
	"runtime"
	"strconv"
)

func CallerStr(skip int) string {
	return string(locationBytes(CallerInfo(skip+1)))
}

func CallerInfo(skip int) (string, string, int) {
	// \todo tidy up file path
	// most go programs live under GOPATH, and therefore
	// the full function name and filename has parts in common.
	// and it is desirable to trim the beginning file path,
	// if it starts with GOPATH
	//
	// the following is a workaround to get neat filename,
	// from stack.go of log15 pkg.
	// unfortunately, it can handle only the cases where the src
	// file is under GOPATH
	// so I need to find a way to get GOPATH.
	//
	//const sep = "/"
	//impCnt := strings.Count(fn.Name(), sep) + 1
	//pathCnt := strings.Count(file, sep)
	//for pathCnt > impCnt {
	//	i := strings.Index(file, sep)
	//	if i == -1 {
	//		break
	//	}
	//	file = file[i+len(sep):]
	//	pathCnt--
	//}

	pc := make([]uintptr, 1)
	count := runtime.Callers(skip+2, pc)
	var fn, file string
	var line int
	if count >= 1 {
		f := runtime.FuncForPC(pc[0])
		fn = f.Name()
		//fn = fn[strings.LastIndex(fn, "/")+1:]
		file, line = f.FileLine(pc[0])
	}
	return fn, file, line
}

func CallStack(skip int) string {
	b := []byte{}
	pc := make([]uintptr, 512)
	count := runtime.Callers(skip+2, pc)
	for i := 0; i < count; i++ {
		f := runtime.FuncForPC(pc[i])
		fn := f.Name()
		file, line := f.FileLine(pc[i])
		if i > 0 {
			b = append(b, []byte(" <== ")...)
		}
		b = append(b, locationBytes(fn, file, line)...)
	}
	return string(b)
}



////////////////////////////////////////////////////////////////////

func locationBytes(fn string, file string, line int) []byte {
	b := []byte{}
	b = append(b, []byte(file)...)
	b = append(b, ':')
	b = strconv.AppendInt(b, int64(line), 10)
	b = append(b, ':')
	b = append(b, []byte(fn)...)
	return b
}