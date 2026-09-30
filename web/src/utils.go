package main

import "strconv"

func format32BitHex(u uint32) string {
	res := strconv.FormatUint(uint64(u), 16)
	for len(res) < 8 {
		res = "0" + res
	}
	return "0x" + res
}

func format8BitHex(u uint8) string {
	res := strconv.FormatUint(uint64(u), 16)
	for len(res) < 2 {
		res = "0" + res
	}
	return res
}
