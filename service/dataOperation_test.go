package service

import (
	"testing"
)

func TestDataOperation(t *testing.T) {
	//initialize.MongoInit()
	//DataOperation(time.Minute*10)
	s := []byte{170,85,1, 3 ,2, 0, 86, 56, 122,0,0,0}
	ParseDTUData(s,15)
}