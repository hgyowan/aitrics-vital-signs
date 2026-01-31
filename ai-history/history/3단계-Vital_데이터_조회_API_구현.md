# 3단계 - Vital 데이터 조회 API 구현 (도메인 간 협력)

## 💬 대화 내용 요약

### 사용자 요청
1. **초기 요청**: 기능 요구사항 2-2-(2) Vital 데이터 조회 API 구현
   - Endpoint: GET /api/v1/patients/{patient_id}/vitals
   - Query Parameters: from, to (필수), vital_type (선택)
   - vital_type 유무에 따른 필터링 동작

2. **설계 요구사항**:
   - **도메인 분리**: patients의 vitals를 조회하는 것이므로 Controller/Router는 **patient 도메인**에 작성
   - **비즈니스 로직 분리**: Vital 데이터 조회/가공 로직은 **vital 도메인**에 작성
   - **서비스 간 협력**: patient service에서 vital service를 주입받아 사용
   - **테스트 코드 필수**: 모든 도메인(patient, vital)의 모든 레이어에 테스트 코드 작성

3. **응답 구조 요청**:
   - vital_type 있을 때와 없을 때의 **응답 구조는 동일**해야 함
   - vital_type이 없으면 모든 타입의 item이 포함되어야 함

## 🤔 설계 고민 과정

### 1. 도메인 경계와 책임 분리

#### 문제 상황
"/api/v1/patients/{patient_id}/vitals" 엔드포인트는:
- **환자(patient) 리소스의 하위 리소스**로서 vital을 조회
- 하지만 **vital 데이터 자체의 조회/가공 로직**이 필요

어느 도메인에 구현할 것인가?

#### 고민 과정

**옵션 1: 모든 로직을 patient 도메인에 구현**
```go
// patient service에서 vital repository를 직접 주입받아 사용
type patientService struct {
    repo patient.PatientRepository
    vitalRepo vital.VitalRepository  // vital repository 직접 의존
}
```
- ❌ 문제점:
  - patient service가 vital의 내부 구현에 직접 의존
  - vital 조회 로직이 patient 도메인에 노출
  - vital 도메인의 재사용성 저하
  - 도메인 경계가 모호해짐

**옵션 2: 모든 로직을 vital 도메인에 구현**
```go
// vital controller에 patient_id 파라미터 처리
func (v *vitalController) GetVitalsByPatientID(ctx *gin.Context)
```
- ❌ 문제점:
  - REST 리소스 계층 구조와 불일치 (/patients/{id}/vitals)
  - patient 컨텍스트를 vital 도메인에서 처리
  - 확장성 저하 (patient와 관련된 다른 정보 조합 시 어려움)

**옵션 3: 도메인 간 협력 (✅ 선택한 방식)**
```go
// patient service에서 vital service를 주입받아 사용
type patientService struct {
    repo patient.PatientRepository
    vitalService vital.VitalService  // service 레벨에서 협력
}
```
- ✅ 장점:
  - 각 도메인의 책임이 명확히 분리
  - vital 도메인은 vital 데이터 조회/가공에만 집중
  - patient 도메인은 HTTP 요청 처리 및 환자 컨텍스트 관리
  - 도메인 경계 명확, 재사용성 높음
  - 확장 가능 (추가 집계 로직 등)

#### 최종 설계 결정

```
┌─────────────────────────────────────────────────────┐
│            Patient Domain (HTTP Layer)               │
├─────────────────────────────────────────────────────┤
│ Controller: HTTP 요청/응답 처리                        │
│ - Path parameter 추출 (patient_id)                   │
│ - Query parameter 바인딩 (from, to, vital_type)     │
│ - 에러 처리 및 응답                                    │
├─────────────────────────────────────────────────────┤
│ Service: 환자 컨텍스트 관리                            │
│ - 날짜 파싱 및 검증                                    │
│ - Vital Service 호출 (도메인 간 협력)                  │
└─────────────────────────────────────────────────────┘
                        ↓ (의존)
┌─────────────────────────────────────────────────────┐
│             Vital Domain (Data Layer)                │
├─────────────────────────────────────────────────────┤
│ Service: Vital 데이터 조회/가공                        │
│ - Repository 호출                                     │
│ - 응답 DTO 변환                                       │
├─────────────────────────────────────────────────────┤
│ Repository: DB 접근                                   │
│ - 날짜 범위 조회                                       │
│ - vital_type 필터링 (선택적)                          │
└─────────────────────────────────────────────────────┘
```

### 2. 응답 구조 설계

#### 문제 상황
사용자 요구사항:
- vital_type이 **있을 때**: 해당 타입만 필터링
- vital_type이 **없을 때**: 모든 타입 포함
- 두 경우 모두 **응답 구조는 동일**해야 함

#### 고민 과정

**옵션 1: 타입별 그룹핑 응답**
```json
{
  "patient_id": "P00001234",
  "vitals": [
    {
      "vital_type": "HR",
      "items": [
        {"recorded_at": "...", "value": 110.0}
      ]
    }
  ]
}
```
- ❌ 문제점: vital_type 있을 때와 없을 때 구조가 달라짐

**옵션 2: Flat 구조 with vital_type 포함 (✅ 선택)**
```json
{
  "patient_id": "P00001234",
  "items": [
    {
      "vital_type": "HR",
      "recorded_at": "2025-12-01T10:15:00Z",
      "value": 110.0
    },
    {
      "vital_type": "RR",
      "recorded_at": "2025-12-01T10:15:00Z",
      "value": 20.0
    }
  ]
}
```
- ✅ 장점:
  - vital_type 유무와 관계없이 **동일한 구조**
  - vital_type이 있으면 같은 타입만, 없으면 여러 타입 포함
  - 클라이언트 처리 로직 단순화
  - 확장성 좋음

#### 최종 설계 결정

```go
// vital_type 유무와 관계없이 동일한 응답 구조
type GetVitalsResponse struct {
    PatientID string              `json:"patient_id"`
    Items     []VitalItemResponse `json:"items"`
}

type VitalItemResponse struct {
    VitalType  string    `json:"vital_type"`  // 항상 포함
    RecordedAt time.Time `json:"recorded_at"`
    Value      float64   `json:"value"`
}
```

### 3. Repository 쿼리 설계

#### 문제 상황
vital_type이 선택적 파라미터인데, WHERE 조건을 어떻게 구성할 것인가?

#### 고민 과정

**옵션 1: 두 개의 메서드 분리**
```go
FindVitalsByPatientIDAndDateRange(ctx, patientID, from, to) ([]Vital, error)
FindVitalsByPatientIDAndDateRangeAndType(ctx, patientID, from, to, vitalType) ([]Vital, error)
```
- ❌ 문제점: 중복 코드, 인터페이스 복잡도 증가

**옵션 2: 단일 메서드 with 조건부 WHERE (✅ 선택)**
```go
func (v *vitalRepository) FindVitalsByPatientIDAndDateRange(
    ctx context.Context,
    patientID string,
    from time.Time,
    to time.Time,
    vitalType string,  // 빈 문자열 허용
) ([]Vital, error) {
    query := db.Where("patient_id = ? AND recorded_at >= ? AND recorded_at <= ?",
                      patientID, from, to)

    // vitalType이 있으면 추가 조건
    if vitalType != "" {
        query = query.Where("vital_type = ?", vitalType)
    }

    return query.Order("recorded_at ASC").Find(&results).Error
}
```
- ✅ 장점:
  - 단일 메서드로 두 가지 경우 처리
  - 코드 중복 없음
  - 확장 가능 (추가 선택적 조건 처리 용이)

## 구현 내용

### 1. Domain Layer (Interface 정의)

#### vital/param.go
```go
type GetVitalsRequest struct {
    PatientID string
    From      time.Time
    To        time.Time
    VitalType string // optional: 있으면 해당 타입만, 없으면 모든 타입
}

// vital_type 유무와 관계없이 동일한 응답 구조
type GetVitalsResponse struct {
    PatientID string              `json:"patient_id"`
    Items     []VitalItemResponse `json:"items"`
}

type VitalItemResponse struct {
    VitalType  string    `json:"vital_type"`
    RecordedAt time.Time `json:"recorded_at"`
    Value      float64   `json:"value"`
}
```

#### patient/param.go
```go
type GetPatientVitalsQueryRequest struct {
    From      string `form:"from" binding:"required"`       // RFC3339 format
    To        string `form:"to" binding:"required"`         // RFC3339 format
    VitalType string `form:"vital_type" binding:"omitempty,oneof=HR RR SBP DBP SpO2 BT"`
}
```

#### vital/repository.go
```go
type VitalRepository interface {
    FindVitalByPatientIDAndRecordedAtAndVitalType(...) (*Vital, error)
    FindVitalsByPatientIDAndDateRange(ctx context.Context, patientID string, from time.Time, to time.Time, vitalType string) ([]Vital, error)  // 추가
    CreateVital(...) error
    UpdateVital(...) error
}
```

#### vital/service.go
```go
type VitalService interface {
    UpsertVital(...) error
    GetVitalsByPatientIDAndDateRange(ctx context.Context, request GetVitalsRequest) (*GetVitalsResponse, error)  // 추가
}
```

#### patient/service.go
```go
type PatientService interface {
    CreatePatient(...) error
    UpdatePatient(...) error
    GetPatientVitals(ctx context.Context, patientID string, request GetPatientVitalsQueryRequest) (*vital.GetVitalsResponse, error)  // 추가
}
```

#### patient/controller.go
```go
type PatientController interface {
    CreatePatient(ctx *gin.Context)
    UpdatePatient(ctx *gin.Context)
    GetPatientVitals(ctx *gin.Context)  // 추가
}
```

### 2. Repository Layer (DB 접근)
**파일**: `/api-server/app/repository/vital_repository.go`

```go
func (v *vitalRepository) FindVitalsByPatientIDAndDateRange(
    ctx context.Context,
    patientID string,
    from time.Time,
    to time.Time,
    vitalType string,
) ([]Vital, error) {
    var results []vital.Vital
    query := v.externalGormClient.MySQL().WithContext(ctx).
        Where("patient_id = ? AND recorded_at >= ? AND recorded_at <= ?",
              patientID, from, to)

    // vitalType이 있으면 해당 타입만 필터링
    if vitalType != "" {
        query = query.Where("vital_type = ?", vitalType)
    }

    if err := query.Order("recorded_at ASC").Find(&results).Error; err != nil {
        return nil, pkgError.Wrap(err)
    }

    return results, nil
}
```

**핵심**:
- 날짜 범위 필터링: `recorded_at >= ? AND recorded_at <= ?`
- 선택적 타입 필터링: `if vitalType != ""` 조건부 WHERE 추가
- 시간순 정렬: `ORDER BY recorded_at ASC`

### 3. Service Layer (비즈니스 로직)

#### vital_service.go
```go
func (v *vitalService) GetVitalsByPatientIDAndDateRange(
    ctx context.Context,
    request vital.GetVitalsRequest,
) (*vital.GetVitalsResponse, error) {
    // Repository에서 Vital 데이터 조회
    vitals, err := v.repo.FindVitalsByPatientIDAndDateRange(
        ctx,
        request.PatientID,
        request.From,
        request.To,
        request.VitalType,
    )
    if err != nil {
        return nil, pkgError.WrapWithCode(err, pkgError.Get)
    }

    // Response 변환
    items := make([]vital.VitalItemResponse, 0, len(vitals))
    for _, v := range vitals {
        items = append(items, vital.VitalItemResponse{
            VitalType:  v.VitalType,
            RecordedAt: v.RecordedAt,
            Value:      v.Value,
        })
    }

    return &vital.GetVitalsResponse{
        PatientID: request.PatientID,
        Items:     items,
    }, nil
}
```

**책임**:
- Repository 호출하여 데이터 조회
- Entity → Response DTO 변환
- 에러를 pkgError.Get 코드로 래핑

#### patient_service.go
```go
type patientService struct {
    repo         patient.PatientRepository
    vitalService vital.VitalService  // vital service 주입
}

func (p *patientService) GetPatientVitals(
    ctx context.Context,
    patientID string,
    request patient.GetPatientVitalsQueryRequest,
) (*vital.GetVitalsResponse, error) {
    // Query Parameter 날짜 파싱
    from, err := time.Parse(time.RFC3339, request.From)
    if err != nil {
        return nil, pkgError.WrapWithCode(err, pkgError.WrongParam, "invalid from date format")
    }

    to, err := time.Parse(time.RFC3339, request.To)
    if err != nil {
        return nil, pkgError.WrapWithCode(err, pkgError.WrongParam, "invalid to date format")
    }

    // Vital Service를 통해 데이터 조회
    return p.vitalService.GetVitalsByPatientIDAndDateRange(ctx, vital.GetVitalsRequest{
        PatientID: patientID,
        From:      from,
        To:        to,
        VitalType: request.VitalType,
    })
}

func NewPatientService(
    repo patient.PatientRepository,
    vitalService vital.VitalService,  // 생성자에서 주입
) patient.PatientService {
    return &patientService{
        repo:         repo,
        vitalService: vitalService,
    }
}
```

**책임**:
- Query parameter 날짜 파싱 및 검증
- Vital service 호출 (도메인 간 협력)
- 날짜 파싱 에러를 pkgError.WrongParam으로 래핑

### 4. Controller Layer (HTTP 요청/응답)
**파일**: `/api-server/app/controller/patient_controller.go`

```go
// GetPatientVitals
// @Title GetPatientVitals
// @Description 환자 Vital 데이터 조회
// @Tags V1 - Patient
// @Accept json
// @Produce json
// @Param patient_id path string true "환자 ID"
// @Param from query string true "조회 시작 시간 (RFC3339 format)"
// @Param to query string true "조회 종료 시간 (RFC3339 format)"
// @Param vital_type query string false "Vital 타입 (HR, RR, SBP, DBP, SpO2, BT)"
// @Success 200 {object} output.Output{data=vital.GetVitalsResponse}
// @Failure 400 {object} output.Output "code: 400001 - Wrong parameter"
// @Failure 500 {object} output.Output "code: 100003 - Fail to get data from db"
// @Router /v1/patients/{patient_id}/vitals [Get]
func (p *patientController) GetPatientVitals(ctx *gin.Context) {
    patientID := ctx.Param("patient_id")
    if patientID == "" {
        output.AppendErrorContext(ctx, pkgError.WrapWithCode(
            pkgError.EmptyBusinessError(),
            pkgError.WrongParam,
            "patient_id is required"), nil)
        return
    }

    var queryParams patient.GetPatientVitalsQueryRequest
    if err := ctx.ShouldBindQuery(&queryParams); err != nil {
        output.AppendErrorContext(ctx, pkgError.WrapWithCode(
            err,
            pkgError.WrongParam,
            err.Error(),
            "fail to parse query parameters"), nil)
        return
    }

    result, err := p.service.GetPatientVitals(ctx, patientID, queryParams)
    if err != nil {
        output.AppendErrorContext(ctx, pkgError.Wrap(err), nil)
        return
    }

    output.Send(ctx, result)
}
```

### 5. Router Layer (라우팅)
**파일**: `/api-server/app/router/patient_router.go`

```go
patientGroup := v1Group.Group("/patients")
{
    patientGroup.POST("", controller.CreatePatient)
    patientGroup.PUT("/:patient_id", controller.UpdatePatient)
    patientGroup.GET("/:patient_id/vitals", controller.GetPatientVitals)  // 추가
}
```

**REST 계층 구조**:
- `/api/v1/patients/{patient_id}` - 환자 리소스
- `/api/v1/patients/{patient_id}/vitals` - 환자의 하위 리소스 (vitals)

## 테스트 케이스

### Repository Layer (vital_repository_test.go)
1. ✅ 성공 - vital_type 있을 때 해당 타입만 조회
2. ✅ 성공 - vital_type 없을 때 모든 타입 조회
3. ✅ 성공 - 조회 결과 없음 (빈 배열 반환)

### Service Layer

#### vital_service_test.go
1. ✅ 성공 - vital_type 있을 때
2. ✅ 성공 - vital_type 없을 때 (모든 타입)
3. ✅ 성공 - 조회 결과 없음
4. ✅ 실패 - Repository 에러

#### patient_service_test.go
1. ✅ 성공 - vital_type 있을 때
2. ✅ 성공 - vital_type 없을 때 (모든 타입)
3. ✅ 실패 - 잘못된 from 날짜 형식
4. ✅ 실패 - 잘못된 to 날짜 형식
5. ✅ 실패 - Vital Service 에러

### Controller Layer (patient_controller_test.go)
1. ✅ 성공 - vital_type 있을 때
2. ✅ 성공 - vital_type 없을 때 (모든 타입)
3. ✅ 실패 - patient_id 파라미터 없음
4. ✅ 실패 - from 파라미터 없음
5. ✅ 실패 - to 파라미터 없음
6. ✅ 실패 - 잘못된 vital_type
7. ✅ 실패 - Service 에러 (잘못된 날짜 형식)
8. ✅ 실패 - Service 에러 (DB 조회 실패)

**테스트 실행 결과**:
```
✅ Repository: 모든 테스트 통과
✅ Vital Service: 모든 테스트 통과
✅ Patient Service: 모든 테스트 통과
✅ Controller: 모든 테스트 통과
```

## API 요청/응답 예시

### vital_type 있을 때 (HR만 조회)
**Request**:
```http
GET /api/v1/patients/P00001234/vitals?from=2025-12-01T10:00:00Z&to=2025-12-01T12:00:00Z&vital_type=HR
Authorization: Bearer <token>
```

**Response**:
```http
HTTP/1.1 200 OK

{
  "success": true,
  "data": {
    "patient_id": "P00001234",
    "items": [
      {
        "vital_type": "HR",
        "recorded_at": "2025-12-01T10:15:00Z",
        "value": 110.0
      },
      {
        "vital_type": "HR",
        "recorded_at": "2025-12-01T11:30:00Z",
        "value": 115.0
      }
    ]
  }
}
```

### vital_type 없을 때 (모든 타입 조회)
**Request**:
```http
GET /api/v1/patients/P00001234/vitals?from=2025-12-01T10:00:00Z&to=2025-12-01T12:00:00Z
Authorization: Bearer <token>
```

**Response**:
```http
HTTP/1.1 200 OK

{
  "success": true,
  "data": {
    "patient_id": "P00001234",
    "items": [
      {
        "vital_type": "HR",
        "recorded_at": "2025-12-01T10:15:00Z",
        "value": 110.0
      },
      {
        "vital_type": "RR",
        "recorded_at": "2025-12-01T10:15:00Z",
        "value": 20.0
      },
      {
        "vital_type": "HR",
        "recorded_at": "2025-12-01T11:30:00Z",
        "value": 115.0
      }
    ]
  }
}
```

**동일한 응답 구조**: vital_type 유무와 관계없이 `items` 배열에 `vital_type` 필드가 항상 포함됨

### 에러 응답 예시

#### 필수 파라미터 누락
**Request**:
```http
GET /api/v1/patients/P00001234/vitals?from=2025-12-01T10:00:00Z
```

**Response**:
```http
HTTP/1.1 400 Bad Request

{
  "code": 400001,
  "message": "wrong parameter",
  "detail": ["Key: 'GetPatientVitalsQueryRequest.To' Error:Field validation for 'To' failed on the 'required' tag"]
}
```

#### 잘못된 vital_type
**Request**:
```http
GET /api/v1/patients/P00001234/vitals?from=2025-12-01T10:00:00Z&to=2025-12-01T12:00:00Z&vital_type=INVALID
```

**Response**:
```http
HTTP/1.1 400 Bad Request

{
  "code": 400001,
  "message": "wrong parameter",
  "detail": ["Key: 'GetPatientVitalsQueryRequest.VitalType' Error:Field validation for 'VitalType' failed on the 'oneof' tag"]
}
```

## 준수한 설계 규칙

### Domain Layer
- ✅ Interface 추상화로 계층 분리
- ✅ Request/Response DTO 정의
- ✅ 도메인 간 협력을 위한 service 의존성 명시

### Repository Layer
- ✅ Context를 첫 번째 인자로 전달
- ✅ WithContext(ctx) 사용
- ✅ DB 행위 중심 메서드명 (FindVitalsByPatientIDAndDateRange)
- ✅ (slice, error) 반환 형식
- ✅ 조건부 WHERE 절 처리

### Service Layer
- ✅ Context를 첫 번째 인자로 전달
- ✅ Request 해석 및 비즈니스 로직 수행
- ✅ 도메인 간 협력 (patient service → vital service)
- ✅ Entity → Response DTO 변환
- ✅ 에러를 pkgError.WrapWithCode로 래핑
- ✅ 생성자를 통한 의존성 주입

### Controller Layer
- ✅ HTTP 요청/응답만 처리
- ✅ ctx.ShouldBindQuery 사용 (query parameter)
- ✅ pkgError.WrapWithCode로 에러 래핑
- ✅ output.Send로 응답
- ✅ Swagger 주석 포함 (path, query parameter 명시)

### Router Layer
- ✅ Version Group (/api/v1) 하위 위치
- ✅ REST 의미에 맞는 HTTP Method (GET)
- ✅ 도메인별 Resource Group 분리
- ✅ REST 계층 구조 (/{parent_resource}/{id}/{child_resource})

## 주요 학습 사항

1. **도메인 간 협력**: Service 레벨에서 다른 도메인의 service를 주입받아 사용하는 패턴
2. **책임 분리**: Controller/Router는 HTTP 계층 처리, Service는 비즈니스 로직 담당
3. **선택적 파라미터 처리**: 빈 문자열로 선택적 파라미터를 표현하고 조건부 WHERE 절로 처리
4. **일관된 응답 구조**: 조건부 동작(vital_type 유무)에도 동일한 응답 구조 유지
5. **RFC3339 날짜 형식**: Query parameter로 날짜 전달 시 표준 형식 사용
6. **테스트 격리**: Mock을 통한 계층별 독립 테스트

## 설계 원칙 적용

### Single Responsibility Principle (단일 책임 원칙)
- patient 도메인: HTTP 요청 처리 및 환자 컨텍스트 관리
- vital 도메인: Vital 데이터 조회 및 가공

### Dependency Inversion Principle (의존성 역전 원칙)
- patient service는 vital.VitalService **인터페이스**에 의존 (구현체가 아님)
- 생성자를 통한 의존성 주입으로 테스트 가능성 향상

### Open/Closed Principle (개방-폐쇄 원칙)
- 새로운 조회 조건 추가 시 기존 코드 수정 없이 확장 가능
- 선택적 파라미터 패턴으로 확장성 확보

## 다음 단계 가능한 확장

- [ ] 페이지네이션 추가 (limit, offset)
- [ ] 정렬 옵션 추가 (ASC/DESC)
- [ ] 집계 기능 추가 (평균, 최대, 최소)
- [ ] 캐싱 적용 (Redis)