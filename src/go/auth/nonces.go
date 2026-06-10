package auth

import (
	"fmt"
	"m2cpcli/tools"
	"strconv"
	"sync"
	"time"
)

type Nonces struct {
	sync.Mutex
	nonces []string
	MaxAge int64
}

type SignedNonce struct {
	StoreName string `json:"store_name"`
	Account   string `json:"account"`
	Nonce     string `json:"nonce"`
	Format    string `json:"format"`
	Signature string `json:"signature"`
}

func (n *Nonces) Make() string {
	n.purgeOld()
	n.Lock()
	defer n.Unlock()
	nonce := fmt.Sprintf("%s:%d", tools.RandomDigits(29), time.Now().Unix())
	n.nonces = append(n.nonces, nonce)
	return nonce
}

func (n *Nonces) Validate(nonce string) bool {
	n.purgeOld()
	n.Lock()
	defer n.Unlock()
	for i := len(n.nonces) - 1; i >= 0; i-- {
		if n.nonces[i] == nonce {
			n.nonces = append(n.nonces[:i], n.nonces[i+1:]...)
			return true
		}
	}
	return false
}

func (n *Nonces) purgeOld() {
	n.Lock()
	defer n.Unlock()
	if n.MaxAge == 0 {
		n.MaxAge = 60
	}
	tDrop := time.Now().Unix() - n.MaxAge
	for i := len(n.nonces) - 1; i >= 0; i-- {
		t, _ := strconv.ParseInt(n.nonces[i][30:], 10, 64)
		if t < tDrop {
			if i == len(n.nonces)-1 {
				n.nonces = []string{}
			} else {
				n.nonces = n.nonces[i+1:]
			}
			return
		}
	}
}
