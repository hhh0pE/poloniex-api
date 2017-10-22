package poloniex

import (
	"encoding/json"
	"log"
	"time"

	"github.com/hhh0pE/ggm"
	"github.com/k0kubun/pp"
	"gopkg.in/beatgammit/turnpike.v2"
)

type (
	//WSTicker describes a ticker item
	WSTicker struct {
		Pair          string
		Last          ggm.Decimal
		Ask           ggm.Decimal
		Bid           ggm.Decimal
		PercentChange ggm.Decimal
		BaseVolume    ggm.Decimal
		QuoteVolume   ggm.Decimal
		IsFrozen      bool
		DailyHigh     ggm.Decimal
		DailyLow      ggm.Decimal
	}

	// WSTickerChan is a onduit through which WSTicker items are sent
	WSTickerChan chan WSTicker

	//WSTrade describes a trade, a new order, or an order update
	WSTrade struct {
		TradeID string
		Rate    ggm.Decimal `json:",string"`
		Amount  ggm.Decimal `json:",string"`
		Type    string
		Date    string
		TS      time.Time
	}

	//WSOrderOrTrade is a slice of WSTrades with an indicator of the type (trade, new order, update order)
	WSOrderOrTrade struct {
		Seq    int64
		Orders WSOrders
	}

	WSOrders []struct {
		Data WSTrade
		Type string
	}

	// WSOrderOrTradeChan is a onduit through which WSTicker items are sent
	WSOrderOrTradeChan chan WSOrderOrTrade
)

const (
	//SENTINEL is used to mark items without a sequence number
	SENTINEL = int64(-1)
)

//SubscribeTicker subscribes to the ticker feed and returns a channel over which it will send updates
func (p *Poloniex) SubscribeTicker() WSTickerChan {
	p.InitWS()
	p.subscribedTo["ticker"] = true
	ch := make(WSTickerChan)
	p.ws.Subscribe("ticker", p.makeTickerHandler(ch))
	return ch
}

//SubscribeOrder subscribes to the order and trade feed and returns a channel over which it will send updates
func (p *Poloniex) SubscribeOrder(code string) WSOrderOrTradeChan {
	p.InitWS()
	p.subscribedTo[code] = true
	ch := make(WSOrderOrTradeChan)
	p.ws.Subscribe(code, p.makeOrderHandler(code, ch))
	return ch
}

//UnsubscribeTicker ... I think you can guess
func (p *Poloniex) UnsubscribeTicker() {
	p.InitWS()
	p.Unsubscribe("ticker")
}

//UnsubscribeOrder ... I think you can guess
func (p *Poloniex) UnsubscribeOrder(code string) {
	p.InitWS()
	p.Unsubscribe(code)
}

//Unsubscribe from the relevant feed
func (p *Poloniex) Unsubscribe(code string) {
	p.InitWS()
	if p.isSubscribed(code) {
		delete(p.subscribedTo, code)
		p.ws.Unsubscribe(code)
	}
}

//makeTickerHandler takes a WS Order or Trade and send it over the channel sepcified by the user
func (p *Poloniex) makeTickerHandler(ch chan WSTicker) turnpike.EventHandler {
	return func(p []interface{}, n map[string]interface{}) {
		var t WSTicker

		t.Pair = p[0].(string)
		if parsed, err := ggm.ParseDecimal(p[1]); err != nil {
			log.Println("ws ticker parse Last error: " + err.Error())
		} else {
			t.Last = parsed
		}

		if parsed, err := ggm.ParseDecimal(p[2]); err != nil {
			log.Println("ws ticker parse Ask error: " + err.Error())
		} else {
			t.Ask = parsed
		}

		if parsed, err := ggm.ParseDecimal(p[3]); err != nil {
			log.Println("ws ticker parse Bid error: " + err.Error())
		} else {
			t.Bid = parsed
		}

		if parsed, err := ggm.ParseDecimal(p[4]); err != nil {
			log.Println("ws ticker parse PercentChange error: " + err.Error())
		} else {
			t.Last = parsed.MultiplyFloat(100)
		}

		if parsed, err := ggm.ParseDecimal(p[5]); err != nil {
			log.Println("ws ticker parse BaseVolume error: " + err.Error())
		} else {
			t.BaseVolume = parsed
		}

		if parsed, err := ggm.ParseDecimal(p[6]); err != nil {
			log.Println("ws ticker parse Quote Volume error: " + err.Error())
		} else {
			t.QuoteVolume = parsed
		}

		if parsed, err := ggm.ParseDecimal(p[7]); err != nil {
			log.Println("ws ticker parse IsFrozen error: " + err.Error())
		} else {
			t.IsFrozen = !parsed.EqualFloat(0.0)
		}

		if parsed, err := ggm.ParseDecimal(p[8]); err != nil {
			log.Println("ws ticker parse Daily High error: " + err.Error())
		} else {
			t.DailyHigh = parsed
		}

		if parsed, err := ggm.ParseDecimal(p[9]); err != nil {
			log.Println("ws ticker parse Daily Low error: " + err.Error())
		} else {
			t.DailyLow = parsed
		}

		//t := WSTicker{
		//	Pair:          p[0].(string),
		//	Last:          f(p[1]),
		//	Ask:           f(p[2]),
		//	Bid:           f(p[3]),
		//	PercentChange: f(p[4]) * 100.0,
		//	BaseVolume:    f(p[5]),
		//	QuoteVolume:   f(p[6]),
		//	IsFrozen:      p[7].(float64) != 0.0,
		//	DailyHigh:     f(p[8]),
		//	DailyLow:      f(p[9]),
		//}
		ch <- t
	}
}

//makeOrderHandler takes a WS Order or Trade and send it over the channel sepcified by the user
func (p *Poloniex) makeOrderHandler(coin string, ch WSOrderOrTradeChan) turnpike.EventHandler {
	return func(p []interface{}, n map[string]interface{}) {
		seq := int64(SENTINEL)
		if s, ok := n["seq"]; ok {
			seq = int64(s.(float64))
		}
		b, err := json.Marshal(p)
		if err != nil {
			log.Println(err)
			return
		}
		oot := WSOrders{}
		err = json.Unmarshal(b, &oot)
		if err != nil {
			log.Println(err)
			return
		}
		ootTmp := WSOrders{}
		for _, o := range oot {
			if o.Type == "newTrade" {
				pp.Println("Date:", o.Data.Date)
				d, err := time.Parse("2006-01-02 15:04:05", o.Data.Date)
				if err != nil {
					log.Println(err)
				}
				o.Data.TS = d
			}
			ootTmp = append(ootTmp, o)
		}
		o := WSOrderOrTrade{Seq: seq, Orders: ootTmp}
		ch <- o
	}
}
