# yauga 🧘‍♂️
Yet another Upbit API for golang

# Precautions
1. WIP
2. I am not responsible for anything done with this.
3. YOU USE IT AT YOUR OWN RISK.
4. But, contribution welcomed!

# Dependencies
* [google/uuid](https://github.com/google/uuid)
* [golang-jwt/jwt](https://github.com/golang-jwt/jwt)

Check out `dependencies.sh`.

# Test
`go test ./` or `go test ./ -v`

# Progress status
## Exchange API
* [x] GET @ accounts
* [ ] GET @ orders/chance
* [ ] GET @ order
* [ ] GET @ orders
* [ ] DELETE @ order
* [ ] POST @ orders
* [ ] GET @ withdraws
* [ ] GET @ withdraw
* [ ] GET @ withdraws/chance
* [ ] POST @ withdraws/coin
* [ ] POST @ withdraws/krw
* [ ] GET @ deposits
* [ ] GET @ deposit
* [ ] POST @ deposits/generate_coin_address
* [ ] GET @ deposits/coin_addresses
* [ ] GET @ deposits/coin_address
* [ ] POST @ deposits/krw
* [ ] GET @ status/wallet
* [ ] GET @ api_keys
### Quotation API
* [x] GET @ market/all
* [x] GET @ candles/minutes/{unit}
* [x] GET @ candles/days
* [x] GET @ candles/weeks
* [ ] GET @ candles/months
* [ ] GET @ trades/ticks
* [ ] GET @ ticker
* [ ] GET @ orderbook

# Example
## 전체계좌 조회
* [Upbit API document @ /v1/accounts](https://docs.upbit.com/reference/%EC%A0%84%EC%B2%B4-%EA%B3%84%EC%A2%8C-%EC%A1%B0%ED%9A%8C)
```.go
upbit := NewUpbit(accessKey)
upbit.SetSecretKey(secretKey)
raw := upbit.Accounts()
fmt.Print(raw.Response[0].Currency) // Result: KRW (통화코드)
fmt.Print(raw.Response[0].Balance) // Result: <Numberic> (잔액)
```
