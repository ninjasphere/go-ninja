package logger

import "testing"

func TestDeviceReaderOpen(t *testing.T) {
	log := GetLogger("test")
	log.Errorf("some test %s", "woot")
}
