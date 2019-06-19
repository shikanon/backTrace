package backTrace

import (
	"fmt"
	"reflect"
	"sync"
)

type StrategyRegister struct {
	Value sync.Map
	Names []string
}

func (r *StrategyRegister) Regist(s interface{}) {
	var name string
	typeStrategy := reflect.TypeOf(s)
	valueStrategy := reflect.ValueOf(s)
	name = typeStrategy.Name()
	for i := 0; i < typeStrategy.NumField(); i++ {
		field := typeStrategy.Field(i)
		fieldValue := reflect.Indirect(valueStrategy).FieldByName(field.Name)
		name = name + "/" + fmt.Sprintf("%s:%v", field.Name, fieldValue)
	}
	r.Names = append(r.Names, name)
	r.Value.Store(name, s)
}
