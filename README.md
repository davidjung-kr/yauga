# yauga 🧘‍♂️
Yet another Upbit API for golang

# Precautions
1. I am not responsible for anything done with this.
2. YOU USE IT AT YOUR OWN RISK.
3. But, contribution welcomed!

# Dependencies
* [google/uuid](https://github.com/google/uuid)
* [golang-jwt/jwt](https://github.com/golang-jwt/jwt)

Check out `dependencies.sh`.

# Test
`go test ./` or `go test ./ -v`
# Example
## 전체계좌 조회
* [Upbit API document @ /v1/accounts](https://docs.upbit.com/reference/%EC%A0%84%EC%B2%B4-%EA%B3%84%EC%A2%8C-%EC%A1%B0%ED%9A%8C)
```.go
upbit := NewUpbit(accessKey, uuid.NewString())
upbit.Payload(secretKey)
raw := upbit.Accounts()
fmt.Print(raw.Response[0].Currency) // Result: KRW (통화코드)
fmt.Print(raw.Response[0].Balance) // Result: <Numberic> (잔액)
```
