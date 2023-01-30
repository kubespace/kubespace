package datatype

type DataType interface {
	Unmarshal([]byte) (interface{}, error)
}

var StringType = &stringType{}

type stringType struct{}

func (s *stringType) Unmarshal(data []byte) (interface{}, error) {
	return string(data), nil
}
