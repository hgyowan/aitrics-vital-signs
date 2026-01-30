# router 작성시 필수 참고 규칙

## Router의 역할과 책임
- Router는 URL 경로와 Controller 메서드를 연결하는 역할만 수행합니다.
- HTTP Method는 REST 의미에 맞게 사용합니다.

## API Versioning 규칙
- 모든 API는 반드시 Version Group 하위에 위치합니다. 
```go
    v1Group := engine.Group("/v1")
```

## Middleware 적용 규칙
- 인증/인가 관련 미들웨어는 Version Group 레벨에서 적용합니다.

## Resource Grouping 규칙
- 도메인별로 Resource Group을 분리합니다.