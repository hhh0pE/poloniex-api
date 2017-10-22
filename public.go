package poloniex

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"log"

	"github.com/franela/goreq"
	"github.com/hhh0pE/ggm"
	"github.com/k0kubun/pp"
)

type (
	Ticker      map[string]TickerEntry
	TickerEntry struct {
		Last        ggm.Decimal `json:",string"`
		Ask         ggm.Decimal `json:"lowestAsk,string"`
		Bid         ggm.Decimal `json:"highestBid,string"`
		Change      ggm.Decimal `json:"percentChange,string"`
		BaseVolume  ggm.Decimal `json:"baseVolume,string"`
		QuoteVolume ggm.Decimal `json:"quoteVolume,string"`
		IsFrozen    int64       `json:"isFrozen,string"`
	}

	DailyVolume          map[string]DailyVolumeEntry
	DailyVolumeEntry     map[string]ggm.Decimal
	DailyVolumeTemp      map[string]interface{}
	DailyVolumeEntryTemp map[string]interface{}

	OrderBook struct {
		Asks     []Order
		Bids     []Order
		IsFrozen bool
	}
	Order struct {
		Rate   ggm.Decimal
		Amount ggm.Decimal
	}

	OrderBookTemp struct {
		Asks     []OrderTemp
		Bids     []OrderTemp
		IsFrozen interface{}
	}
	OrderTemp        []interface{}
	OrderBookAll     map[string]OrderBook
	OrderBookAllTemp map[string]OrderBookTemp

	TradeHistory      []TradeHistoryEntry
	TradeHistoryEntry struct {
		ID     int64 `json:"globalTradeID"`
		Date   string
		Type   string
		Rate   ggm.Decimal `json:",string"`
		Amount ggm.Decimal `json:",string"`
		Total  ggm.Decimal `json:",string"`
	}

	ChartData      []ChartDataEntry
	ChartDataEntry struct {
		Date            int64
		High            ggm.Decimal
		Low             ggm.Decimal
		Open            ggm.Decimal
		Close           ggm.Decimal
		Volume          ggm.Decimal
		QuoteVolume     ggm.Decimal
		WeightedAverage ggm.Decimal
	}

	Currencies map[string]Currency
	Currency   struct {
		Name           string
		TxFee          ggm.Decimal `json:",string"`
		MinConf        ggm.Decimal
		DepositAddress string
		Disabled       int64
		Delisted       int64
		Frozen         int64
	}

	LoanOrders struct {
		Offers  []LoanOrder
		Demands []LoanOrder
	}
	LoanOrder struct {
		Rate     ggm.Decimal `json:",string"`
		Amount   ggm.Decimal `json:",string"`
		RangeMin ggm.Decimal
		RangeMax ggm.Decimal
	}
)

func (p *Poloniex) Ticker() (ticker Ticker, err error) {
	err = p.public("returnTicker", nil, &ticker)
	return
}

func (p *Poloniex) DailyVolume() (dailyVolume DailyVolume, err error) {
	dvt := DailyVolumeTemp{}
	err = p.public("return24hVolume", nil, &dvt)
	if err != nil {
		return
	}
	dailyVolume = DailyVolume{}
	for k := range dvt {
		v := dvt[k]
		dve := DailyVolumeEntry{}
		switch i := v.(type) {
		default:
			v := i.(map[string]interface{})
			for kk, vv := range v {
				if parsed, err := ggm.ParseDecimal(vv); err != nil {
					log.Println("Error when parsing \"", vv, "\": "+err.Error())
					dve[kk] = parsed
				} else {
					dve[kk] = parsed
				}
			}
			dailyVolume[k] = dve
		case string:
			//ignore anything that isn't a map
		}
	}
	return
}

func (p *Poloniex) OrderBook(pair string) (orderBook OrderBook, err error) {
	params := url.Values{}
	params.Add("currencyPair", pair)
	params.Add("depth", "40")
	obt := OrderBookTemp{}
	err = p.public("returnOrderBook", params, &obt)
	if err != nil {
		return
	}
	orderBook = tempToOrderBook(obt)
	return
}

func (p *Poloniex) OrderBookAll() (orderBook OrderBookAll, err error) {
	params := url.Values{}
	params.Add("depth", "5")
	params.Add("currencyPair", "all")
	obt := OrderBookAllTemp{}
	err = p.public("returnOrderBook", params, &obt)
	if err != nil {
		return
	}
	orderBook = OrderBookAll{}
	for k, v := range obt {
		orderBook[k] = tempToOrderBook(v)
	}
	return
}

func (p *Poloniex) TradeHistory(in ...interface{}) (tradeHistory TradeHistory, err error) {
	pp.Println(in)
	params := url.Values{}
	params.Add("currencyPair", in[0].(string))
	if len(in) > 1 {
		// we have a start date
		params.Add("start", fmt.Sprintf("%d", in[1].(int64)))
	}
	if len(in) > 2 {
		// we have an end date
		params.Add("end", fmt.Sprintf("%d", in[2].(int64)))
	}
	err = p.public("returnTradeHistory", params, &tradeHistory)
	return
}

func (p *Poloniex) ChartData(pair string) (chartData ChartData, err error) {
	params := url.Values{}
	params.Add("currencyPair", pair)
	params.Add("start", fmt.Sprintf("%d", time.Now().Add(-24*time.Hour).Unix()))
	params.Add("end", "9999999999")
	params.Add("period", "300")
	err = p.public("returnChartData", params, &chartData)
	return
}

func (p *Poloniex) ChartDataPeriod(pair string, start, end time.Time) (chartData ChartData, err error) {
	params := url.Values{}
	params.Add("currencyPair", pair)
	params.Add("start", fmt.Sprintf("%d", start.Unix()))
	params.Add("end", fmt.Sprintf("%d", end.Unix()))
	params.Add("period", "300")
	err = p.public("returnChartData", params, &chartData)
	return
}

func (p *Poloniex) ChartDataCurrent(pair string) (chartData ChartData, err error) {
	params := url.Values{}
	params.Add("currencyPair", pair)
	params.Add("start", fmt.Sprintf("%d", time.Now().Add(-5*time.Minute).Unix()))
	params.Add("end", "9999999999")
	params.Add("period", "300")
	err = p.public("returnChartData", params, &chartData)
	return
}

func (p *Poloniex) Currencies() (currencies Currencies, err error) {
	err = p.public("returnCurrencies", nil, &currencies)
	return
}

func (p *Poloniex) LoanOrders(currency string) (loanOrders LoanOrders, err error) {
	params := url.Values{}
	params.Add("currency", currency)
	err = p.public("returnLoanOrders", params, &loanOrders)
	return
}

func tempToOrderBook(obt OrderBookTemp) (ob OrderBook) {
	asks := obt.Asks
	bids := obt.Bids
	ob.IsFrozen = obt.IsFrozen.(string) != "0"
	ob.Asks = []Order{}
	ob.Bids = []Order{}
	for k := range asks {
		v := asks[k]
		var o Order
		if parsed, err := ggm.ParseDecimal(v[0]); err != nil {
			log.Println("tempToOrderBook error when parsing ask's Rate: " + err.Error())
		} else {
			o.Rate = parsed
		}

		if parsed, err := ggm.ParseDecimal(v[1]); err != nil {
			log.Println("tempToOrderBook error when parsing ask's Amount: " + err.Error())
		} else {
			o.Amount = parsed
		}
		ob.Asks = append(ob.Asks, o)
	}
	for k := range bids {
		v := bids[k]
		var o Order

		if parsed, err := ggm.ParseDecimal(v[0]); err != nil {
			log.Println("tempToOrderBook error when parsing bid's Rate: " + err.Error())
		} else {
			o.Rate = parsed
		}

		if parsed, err := ggm.ParseDecimal(v[1]); err != nil {
			log.Println("tempToOrderBook error when parsing bid's Amount: " + err.Error())
		} else {
			o.Amount = parsed
		}
		ob.Bids = append(ob.Bids, o)
	}
	return
}

//func floatCmp(a, b interface{}) int {
//	fa := f(a)
//	fb := f(b)
//	if fa < fb {
//		return -1
//	} else if fa > fb {
//		return 1
//	}
//	return 0
//}
//
//func reverseFloatCmp(a, b interface{}) int {
//	return floatCmp(a, b) * -1
//}

func (p *Poloniex) public(command string, params url.Values, retval interface{}) (err error) {
	if p.debug {
		defer un(trace("public: " + command))
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if params == nil {
		params = url.Values{}
	}
	params.Add("command", command)
	req := goreq.Request{Uri: PUBLICURI, QueryString: params, Timeout: 130 * time.Second}
	res, err := req.Do()
	if err != nil {
		return
	}
	if p.debug {
		pp.Println(res.Request.URL.String())
	}

	defer res.Body.Close()

	s, err := res.Body.ToString()
	if err != nil {
		return
	}
	if p.debug {
		pp.Println(s)
	}
	err = json.Unmarshal([]byte(s), retval)
	return
}
