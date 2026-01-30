# controller layer 작성시 필수 참고 규칙

## 아키텍처 및 책임 분리
- Controller는 HTTP 요청/응답 처리에만 책임을 둡니다.
- 비즈니스 로직은 반드시 Service 레이어로 위임합니다.
- Controller에서 DB, Repository, 외부 API를 직접 호출하지 않습니다.

## Controller 구조 규칙
- Controller는 반드시 struct + constructor 패턴을 사용합니다.

## Request 바인딩 규칙
- Request Body는 ctx.ShouldBindJSON만 사용합니다. 
- 바인딩 실패 시 즉시 종료합니다. 
- 파라미터 오류는 반드시 WrongParam 에러 코드로 래핑합니다.

## 에러 처리 규칙
- 모든 에러는 pkgError.Wrap 또는 pkgError.WrapWithCode를 사용합니다.
- 에러 응답은 반드시 output.AppendErrorContext를 통해 반환합니다.

## Response 규칙
- 성공 응답은 반드시 output.Send(ctx, data) 형식을 사용합니다. 
- 데이터가 없는 경우 nil을 전달합니다.

## Swagger / OpenAPI 주석 규칙
- 모든 API 메서드는 Swagger 주석을 포함해야 합니다.
- 필수 항목:
  - Title 
  - Description (한글)
  - Tags (도메인)
  - Param (domain/{도메인}/param.go)
  - Success / Failure 
  - Router