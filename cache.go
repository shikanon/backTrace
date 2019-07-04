package backTrace

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

type CacheStock interface {
	Ready([]string) error
	Load(string) (*StockColumnData, error)
	Delete(string)
}

type StockMap struct {
	lock  sync.RWMutex
	stock sync.Map
}

func (s *StockMap) Load(code string) (*StockColumnData, error) {
	var stockpoint *StockColumnData
	var err error
	value, ok := s.stock.Load(code)
	if ok {
		stockpoint, ok = value.(*StockColumnData)
		if !ok {
			return nil, errors.New("StockMap store type is error and can't load!" + reflect.TypeOf(value).Name())
		}
		return stockpoint, err
	}
	// not ok:
	s.lock.Lock()
	defer s.lock.Unlock()
	value, ok = s.stock.Load(code)
	if ok {
		stockpoint, ok = value.(*StockColumnData)
		if !ok {
			return nil, errors.New("StockMap store type is error and can't load!" + reflect.TypeOf(value).Name())
		}
		return stockpoint, err
	}
	stock, err := GetSockData(code)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("get stock %s error!%v", code, err))
	}
	if len(stock.Date) == 0 {
		return nil, errors.New("stock is nil!")
	}
	stockpoint = &stock
	s.stock.Store(code, stockpoint)

	return stockpoint, err
}

func (s *StockMap) Ready(codes []string) error {
	wg := sync.WaitGroup{}
	for _, c := range codes {
		_, ok := s.stock.Load(c)
		if !ok {
			wg.Add(1)
			go func(code string) error {
				stock, err := GetSockData(code)
				if err != nil {
					return err
				}
				s.stock.Store(code, &stock)
				wg.Done()
				return nil
			}(c)
		}
	}
	wg.Wait()
	return nil
}

func (s *StockMap) Delete(code string) {
	s.stock.Delete(code)
}

func (s *StockMap) Flushall(codes []string) {
	for _, c := range codes {
		go s.stock.Delete(c)
	}
}
