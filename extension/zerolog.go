package extension

import "github.com/rs/zerolog"

type Bytes []byte

func (slice Bytes) MarshalZerologArray(arr *zerolog.Array) {
	arr.Bytes(slice)
}

type IntSlice []int

func (slice IntSlice) MarshalZerologArray(arr *zerolog.Array) {
	for _, i := range slice {
		arr.Int(i)
	}
}

type Int32Slice []int32

func (slice Int32Slice) MarshalZerologArray(arr *zerolog.Array) {
	for _, i := range slice {
		arr.Int32(i)
	}
}

type Int64Slice []int64

func (slice Int64Slice) MarshalZerologArray(arr *zerolog.Array) {
	for _, i := range slice {
		arr.Int64(i)
	}
}

type StringSlice []string

func (slice StringSlice) MarshalZerologArray(arr *zerolog.Array) {
	for _, i := range slice {
		arr.Str(i)
	}
}

type UintSlice []uint

func (slice UintSlice) MarshalZerologArray(arr *zerolog.Array) {
	for _, i := range slice {
		arr.Uint(i)
	}
}

type Uint32Slice []uint32

func (slice Uint32Slice) MarshalZerologArray(arr *zerolog.Array) {
	for _, i := range slice {
		arr.Uint32(i)
	}
}

type Uint64Slice []uint64

func (slice Uint64Slice) MarshalZerologArray(arr *zerolog.Array) {
	for _, i := range slice {
		arr.Uint64(i)
	}
}
