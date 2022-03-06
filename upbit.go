package main

/**
 * yauga -  Yet another Upbit API for golang / LGPL-v2.1
 * 2022, David Jung @ github.com/davidjung-kr/yauga
 *
 * I am not responsible for anything done with this. YOU USE IT AT YOUR OWN RISK.
 */
import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

const (
	// [Exchange API] 전체 계좌 조회(Full account inquiry)
	UPBIT_URL_ACCOUNTS = "https://api.upbit.com/v1/accounts"
	// [Quotation API] 일(Day) 캔들(Days candles inquiry)
	UPBIT_URL_CANDLES_DAYS = "https://api.upbit.com/v1/candles/days"
)

type Upbit struct {
	AccessKey, Nonce, Token string
}

// Initialization
func NewUpbit(AccessKey string, Nonce string) *Upbit {
	return &Upbit{AccessKey: AccessKey, Nonce: Nonce}
}

// 인증 가능한 요청 만들기
//  서명 방식은 HS256 을 권장하며, 서명에 사용할 secret은 발급받은 secret key를 사용합니다.
//  페이로드의 구성은 다음과 같습니다.
// Params:
//	secertKey = That issued by the Upbit developer center
func (o *Upbit) Payload(secertKey string) (string, error) {
	claim := jwt.MapClaims{}
	claim["access_key"] = o.AccessKey
	claim["nonce"] = uuid.New()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	token, err := at.SignedString([]byte(secertKey))
	if err != nil {
		return "", err
	}
	o.Token = "Bearer " + token
	return o.Token, nil
}

// [Exchange API] 전체 계좌 조회 @ accounts
//  내가 보유한 자산 리스트를 보여줍니다.
func (o *Upbit) Accounts() UpbitAccounts {
	req, _ := http.NewRequest("GET", UPBIT_URL_ACCOUNTS, nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", o.Token)

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

// 일(Day) 캔들 @ candles/days 결과
type UpbitCandlesDays struct {
	Response UpbitCandlesDaysBlock
	Common   UpbitCommonBlock
}

// 전체 계좌 조회 @ accounts Block
type UpbitAccountBlock struct {
	// 화폐를 의미하는 영문 대문자 코드 (string)
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