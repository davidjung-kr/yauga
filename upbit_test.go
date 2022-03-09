package main

/**
 * yauga_test -  Yet another Upbit API for golang / LGPL-v2.1
 * 2022, David Jung @ github.com/davidjung-kr/yauga
 *
 * I am not responsible for anything done with this. YOU USE IT AT YOUR OWN RISK.
 */
import (
	"os"
	"testing"
)

// 환경변수 상에서 엑세스 데이터 취득
func getEnvData() (string, string) {
	return os.Getenv("YAUGA_ACCESS_KEY"), os.Getenv("YAUGA_SECRECT_KEY")
}

// 생성자 테스트
func TestUpbitNew(t *testing.T) {
	accessKey, secretKey := getEnvData()

	if accessKey == "" {
		t.Errorf("Please set a `YAUGA_ACCESS_KEY`.")
	}
	if len(secretKey) == 0 {
		t.Errorf("Please set a `YAUGA_SECRECT_KEY`.")
	}

	upbit := NewUpbit(accessKey)

	token, payloadErr := upbit.Payload(secretKey, "")
	if token == "" || payloadErr != nil {
		t.Errorf("TestUpbitNew | token:[%s], payloadErr:[%s]", token, payloadErr)
	}
}

// Account 테스트
func TestUpbitAccount(t *testing.T) {
	accessKey, secretKey := getEnvData()
	upbit := NewUpbit(accessKey)
	upbit.Payload(secretKey, "")
	x := upbit.Accounts()
	if x.Common.StatusCode != 200 || x.Common.Error != nil || len(x.Response) <= 0 {
		t.Errorf("TestUpbitAccount | Status:[%d], accountsErr:[%s]", x.Common.StatusCode, x.Common.Error)
	}
}

// MarketAll 테스트
func TestUpbitMarketAll(t *testing.T) {
	accessKey, _ := getEnvData()
	upbit := NewUpbit(accessKey)
	x := upbit.MarketAll(true)
	if x.Common.StatusCode != 200 || x.Common.Error != nil {
		t.Errorf("TestUpbitMarketAll | Status:[%d], MarketAllErr:[%s]", x.Common.StatusCode, x.Common.Error)
	}
}

// CandlesMinutes 테스트
func TestUpbitCandlesMinutes(t *testing.T) {
	accessKey, _ := getEnvData()
	upbit := NewUpbit(accessKey)
	x := upbit.CandlesMinutes(1, "KRW-BTC", "", 1)
	if x.Common.StatusCode != 200 || x.Common.Error != nil {
		t.Errorf("TestUpbitCandlesMinutes | Status:[%d], candlesMinutesErr:[%s]", x.Common.StatusCode, x.Common.Error)
	}
}

// CandlesDays 테스트
func TestUpbitCandlesDays(t *testing.T) {
	accessKey, _ := getEnvData()
	upbit := NewUpbit(accessKey)
	x := upbit.CandlesDays("KRW-BTC", "", 1, "KRW")
	if x.Common.StatusCode != 200 || x.Common.Error != nil {
		t.Errorf("TestUpbitCandlesDays | Status:[%d], candlesDaysErr:[%s]", x.Common.StatusCode, x.Common.Error)
	}
}
