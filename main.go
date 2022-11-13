package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/wavesplatform/gowaves/pkg/client"
	"github.com/wavesplatform/gowaves/pkg/proto"
)

var conf *Config

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	conf = initConfig()

	cl, err := client.NewClient(client.Options{BaseUrl: AnoteNodeURL, Client: &http.Client{}})
	if err != nil {
		log.Println(err)
		logTelegram(err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	addr, err := proto.NewAddressFromString(StakeAddress)
	if err != nil {
		log.Println(err)
		logTelegram(err.Error())
	}

	stakes, _, err := cl.Addresses.AddressesData(ctx, addr)
	if err != nil {
		log.Println(err)
		logTelegram(err.Error())
	}

	for _, s := range stakes {
		nv := "%d%s__" + fmt.Sprintf("%d", s.ToProtobuf().GetIntValue()) + "__"
		if isNode(s.GetKey()) {
			nv += nodeOwner(strings.Split(s.GetKey(), "__")[1])
		} else {
			nv += strings.Split(s.GetKey(), "__")[1]
		}
		log.Println(s.GetKey() + " => " + nv)

		dataTransaction(s.GetKey(), &nv, nil, nil)
	}
}
