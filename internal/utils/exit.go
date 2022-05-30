package utils

import (
	"os"
	"reflect"

	"github.com/1azunna/zapgo/internal/types"
)

func SetExitCode(riskcount types.RiskCount, gate string) {
	v := reflect.ValueOf(riskcount)
	typeV := v.Type()
	var fail bool
	for i := 0; i < v.NumField(); i++ {
		arr := v.Field(i).Interface().([]int)
		if typeV.Field(i).Name == gate {
			if CountSum(arr) > 0 {
				fail = true
			} else {
				for x := i; x < v.NumField(); x++ {
					arr := v.Field(x).Interface().([]int)
					if CountSum(arr) > 0 {
						fail = true
					}
				}
			}
		}
	}
	if fail {
		os.Exit(1)
	}
}
