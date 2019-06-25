package backTrace

import (
	"errors"
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

func (r *StrategyRegister) Load(name string) (strategy Strategy, err error) {
	value, ok := r.Value.Load(name)
	if !ok {
		return nil, errors.New("Loaded failed! " + name + " is not in StrategyRegister.")
	}
	switch strag := value.(type) {
	case BreakOutStrategyBuy:
		strategy = &strag
	case BreakOutStrategySell:
		strategy = &strag
	case KDJtStrategyBuy:
		strategy = &strag
	case KDJtStrategySell:
		strategy = &strag
	}
	if strategy == nil {
		return nil, errors.New("Error type in StrategyRegister! is not strategy, so can't find it!")
	}
	return strategy, nil
}
