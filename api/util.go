package ninja

import (
	"crypto/md5"
	"encoding/hex"
)

func getGUID(in ...string) string {
	h := md5.New()
	for _, s := range in {
		h.Write([]byte(s))
	}
	str := hex.EncodeToString(h.Sum(nil))
	return str[:10]
}
