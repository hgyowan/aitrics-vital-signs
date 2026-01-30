package error

import (
	"fmt"
	"runtime"
)

var funcInfoFormat = "{%s:%d} [%s]"

func getFuncInfo(skip int) string {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}

	fn := runtime.FuncForPC(pc)
	name := "unknown"
	if fn != nil {
		name = fn.Name()
	}

	return fmt.Sprintf(funcInfoFormat, file, line, name)
}

func wrapFormat(err error, msg string, skip int) error {
	if err == nil {
		return nil
	}

	loc := getFuncInfo(skip)
	if loc == "" {
		if msg == "" {
			return err
		}
		return fmt.Errorf("%s: %w", msg, err)
	}

	if msg != "" && msg != err.Error() {
		loc = loc + " " + msg
	}

	return fmt.Errorf("%s | %w", loc, err)
}

func wrap(err error) error {
	return wrapFormat(err, "", 4)
}

func wrapBusiness(err error, msg string) error {
	return wrapFormat(err, msg, 5)
}
