package util

import (
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
)

// JsonToMap Convert json string to map
func JsonToMap(jsonStr string) (map[string]int64, error) {
	m := make(map[string]int64)
	err := json.Unmarshal([]byte(jsonStr), &m)
	if err != nil {
		fmt.Printf("Unmarshal with error: %+v\n", err)
		return nil, err
	}
	for k, v := range m {
		fmt.Printf("%v: %v\n", k, v)
	}
	return m, nil
}

// HasElem 元素是否在数组中存在
func HasElem(elem interface{}, slice interface{}) bool {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("HasElem error %s at  %s", err, AtWhere())
		}
	}()
	arrV := reflect.ValueOf(slice)
	if arrV.Kind() == reflect.Slice || arrV.Kind() == reflect.Array {
		for i := 0; i < arrV.Len(); i++ {
			// XXX - panics if slice element points to an unexported struct field
			// see https://golang.org/pkg/reflect/#Value.Interface
			if arrV.Index(i).Interface() == elem {
				return true
			}
		}
	}
	return false
}

// AtWhere return the parent function name.
func AtWhere() string {
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		return runtime.FuncForPC(pc).Name()
	} else {
		return "Method not Found!"
	}
}
