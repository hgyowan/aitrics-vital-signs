# repository layer 작성시 필수 참고 규칙

## Repository의 역할과 책임
- Repository는 DB 접근 및 영속성 처리만 담당합니다. 
- 각 도메인 repository 구현체로 gorm 을 활용한 DB CRUD 로직을 구현합니다.

## Context 전달 규칙
- 모든 Repository 메서드는 반드시 context.Context를 첫 번째 인자로 받습니다. 
- DB 호출 시 반드시 WithContext(ctx)를 사용합니다.

## 반환값 규칙
- Repository 메서드는 다음 중 하나만 반환합니다. 
- error
- (model, error)
- (slice, error)
- (slice, totalCount, error)
- DB 결과를 가공하거나 변환하지 않습니다.

## 네이밍 규칙
- Repository 메서드는 DB 행위 중심으로 명명합니다. (예: CreatePatient, FindByID, FindByCondition, UpdatePatient)
