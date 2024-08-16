package extension

import "github.com/rs/zerolog"

func ZerolifyBytes(bb []byte) *zerolog.Array {
	arr := zerolog.Arr().Bytes(bb)
	return arr
}

func ZerolifyInts(ii []int) *zerolog.Array {
	arr := zerolog.Arr()
	for _, i := range ii {
		arr.Int(i)
	}
	return arr
}

func ZerolifyInt32s(ii []int32) *zerolog.Array {
	arr := zerolog.Arr()
	for _, i := range ii {
		arr.Int32(i)
	}
	return arr
}

func ZerolifyInt64s(ii []int64) *zerolog.Array {
	arr := zerolog.Arr()
	for _, i := range ii {
		arr.Int64(i)
	}
	return arr
}

func ZerolifyUints(uu []uint) *zerolog.Array {
	arr := zerolog.Arr()
	for _, u := range uu {
		arr.Uint(u)
	}
	return arr
}

func ZerolifyUint32s(uu []uint32) *zerolog.Array {
	arr := zerolog.Arr()
	for _, u := range uu {
		arr.Uint32(u)
	}
	return arr
}

func ZerolifyUint64s(uu []uint64) *zerolog.Array {
	arr := zerolog.Arr()
	for _, u := range uu {
		arr.Uint64(u)
	}
	return arr
}

func ZerolifyStrings(ss []string) *zerolog.Array {
	arr := zerolog.Arr()
	for _, s := range ss {
		arr.Str(s)
	}
	return arr
}
