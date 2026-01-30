# 1단계 - 환자 정보 수정 API 구현 (Optimistic Lock 적용)

## 💬 대화 내용 요약

### 사용자 요청
1. **초기 요청**: 프로젝트의 모든 CLAUDE.md 파일 확인 후, 환자 정보 수정 API 구현
   - 기존 환자 등록 API 코드를 참고하여 동일한 패턴으로 구현
   - Optimistic Lock 적용 필수

2. **추가 요청사항**:
   - `FindByPatientID` → `FindPatientByID`로 메서드명 변경 (객체를 명시)
   - 동시성 테스트 작성 요청 (Optimistic Lock 검증)

## 작업 개요
AITRICS 과제 요구사항 2-1-(2) 환자 정보 수정 API 구현
- **Endpoint**: PUT /api/v1/patients/{patient_id}
- **주요 기능**: Optimistic Lock을 통한 동시성 제어
- **에러 처리**: DB version과 요청 version 불일치 시 409 Conflict 반환

## 구현 내용

### 1. Domain Layer (Interface 정의)
**파일**: `/api-server/domain/patient/`

#### param.go
```go
type UpdatePatientRequest struct {
    Name      string `json:"name" binding:"required"`
    Gender    string `json:"gender" binding:"required,oneof=M F"`
    BirthDate string `json:"birthDate" binding:"required,datetime=2006-01-02"`
    Version   int    `json:"version" binding:"required,min=1"`
}
```

#### repository.go
```go
type PatientRepository interface {
    CreatePatient(ctx context.Context, model *Patient) error
    FindPatientByID(ctx context.Context, patientID string) (*Patient, error)
    UpdatePatient(ctx context.Context, model *Patient) error
}
```

#### service.go
```go
type PatientService interface {
    CreatePatient(ctx context.Context, request CreatePatientRequest) error
    UpdatePatient(ctx context.Context, patientID string, request UpdatePatientRequest) error
}
```

#### controller.go
```go
type PatientController interface {
    CreatePatient(ctx *gin.Context)
    UpdatePatient(ctx *gin.Context)
}
```

### 2. Repository Layer (DB 접근)
**파일**: `/api-server/app/repository/patient_repository.go`

#### FindPatientByID
- PatientID로 환자 조회
- GORM의 First() 사용하여 단일 레코드 조회
- soft delete 자동 처리

#### UpdatePatient
- GORM의 Save() 사용하여 전체 레코드 업데이트
- Version 포함한 모든 필드 업데이트

**테스트 커버리지**: 87.5%

### 3. Service Layer (비즈니스 로직)
**파일**: `/api-server/app/service/patient_service.go`

#### UpdatePatient 로직
1. 날짜 파싱 및 검증
2. DB에서 기존 환자 정보 조회
3. **Optimistic Lock 검증**: `existingPatient.Version != request.Version`
4. Version mismatch 시 `Conflict` 에러 반환
5. 환자 정보 업데이트 (Version +1)
6. UpdatedAt 갱신
7. Repository를 통한 DB 업데이트

**핵심 코드**:
```go
if existingPatient.Version != request.Version {
    return pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.Conflict, "version mismatch")
}

existingPatient.Name = request.Name
existingPatient.Gender = request.Gender
existingPatient.BirthDate = birthDate
existingPatient.Version = request.Version + 1
existingPatient.UpdatedAt = &now
```

**테스트 커버리지**: 88.5%

### 4. Controller Layer (HTTP 요청/응답)
**파일**: `/api-server/app/controller/patient_controller.go`

#### UpdatePatient
- Path parameter에서 patient_id 추출 및 검증
- Request Body 바인딩 및 검증
- Service 호출
- 에러 처리 및 응답

**Swagger 주석**:
```go
// @Title UpdatePatient
// @Description 환자 정보 수정 (Optimistic Lock 적용)
// @Tags V1 - Patient
// @Accept json
// @Produce json
// @Param patient_id path string true "환자 ID"
// @Param reqBody body patient.UpdatePatientRequest true "환자 정보 수정 요청"
// @Success 200 {object} output.Output
// @Failure 400 {object} output.Output "code: 400001 - Wrong parameter"
// @Failure 409 {object} output.Output "code: 400002 - Version conflict"
// @Failure 500 {object} output.Output "code: 100002 - Fail to update data from db"
// @Router /v1/patients/{patient_id} [Put]
```

**테스트 커버리지**: 91.7%

### 5. Router Layer (라우팅)
**파일**: `/api-server/app/router/patient_router.go`

```go
patientGroup.PUT("/:patient_id", controller.UpdatePatient)
```

## 테스트 케이스

### Repository Layer
1. ✅ 성공 - 환자 조회
2. ✅ 실패 - 환자 없음
3. ✅ 성공 - 환자 정보 업데이트

### Service Layer
1. ✅ 성공 - 환자 정보 수정
2. ✅ 실패 - Version Conflict (Optimistic Lock)
3. ✅ 실패 - 환자 없음
4. ✅ 실패 - 날짜 파라미터 포맷 에러

### Controller Layer
1. ✅ 성공 - 환자 정보 수정
2. ✅ 실패 - 필수 필드 누락 (version 없음)
3. ✅ 실패 - Version Conflict (Optimistic Lock)
4. ✅ 실패 - 잘못된 Gender 값
5. ✅ 실패 - 잘못된 날짜 형식
6. ✅ 실패 - patient_id 파라미터 없음
7. ✅ 실패 - 비즈니스 로직 에러 (500)

## Optimistic Lock 구현 상세

### 동작 원리
1. 클라이언트가 환자 정보 조회 시 현재 version 값을 함께 받음
2. 수정 요청 시 조회했던 version 값을 함께 전송
3. 서버에서 DB의 현재 version과 요청의 version 비교
4. 일치하지 않으면 다른 트랜잭션이 수정했다는 의미로 409 Conflict 반환
5. 일치하면 수정 후 version을 +1 하여 저장

### 동시성 제어 시나리오
```
사용자 A: 환자 조회 (version=1) → 수정 요청 (version=1) → 성공, version=2로 저장
사용자 B: 환자 조회 (version=1) → 수정 요청 (version=1) → 실패 (DB는 이미 version=2)
```

## Mock 생성
```bash
cd /api-server/domain/patient
go generate ./...
```

생성된 파일:
- `/api-server/domain/mock/mock_repository.go`
- `/api-server/domain/mock/mock_service.go`
- `/api-server/domain/mock/mock_controller.go`

## 테스트 실행 결과

### 전체 테스트 통과
```
✅ Repository: 87.5% coverage
✅ Service: 88.5% coverage
✅ Controller: 91.7% coverage
```

모든 레이어에서 70% 이상의 테스트 커버리지 달성

## 준수한 설계 규칙

### Repository Layer
- ✅ Context를 첫 번째 인자로 전달
- ✅ WithContext(ctx) 사용
- ✅ DB 행위 중심 메서드명 (FindPatientByID, UpdatePatient)
- ✅ (model, error) 반환 형식

### Service Layer
- ✅ Context를 첫 번째 인자로 전달
- ✅ Request Param 해석
- ✅ 도메인 모델 생성 및 Timestamp 설정
- ✅ 에러를 pkgError.WrapWithCode로 래핑

### Controller Layer
- ✅ HTTP 요청/응답만 처리
- ✅ ctx.ShouldBindJSON 사용
- ✅ pkgError.WrapWithCode로 에러 래핑
- ✅ output.Send로 응답
- ✅ Swagger 주석 포함

### Router Layer
- ✅ Version Group (/v1) 하위 위치
- ✅ REST 의미에 맞는 HTTP Method (PUT)
- ✅ 도메인별 Resource Group 분리

## API 요청/응답 예시

### 성공 케이스
**Request**:
```http
PUT /api/v1/patients/P00001234
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "홍길동",
  "gender": "M",
  "birth_date": "1975-03-01",
  "version": 3
}
```

**Response**:
```http
HTTP/1.1 200 OK

{
  "success": true
}
```

### Version Conflict 케이스
**Request**:
```http
PUT /api/v1/patients/P00001234

{
  "name": "홍길동",
  "gender": "M",
  "birth_date": "1975-03-01",
  "version": 2  // DB의 version은 이미 3
}
```

**Response**:
```http
HTTP/1.1 409 Conflict

{
  "code": 400002,
  "message": "conflict data",
  "detail": ["version mismatch"]
}
```

## 주요 학습 사항

1. **Optimistic Lock 패턴**: Version 필드를 활용한 동시성 제어
2. **Table-Driven Test**: 다양한 시나리오를 체계적으로 테스트
3. **Mock 활용**: 각 레이어를 독립적으로 테스트
4. **Error Wrapping**: 비즈니스 의미있는 에러 코드 체계
5. **Layered Architecture**: 관심사의 분리를 통한 유지보수성 향상
