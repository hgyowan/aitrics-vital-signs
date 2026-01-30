# service layer 작성시 필수 참고 규칙

## Service의 역할과 책임
- Service는 비즈니스 로직과 도메인 규칙을 담당합니다.
- Controller와 Repository 사이의 중재자 역할을 수행합니다.
- 데이터 조합, 값 생성, 상태 판단은 Service에서만 수행합니다.
- 각 도메인 service 구현체로 실제 비즈니스 로직을 구현합니다.

## Context 전달 규칙
- 모든 Service 메서드는 반드시 context.Context를 첫 번째 인자로 받습니다.

## Request 처리 규칙
- Controller로부터 전달받은 Request Param 은 Service에서 해석합니다.

## 도메인 모델 생성 규칙
- DB에 저장될 도메인 모델은 Service에서 생성합니다. 
- ID, Timestamp, 상태값은 Service에서 명시적으로 설정합니다.
```go
    &patient.Patient{
        ID: uuid.NewString(),
        CreatedAt: now,
        UpdatedAt: &now,
    }
```

## Error 처리 규칙
- Service는 비즈니스 의미가 있는 에러 변환을 담당합니다.
- 오류는 반드시 ../library/error/code.go 경로의 에러 코드로 래핑합니다.