package main

import (
	"context"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/wavesplatform/gowaves/pkg/client"
	"github.com/wavesplatform/gowaves/pkg/crypto"
	"github.com/wavesplatform/gowaves/pkg/proto"
)

type Monitor struct {
	Balance      uint64
	BalanceAnote uint64
	Items        map[string]float64
	ItemsAnote   map[string]float64
}

func (m *Monitor) start() {
	for {
		total := 0
		m.Items = make(map[string]float64)
		m.ItemsAnote = make(map[string]float64)
		m.getBalance()

		cl, err := client.NewClient(client.Options{BaseUrl: AnoteNodeURL, Client: &http.Client{}})
		if err != nil {
			log.Println(err)
			logTelegram(err.Error())
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

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
			a := strings.Split(s.GetKey(), SEP)[1]
			v := s.ToProtobuf().GetStringValue()
			if strings.Contains(v, a) {
				am, _ := strconv.Atoi(strings.Split(v, SEP)[1])
				total += am
			}
		}

		for _, s := range stakes {
			a := strings.Split(s.GetKey(), SEP)[1]
			v := s.ToProtobuf().GetStringValue()
			if strings.Contains(v, a) {
				am, _ := strconv.Atoi(strings.Split(v, SEP)[1])
				if am > 0 {
					m.Items[a] = float64(am) / float64(total)
					// items = append()
				}
			}
		}

		total = 0

		ctx, cancelanote := context.WithTimeout(context.Background(), 30*time.Second)

		addr, err = proto.NewAddressFromString(StakeAddressAnote)
		if err != nil {
			log.Println(err)
			logTelegram(err.Error())
		}

		stakes, _, err = cl.Addresses.AddressesData(ctx, addr)
		if err != nil {
			log.Println(err)
			logTelegram(err.Error())
		}

		for _, s := range stakes {
			a := strings.Split(s.GetKey(), SEP)[1]
			v := s.ToProtobuf().GetStringValue()
			if strings.Contains(v, a) {
				am, _ := strconv.Atoi(strings.Split(v, SEP)[1])
				total += am
			}
		}

		for _, s := range stakes {
			a := strings.Split(s.GetKey(), SEP)[1]
			v := s.ToProtobuf().GetStringValue()
			if strings.Contains(v, a) {
				am, _ := strconv.Atoi(strings.Split(v, SEP)[1])
				if am > 0 {
					m.ItemsAnote[a] = float64(am) / float64(total)
				}
			}
		}

		m.processItems()

		m.processItemsAnote()

		log.Println("Done payouts.")

		time.Sleep(time.Second * MonitorTick)

		cancel()
		cancelanote()
	}
}

func (m *Monitor) getBalance() {
	cl, err := client.NewClient(client.Options{BaseUrl: AnoteNodeURL, Client: &http.Client{}})
	if err != nil {
		log.Println(err)
		logTelegram(err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	addr := proto.MustAddressFromString(StakeAddress)

	br, _, err := cl.Addresses.Balance(ctx, addr)
	if err != nil {
		log.Println(err)
		logTelegram(err.Error())
		return
	}

	m.Balance = br.Balance
}

func (m *Monitor) processItems() {
	fee := Fee + (uint64(len(m.Items)) * MassFee)
	if fee%Fee != 0 {
		fee += MassFee
	}
	amount := (m.Balance / 2) - fee

	pk := crypto.MustPublicKeyFromBase58(conf.PublicKey)
	sk := crypto.MustSecretKeyFromBase58(conf.PrivateKey)
	ts := time.Now().Unix() * 1000
	as, _ := proto.NewOptionalAssetFromString("")

	var tr []proto.MassTransferEntry

	for a, i := range m.Items {
		am := uint64(math.Floor(float64(amount) * i))
		addr := proto.MustAddressFromString(a)
		mte := proto.MassTransferEntry{
			Recipient: proto.NewRecipientFromAddress(addr),
			Amount:    am,
		}
		tr = append(tr, mte)
	}

	t := proto.NewUnsignedMassTransferWithProofs(2, pk, *as, tr, fee, uint64(ts), nil)

	err := t.Sign(55, sk)
	if err != nil {
		log.Println(err)
		logTelegram(err.Error())
	}

	client, err := client.NewClient(client.Options{BaseUrl: AnoteNodeURL, Client: &http.Client{}})
	if err != nil {
		log.Println(err)
		logTelegram(err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = client.Transactions.Broadcast(ctx, t)
	if err != nil {
		log.Println(err)
		// logTelegram(err.Error())
	}
}

func (m *Monitor) processItemsAnote() {
	fee := Fee + (uint64(len(m.ItemsAnote)) * MassFee)
	if fee%Fee != 0 {
		fee += MassFee
	}
	amount := (m.Balance / 2) - fee

	pk := crypto.MustPublicKeyFromBase58(conf.PublicKey)
	sk := crypto.MustSecretKeyFromBase58(conf.PrivateKey)
	ts := time.Now().Unix() * 1000
	as, _ := proto.NewOptionalAssetFromString("")

	var tr []proto.MassTransferEntry
	counter := 1

	for a, i := range m.ItemsAnote {
		am := uint64(math.Floor(float64(amount) * i))
		addr := proto.MustAddressFromString(a)
		mte := proto.MassTransferEntry{
			Recipient: proto.NewRecipientFromAddress(addr),
			Amount:    am,
		}
		tr = append(tr, mte)

		if counter%100 == 0 || counter == len(m.ItemsAnote) {
			t := proto.NewUnsignedMassTransferWithProofs(2, pk, *as, tr, fee, uint64(ts), nil)

			err := t.Sign(55, sk)
			if err != nil {
				log.Println(err)
				logTelegram(err.Error())
			}

			client, err := client.NewClient(client.Options{BaseUrl: AnoteNodeURL, Client: &http.Client{}})
			if err != nil {
				log.Println(err)
				logTelegram(err.Error())
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			_, err = client.Transactions.Broadcast(ctx, t)
			if err != nil {
				log.Println(err)
				// logTelegram(err.Error())
			}
		}

		counter++
	}
}

func initMonitor() {
	m := &Monitor{}
	m.start()
}
