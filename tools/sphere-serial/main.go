package main

import (
	"bytes"
	"encoding/base32"
	"encoding/hex"
	"io/ioutil"
	"os"
	"strings"
)

func main() {

	input, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		panic(err)
	}

	serial, err := hex.DecodeString(string(bytes.Split(input, []byte("hwserial="))[1][0:16]))
	if err != nil {
		panic(err)
	}

	b32Serial := strings.TrimRight(base32.StdEncoding.EncodeToString(serial), "=")

	os.Stdout.Write([]byte(b32Serial))
}
