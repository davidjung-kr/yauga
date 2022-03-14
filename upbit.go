package main

/**
 * yauga -  Yet another Upbit API for golang / LGPL-v2.1
 * 2022, David Jung @ github.com/davidjung-kr/yauga
 *
 * I am not responsible for anything done with this. YOU USE IT AT YOUR OWN RISK.
 */
import (
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

const (
	// [Exchange API] 전체 계좌 조회 (Full account inquiry)
	UPBIT_URL_ACCOUNTS = "https://api.upbit.com/v1/accounts"
	// [Exchange API] 주문 가능 정보
	UPBIT_URL_ORDERS_CHANCE = "https://api.upbit.com/v1/orders/chance?market=%s-%s"

	// [Quotation API] 마켓 코드 조회 (Market code inquiry)
	UPBIT_URL_MARKET_ALL = "https://api.upbit.com/v1/market/all"
	// [Quotation API] 분(Minute) 캔들 (Minutes candles inquiry)
	UPBIT_URL_CANDLES_MINUTES = "https://api.upbit.com/v1/candles/minutes/%d"
	// [Quotation API] 일(Day) 캔들 (Days candles inquiry)
	UPBIT_URL_CANDLES_DAYS = "https://api.upbit.com/v1/candles/days"
	// [Quotation API] 주(Week) 캔들 (Weeks candles inquiry)
	UPBIT_URL_CANDLES_WEEKS = "https://api.upbit.com/v1/candles/weeks"
)

type Upbit struct {
	AccessKey        string
	token, secretKey string
}

// Initialization
func NewUpbit(accessKey string) *Upbit {
	return &Upbit{AccessKey: accessKey}
}

/*type NewUpbitRequest struct {
	// 발급 받은 acccess key (필수)
	AccessKey string `json:"access_key"`
	// 무작위의 UUID 문자열 (필수)
	Nonce string `json:"nonce"`
	// 해싱된 query string (파라미터가 있을 경우 필수)
	QueryHash string `json:"query_hash"`
	// query_hash를 생성하는 데에 사용한 알고리즘 (기본값 : SHA512)
	QueryHashAlg string `json:"query_hash_alg"`
}*/

// 시크릿키 세팅
func (o *Upbit) SetSecretKey(secretKey string) {
	o.secretKey = secretKey
}

// 토큰 취득
func (o *Upbit) GetToken() string {
	return o.token
}

type PayloadOption struct {
	WithParams bool
	Url        string
}

// 인증 가능한 요청 만들기
//  서명 방식은 HS256 을 권장하며, 서명에 사용할 secret은 발급받은 secret key를 사용합니다.
//  페이로드의 구성은 다음과 같습니다.
func (o *Upbit) payload(opt PayloadOption) (string, error) {
	claim := jwt.MapClaims{}
	claim["access_key"] = o.AccessKey
	claim["nonce"] = uuid.New()

	if opt.WithParams {
		claim["query_hash"] = fmt.Sprintf("%x", sha512.Sum512([]byte(url.QueryEscape(opt.Url))))
		claim["query_hash_alg"] = "SHA512"
	}

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	token, err := at.SignedString([]byte(o.secretKey))

	if err != nil {
		return "", err
	}
	o.token = "Bearer " + token
	return o.token, nil
}

// [Exchange API] 전체 계좌 조회 @ accounts
//  내가 보유한 자산 리스트를 보여줍니다.
func (o *Upbit) Accounts() UpbitAccounts {
	o.payload(PayloadOption{WithParams: false})
	req, _ := http.NewRequest("GET", UPBIT_URL_ACCOUNTS, nil)
	if o.token == "" {
		panic("Please do `Payload` first.")
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", o.token)

	var res UpbitAccounts

	httpRes, httpErr := http.DefaultClient.Do(req)
	if httpErr != nil {
		res.Common.Error = httpErr
		return res
	}
	body, ioErr := ioutil.ReadAll(httpRes.Body)
	defer httpRes.Body.Close()
	if ioErr != nil {
		res.Common.Error = ioErr
		return res
	}
	content := string(body[:])
	if httpRes.StatusCode != 200 {
		var errorBlock UpbitErrorResponse
		json.Unmarshal([]byte(content), &errorBlock)

		res.Common.StatusCode = httpRes.StatusCode
		res.Common.Error = errors.New(errorBlock.ErrorBlock.Name + " (" + errorBlock.ErrorBlock.Message + ")")
		return res
	}
	var blocks []UpbitAccountBlock
	json.Unmarshal([]byte(content), &blocks)
	res.Common.StatusCode = httpRes.StatusCode
	res.Response = blocks
	return res
}

// [Exchange API] 주문 가능 정보 @ orders/chance
//  마켓별 주문 가능 정보를 확인한다.
// Params:
//	bidCurrencyTicker = 매수 시 사용할 통화
//	AskCurrencyTicker = 매도 시 사용할 통화
func (o *Upbit) OrdersChance(bidCurrencyTicker string, AskCurrencyTicker string) UpbitOrdersChance {
	url := fmt.Sprintf(UPBIT_URL_ORDERS_CHANCE, bidCurrencyTicker, AskCurrencyTicker)
	o.payload(PayloadOption{WithParams: true, Url: url})
	req, _ := http.NewRequest("GET", url, nil)
	if o.token == "" {
		panic("Please do `Payload` first.")
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", o.token)

	var res UpbitOrdersChance

	httpRes, httpErr := http.DefaultClient.Do(req)
	if httpErr != nil {
		res.Common.Error = httpErr
		return res
	}
	body, ioErr := ioutil.ReadAll(httpRes.Body)
	defer httpRes.Body.Close()
	if ioErr != nil {
		res.Common.Error = ioErr
		return res
	}
	content := string(body[:])
	if httpRes.StatusCode != 200 {
		var errorBlock UpbitErrorResponse
		json.Unmarshal([]byte(content), &errorBlock)

		res.Common.StatusCode = httpRes.StatusCode
		res.Common.Error = errors.New(errorBlock.ErrorBlock.Name + " (" + errorBlock.ErrorBlock.Message + ")")
		return res
	}
	var block UpbitOrdersChanceBlock
	json.Unmarshal([]byte(content), &block)
	res.Common.StatusCode = httpRes.StatusCode
	res.Response = block
	return res
}

// [Quotation API] 마켓 코드 조회 @ market/all
//  업비트에서 거래 가능한 마켓 목록
// Params:
// 	isDetails = 유의종목 필드과 같은 상세 정보 노출 여부
func (o *Upbit) MarketAll(isDetails bool) UpbitMarketAll {
	params := url.Values{}
	params.Add("isDetails", strconv.FormatBool(isDetails))
	var encodedUrl string
	if len(params) > 0 {
		encodedUrl = UPBIT_URL_MARKET_ALL + "?" + params.Encode()
	} else {
		encodedUrl = UPBIT_URL_MARKET_ALL
	}

	req, _ := http.NewRequest("GET", encodedUrl, nil)
	req.Header.Add("Accept", "application/json")

	var res UpbitMarketAll

	httpRes, httpErr := http.DefaultClient.Do(req)
	if httpErr != nil {
		res.Common.Error = httpErr
		return res
	}
	body, ioErr := ioutil.ReadAll(httpRes.Body)
	defer httpRes.Body.Close()
	if ioErr != nil {
		res.Common.Error = ioErr
		return res
	}
	content := string(body[:])
	if httpRes.StatusCode != 200 {
		var errorBlock UpbitErrorResponse
		json.Unmarshal([]byte(content), &errorBlock)

		res.Common.StatusCode = httpRes.StatusCode
		res.Common.Error = errors.New(errorBlock.ErrorBlock.Name + " (" + errorBlock.ErrorBlock.Message + ")")
		return res
	}
	res.Common.StatusCode = httpRes.StatusCode

	var blocks []UpbitMarketAllBlock
	json.Unmarshal([]byte(content), &blocks)

	if len(blocks) <= 0 {
		res.Common.Error = errors.New("HTTP STATUS IS 200 BUT RESULT IS EMPTY")
		return res
	}

	res.Response = blocks
	return res
}

// [Quotation API] 분(Minute) 캔들 @ candles/minutes/
// Params:
// 	unit = 분 단위. 가능한 값 : 1, 3, 5, 15, 10, 30, 60, 240
//	market = 마켓 코드 (ex. KRW-BTC)
//	to = 마지막 캔들 시각 (exclusive). 포맷 : yyyy-MM-dd'T'HH:mm:ss'Z' or yyyy-MM-dd HH:mm:ss. 비워서 요청시 가장 최근 캔들
//	count = 캔들 개수(최대 200개까지 요청 가능)
func (o *Upbit) CandlesMinutes(unit int, market string, to string, count int) UpbitCandlesMinutes {
	var targetUrl string
	params := url.Values{}
	switch unit {
	case 1:
		fallthrough
	case 3:
		fallthrough
	case 5:
		fallthrough
	case 15:
		fallthrough
	case 10:
		fallthrough
	case 30:
		fallthrough
	case 60:
		fallthrough
	case 240:
		targetUrl = fmt.Sprintf(UPBIT_URL_CANDLES_MINUTES, unit)
	default:
		panic("unit was wrong!")
	}
	if market != "" {
		params.Add("market", market)
	}
	if to != "" {
		params.Add("to", to)
	}
	if count > 0 {
		if count > 200 {
			panic("Count field only accept until 200!")
		}
		params.Add("count", strconv.Itoa(count))
	}

	encodedUrl := targetUrl + "?" + params.Encode()
	req, _ := http.NewRequest("GET", encodedUrl, nil)
	req.Header.Add("Accept", "application/json")

	var res UpbitCandlesMinutes

	httpRes, httpErr := http.DefaultClient.Do(req)
	if httpErr != nil {
		res.Common.Error = httpErr
		return res
	}
	body, ioErr := ioutil.ReadAll(httpRes.Body)
	defer httpRes.Body.Close()
	if ioErr != nil {
		res.Common.Error = ioErr
		return res
	}
	content := string(body[:])
	if httpRes.StatusCode != 200 {
		var errorBlock UpbitErrorResponse
		json.Unmarshal([]byte(content), &errorBlock)

		res.Common.StatusCode = httpRes.StatusCode
		res.Common.Error = errors.New(errorBlock.ErrorBlock.Name + " (" + errorBlock.ErrorBlock.Message + ")")
		return res
	}
	res.Common.StatusCode = httpRes.StatusCode

	var blocks []UpbitCandlesMinutesBlock
	json.Unmarshal([]byte(content), &blocks)

	if len(blocks) <= 0 {
		res.Common.Error = errors.New("HTTP STATUS IS 200 BUT RESULT IS EMPTY")
		return res
	}

	res.Response = blocks
	return res
}

// [Quotation API] 일(Day) 캔들 @ candles/days
//  `convertingPriceUnit` 파라미터의 경우, 원화 마켓이 아닌 다른 마켓(ex. BTC, ETH)의 일봉 요청시
//	종가를 명시된 파라미터 값으로 환산해 `converted_trade_price` 필드에 추가하여 반환합니다.
//	현재는 원화(KRW) 로 변환하는 기능만 제공하며 추후 기능을 확장할 수 있습니다.
// Params:
// 	market = 마켓 코드 (ex. KRW-BTC)
// 	to = 마지막 캔들 시각 (exclusive). 포맷 : yyyy-MM-dd'T'HH:mm:ss'Z' or yyyy-MM-dd HH:mm:ss. 비워서 요청시 가장 최근 캔들
//	count = 캔들 개수
//	convertingPriceUnit = 종가 환산 화폐 단위 (생략 가능, KRW로 명시할 시 원화 환산 가격을 반환.)
func (o *Upbit) CandlesDays(market string, to string, count int, convertingPriceUnit string) UpbitCandlesDays {
	params := url.Values{}
	if market != "" {
		params.Add("market", market)
	}
	if to != "" {
		params.Add("to", to)
	}
	if count > 0 {
		params.Add("count", strconv.Itoa(count))
	}
	if convertingPriceUnit != "" {
		params.Add("convertingPriceUnit", convertingPriceUnit)
	}

	var encodedUrl string
	if len(params) > 0 {
		encodedUrl = UPBIT_URL_CANDLES_DAYS + "?" + params.Encode()
	} else {
		encodedUrl = UPBIT_URL_CANDLES_DAYS
	}

	req, _ := http.NewRequest("GET", encodedUrl, nil)
	req.Header.Add("Accept", "application/json")

	var res UpbitCandlesDays

	httpRes, httpErr := http.DefaultClient.Do(req)
	if httpErr != nil {
		res.Common.Error = httpErr
		return res
	}
	body, ioErr := ioutil.ReadAll(httpRes.Body)
	defer httpRes.Body.Close()
	if ioErr != nil {
		res.Common.Error = ioErr
		return res
	}
	content := string(body[:])
	if httpRes.StatusCode != 200 {
		var errorBlock UpbitErrorResponse
		json.Unmarshal([]byte(content), &errorBlock)

		res.Common.StatusCode = httpRes.StatusCode
		res.Common.Error = errors.New(errorBlock.ErrorBlock.Name + " (" + errorBlock.ErrorBlock.Message + ")")
		return res
	}
	res.Common.StatusCode = httpRes.StatusCode

	var blocks []UpbitCandlesDaysBlock
	json.Unmarshal([]byte(content), &blocks)

	if len(blocks) <= 0 {
		res.Common.Error = errors.New("HTTP STATUS IS 200 BUT RESULT IS EMPTY")
		return res
	}

	res.Response = blocks[0]
	return res
}

// [Quotation API] 주(Week) 캔들 @ candles/weeks
// Params:
// 	market = 마켓 코드 (ex. KRW-BTC)
// 	to = 마지막 캔들 시각 (exclusive). 포맷 : yyyy-MM-dd'T'HH:mm:ss'Z' or yyyy-MM-dd HH:mm:ss. 비워서 요청시 가장 최근 캔들
//	count = 캔들 개수
func (o *Upbit) CandlesWeeks(market string, to string, count int, convertingPriceUnit string) UpbitCandlesWeeks {
	params := url.Values{}
	if market != "" {
		params.Add("market", market)
	}
	if to != "" {
		params.Add("to", to)
	}
	if count > 0 {
		params.Add("count", strconv.Itoa(count))
	}

	var encodedUrl string
	if len(params) > 0 {
		encodedUrl = UPBIT_URL_CANDLES_WEEKS + "?" + params.Encode()
	} else {
		encodedUrl = UPBIT_URL_CANDLES_WEEKS
	}

	req, _ := http.NewRequest("GET", encodedUrl, nil)
	req.Header.Add("Accept", "application/json")

	var res UpbitCandlesWeeks

	httpRes, httpErr := http.DefaultClient.Do(req)
	if httpErr != nil {
		res.Common.Error = httpErr
		return res
	}
	body, ioErr := ioutil.ReadAll(httpRes.Body)
	defer httpRes.Body.Close()
	if ioErr != nil {
		res.Common.Error = ioErr
		return res
	}
	content := string(body[:])
	if httpRes.StatusCode != 200 {
		var errorBlock UpbitErrorResponse
		json.Unmarshal([]byte(content), &errorBlock)

		res.Common.StatusCode = httpRes.StatusCode
		res.Common.Error = errors.New(errorBlock.ErrorBlock.Name + " (" + errorBlock.ErrorBlock.Message + ")")
		return res
	}
	res.Common.StatusCode = httpRes.StatusCode

	var blocks []UpbitCandlesWeeksBlock
	json.Unmarshal([]byte(content), &blocks)

	if len(blocks) <= 0 {
		res.Common.Error = errors.New("HTTP STATUS IS 200 BUT RESULT IS EMPTY")
		return res
	}

	res.Response = blocks
	return res
}

// Error Response
type UpbitErrorResponse struct {
	ErrorBlock UpbitErrorBlock `json:"error"`
}

// Error block
type UpbitErrorBlock struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

// Common part
type UpbitCommonBlock struct {
	// HTTP Status code
	StatusCode int
	// Error
	Error error
}

// 전체 계좌 조회 @ accounts 결과
type UpbitAccounts struct {
	Response []UpbitAccountBlock
	Common   UpbitCommonBlock
}

// 주문 가능 정보 @ orders/chance 결과
type UpbitOrdersChance struct {
	Response UpbitOrdersChanceBlock
	Common   UpbitCommonBlock
}

// 마켓 코드 조회 @ market/all
type UpbitMarketAll struct {
	Response []UpbitMarketAllBlock
	Common   UpbitCommonBlock
}

// 분(Minute) 캔들 @ candles/minutes 결과
type UpbitCandlesMinutes struct {
	Response []UpbitCandlesMinutesBlock
	Common   UpbitCommonBlock
}

// 일(Day) 캔들 @ candles/days 결과
type UpbitCandlesDays struct {
	Response UpbitCandlesDaysBlock
	Common   UpbitCommonBlock
}

// 주(Week) 캔들 @ candles/weeks 결과
type UpbitCandlesWeeks struct {
	Response []UpbitCandlesWeeksBlock
	Common   UpbitCommonBlock
}

// 전체 계좌 조회 @ accounts Block
type UpbitAccountBlock struct {
	// 화폐를 의미하는 영문 대문자 코드 [Stirng]
	Currency string `json:"currency"`
	// 주문가능 금액/수량 [NumberString]
	Balance int64 `json:"balance"`
	// 주문 중 묶여있는 금액/수량 [NumberString]
	Locked int64 `json:"locked"`
	// 매수평균가 [NumberString]
	AvgBuyPrice int64 `json:"avg_buy_price"`
	// 매수평균가 수정 여부	[Boolean]
	AvgBuyPriceModified bool `json:"avg_buy_price_modified"`
	// 평단가 기준 화폐	[String]
	UnitCurreny string `json:"unit_currency"`
}

// 주문 가능 정보 @ orders/chance Block
type UpbitOrdersChanceBlock struct {
	// 매도 수수료 비율 [NumberString]
	AskFee string `json:"ask_fee"`
	// 마켓에 대한 정보 [Object]
	Market MarketBlock `json:"market"`
	// 매수 시 사용하는 화폐의 계좌 상태 [Object]
	BidAccount BidAskAccountBlock `json:"bid_account"`
	// 매도 시 사용하는 화폐의 계좌 상태 [Object]
	AskAccount BidAskAccountBlock `json:"ask_account"`
}

type MarketBlock struct {
	// 마켓의 유일 키 [String]
	Id string `json:"id"`
	// 마켓 이름 [String]
	Name string `json:"name"`
	// 지원 주문 방식 [Array[String]]
	OrderTypes []string `json:"order_types"`
	// 지원 주문 종류	[Array[String]]
	OrderSides []string `json:"order_sides"`
	// 매수 시 제약 사항 [Object]
	Bid BidAskBlock `json:"bid"`
	// 매도 시 제약 사항 [Object]
	Ask BidAskBlock `json:"ask"`
	// 최대 매도/매수 금액 [NumberString]
	MaxTotal string `json:"max_total"`
	// 마켓 운영 상태 [String]
	State string `json:"state"`
}

// 매수/매도 시 제약사항
type BidAskBlock struct {
	// 화폐를 의미하는 영문 대문자 코드	[String]
	Currency string `json:"currency"`
	// 주문금액 단위 [String]
	PriceUnit string `json:"price_unit"`
	// 최소 매도/매수 금액 [Number]
	MinTotal int32 `json:"min_total"`
}

// 매수/매도 시 사용하는 화폐의 계좌 상태
type BidAskAccountBlock struct {
	// 화폐를 의미하는 영문 대문자 코드	[String]
	Currency string `json:"currency"`
	// 주문가능 금액/수량 [NumberString]
	Balance string `json:"balance"`
	// 주문 중 묶여있는 금액/수량 [NumberString]
	Locked string `json:"locked"`
	// 매수평균가 [NumberString]
	AvgBuyPrice string `json:"avg_buy_price"`
	// 매수평균가 수정 여부 [Boolean]
	AvgBuyPriceModified bool `json:"avg_buy_price_modified"`
	// 평단가 기준 화폐 [String]
	UnitCurrency string `json:"unit_currency"`
}

// 마켓 코드 조회 @ market/all Block
type UpbitMarketAllBlock struct {
	// 업비트에서 제공중인 시장 정보 [String]
	Market string `json:"market"`
	// 거래 대상 암호화폐 한글명 [String]
	KoreanName string `json:"korean_name"`
	// 거래 대상 암호화폐 영문명 [String]
	EnglishName string `json:"english_name"`
	// 	유의 종목 여부 - NONE (해당 사항 없음), CAUTION(투자유의) [String]
	MarketWarning string `json:"market_warning"`
}

// 분(Minute) 캔들 @ candles/minutes Block
type UpbitCandlesMinutesBlock struct {
	// 마켓명 [String]
	Market string `json:"market"`
	// 캔들 기준 시각(UTC 기준) [String]
	CandleDateTimeUtc string `json:"candle_ddate_time_utc"`
	// 캔들 기준 시각(KST 기준)	[String]
	CandleDateTimeKst string `json:"candle_date_time_kst"`
	// 시가	[Double]
	OpeningPrice float64 `json:"opening_price"`
	// 고가	[Double]
	HighPrice float64 `json:"high_price"`
	// 저가	[Double]
	LowPrice float64 `json:"low_price"`
	// 종가	[Double]
	TradePrice float64 `json:"trade_price"`
	// 해당 캔들에서 마지막 틱이 저장된 시각 [Long]
	Timestamp int64 `json:"timestamp"`
	// 누적 거래 금액 [Double]
	CandleAccTradePrice float64 `json:"candle_acc_trade_price"`
	// 누적 거래량	[Double]
	CandleAccTradeVolume float64 `json:"candle_acc_trade_volume"`
	// 분 단위(유닛) [Integer]
	Unit int32 `json:"unit"`
}

// 일(Day) 캔들 @ candles/days Block
type UpbitCandlesDaysBlock struct {
	// 마켓명 [String]
	Market string `json:"market"`
	// 캔들 기준 시각(UTC 기준) [String]
	CandleDateTimeUtc string `json:"candle_ddate_time_utc"`
	// 캔들 기준 시각(KST 기준)	[String]
	CandleDateTimeKst string `json:"candle_date_time_kst"`
	// 시가	[Double]
	OpeningPrice float64 `json:"opening_price"`
	// 고가	[Double]
	HighPrice float64 `json:"high_price"`
	// 저가	[Double]
	LowPrice float64 `json:"low_price"`
	// 종가	[Double]
	TradePrice float64 `json:"trade_price"`
	// 마지막 틱이 저장된 시각 [Long]
	Timestamp int64 `json:"timestamp"`
	// 누적 거래 금액 [Double]
	CandleAccTradePrice float64 `json:"candle_acc_trade_price"`
	// 누적 거래량	[Double]
	CandleAccTradeVolume float64 `json:"candle_acc_trade_volume"`
	// 전일 종가(UTC 0시 기준)	[Double]
	PrevClosingPrice float64 `json:"prev_closing_price"`
	// 전일 종가 대비 변화 금액	[Double]
	Change_price float64 `json:"change_price"`
	// 전일 종가 대비 변화량	[Double]
	ChangeRate float64 `json:"change_rate"`
	// 종가 환산 화폐 단위로 환산된 가격(요청에 convertingPriceUnit 파라미터 없을 시 해당 필드 포함되지 않음.)	[Double]
	ConvertedTradePrice float64 `json:"converted_trade_price"`
}

// 주(Week) 캔들 @ candles/weeks Block
type UpbitCandlesWeeksBlock struct {
	// 마켓명 [String]
	Market string `json:"market"`
	// 캔들 기준 시각(UTC 기준) [String]
	CandleDateTimeUtc string `json:"candle_ddate_time_utc"`
	// 캔들 기준 시각(KST 기준)	[String]
	CandleDateTimeKst string `json:"candle_date_time_kst"`
	// 시가	[Double]
	OpeningPrice float64 `json:"opening_price"`
	// 고가	[Double]
	HighPrice float64 `json:"high_price"`
	// 저가	[Double]
	LowPrice float64 `json:"low_price"`
	// 종가	[Double]
	TradePrice float64 `json:"trade_price"`
	// 마지막 틱이 저장된 시각 [Long]
	Timestamp int64 `json:"timestamp"`
	// 누적 거래 금액 [Double]
	CandleAccTradePrice float64 `json:"candle_acc_trade_price"`
	// 누적 거래량	[Double]
	CandleAccTradeVolume float64 `json:"candle_acc_trade_volume"`
	// 캔들 기간의 가장 첫 날	[String]
	FirstDayOfPeriod string `json:"first_day_of_period"`
}
