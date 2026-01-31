# 4단계 - Inference API 구현 (Rule 기반 위험 스코어)

## 💬 대화 내용 요약

### 사용자 요청
1. **초기 요청**: 기능 요구사항 2-3-(1) Inference 요청 API 구현
   - Endpoint: POST /api/v1/inference/vital-risk
   - Request: { "patient_id": "P00001234" }
   - 환경변수로 설정된 시간 범위 내의 Vital 데이터 조회
   - 평균값 기반 위험 스코어 계산

2. **평가 규칙**:
   - HR > 120: 위험 증가
   - SBP < 90: 위험 증가
   - SpO2 < 90: 위험 증가
   - 충족 조건 개수에 따른 위험 등급:
     - 0개: LOW
     - 1-2개: MEDIUM
     - 3개 이상: HIGH

3. **환경변수**:
   - VITAL_RISK_TIME_WINDOW_HOURS: 조회 시간 범위 (기본값: 24시간)

4. **Response 구조**:
   - patient_id, risk_level, triggered_rules
   - vital_averages, data_points_analyzed
   - time_range, evaluated_at

5. **리팩토링 요청**: 코드 품질 개선
   - **환경변수 처리 개선**: `/library/envs/envs.go` 패턴 적용
     - 기존: 서비스에서 `os.Getenv` 직접 호출
     - 개선: `envs.VitalRiskTimeWindowHours` 사용으로 중앙화
   - **상수 정의 및 재사용**: `/api-server/pkg/constant/vital.go` 생성
     - 기존: 하드코딩된 vital type 문자열 ("HR", "SBP", "SpO2")
     - 개선: `constant.VitalTypeHR`, `constant.VitalTypeSBP`, `constant.VitalTypeSpO2` 사용
     - 전체 vital type 상수 정의: HR, SBP, DBP, SpO2, RR, BT
   - 테스트 코드 업데이트 및 검증 완료

## 🤔 설계 고민 과정

### 1. 도메인 분리: Inference를 어디에 구현할 것인가?

#### 문제 상황
Inference API는 Vital 데이터를 조회하고 분석하지만, 새로운 비즈니스 기능입니다. 어느 도메인에 구현해야 할까?

#### 고민 과정

**옵션 1: vital 도메인에 포함**
```
/api-server/domain/vital/
  - service.go: UpsertVital, GetVitals, CalculateRisk
```
- ❌ 문제점:
  - vital 도메인이 "데이터 관리"와 "분석/추론" 두 가지 책임을 가짐
  - 향후 다른 inference 기능(예: 질병 예측) 추가 시 vital 도메인이 비대해짐
  - 단일 책임 원칙(SRP) 위반
  - 확장성 저하

**옵션 2: patient 도메인에 포함**
```
/api-server/domain/patient/
  - service.go: GetPatientVitals, CalculateRisk
```
- ❌ 문제점:
  - patient 도메인이 "환자 정보 관리"와 "위험 스코어 계산" 두 가지 책임
  - Inference 로직이 환자 관리와 관련 없음
  - 도메인 경계가 모호해짐

**옵션 3: 독립적인 inference 도메인 생성 (✅ 선택)**
```
/api-server/domain/inference/
  - service.go: CalculateVitalRisk, (향후) PredictDisease, etc.
  - controller.go: InferenceController
```
- ✅ 장점:
  - 명확한 책임 분리: 데이터 관리 vs 분석/추론
  - 확장성 우수: 새로운 inference 기능 추가 용이
  - 도메인 경계 명확
  - 단일 책임 원칙 준수
  - 재사용성 향상

#### 최종 설계 결정

```
┌─────────────────────────────────────────────────────┐
│         Inference Domain (새로 생성)                  │
├─────────────────────────────────────────────────────┤
│ Controller: HTTP 요청/응답 처리                        │
│ - Request Body 바인딩                                 │
│ - 에러 처리 및 응답                                    │
├─────────────────────────────────────────────────────┤
│ Service: 위험 스코어 계산 로직                          │
│ - 환경변수에서 시간 범위 읽기                           │
│ - Vital Repository 호출 (도메인 간 협력)               │
│ - 평균 계산 및 위험 조건 평가                           │
│ - risk_level 결정                                     │
└─────────────────────────────────────────────────────┘
                        ↓ (의존)
┌─────────────────────────────────────────────────────┐
│             Vital Domain (기존)                       │
├─────────────────────────────────────────────────────┤
│ Repository: Vital 데이터 조회                          │
│ - FindVitalsByPatientIDAndDateRange                 │
└─────────────────────────────────────────────────────┘
```

**설계 원칙 적용**:
- **Single Responsibility**: Inference 도메인은 "분석/추론"만 담당
- **Open/Closed**: 새로운 inference 기능 추가 시 기존 코드 수정 없이 확장
- **Dependency Inversion**: Inference Service는 VitalRepository 인터페이스에 의존

### 2. 환경변수 처리: 어디서 읽고 관리할 것인가?

#### 문제 상황
VITAL_RISK_TIME_WINDOW_HOURS 환경변수를 어디서 읽어야 할까?

#### 고민 과정

**옵션 1: Controller에서 읽기**
```go
func (i *inferenceController) CalculateVitalRisk(ctx *gin.Context) {
    timeWindowHours := getEnvInt("VITAL_RISK_TIME_WINDOW_HOURS", 24)
    // ...
}
```
- ❌ 문제점:
  - Controller는 HTTP 계층 처리만 담당해야 함
  - 비즈니스 로직(시간 범위 설정)이 Controller에 노출

**옵션 2: Service에서 읽기 (✅ 선택)**
```go
func (i *inferenceService) CalculateVitalRisk(...) {
    timeWindowHours := 24  // 기본값
    if envValue := os.Getenv("VITAL_RISK_TIME_WINDOW_HOURS"); envValue != "" {
        if hours, err := strconv.Atoi(envValue); err == nil && hours > 0 {
            timeWindowHours = hours
        }
    }
    // ...
}
```
- ✅ 장점:
  - Service가 비즈니스 로직과 설정 관리
  - Controller는 HTTP 처리에만 집중
  - 환경변수 파싱 오류 처리 가능

**옵션 3: 별도 Config 패키지**
```go
type Config struct {
    VitalRiskTimeWindowHours int
}
```
- 🤔 고려사항:
  - 설정이 많아지면 Config 패키지 도입 고려
  - 현재는 단일 설정이므로 Service에서 처리가 더 간단

#### 최종 설계 결정

**Service 레벨에서 환경변수 처리**:
```go
// 기본값 설정 + 환경변수 우선
timeWindowHours := 24
if envValue := os.Getenv("VITAL_RISK_TIME_WINDOW_HOURS"); envValue != "" {
    if hours, err := strconv.Atoi(envValue); err == nil && hours > 0 {
        timeWindowHours = hours
    }
}
```

**장점**:
- 기본값 보장 (환경변수 없거나 잘못된 값일 때)
- 런타임 유연성 (환경변수 변경으로 동작 조정 가능)
- 오류 허용 (파싱 실패 시 기본값 사용)

### 3. 평균 계산 로직: 어떻게 처리할 것인가?

#### 문제 상황
조회된 Vital 데이터에서 타입별(HR, SBP, SpO2) 평균을 계산해야 합니다.

#### 고민 과정

**옵션 1: Repository에서 SQL AVG 사용**
```sql
SELECT vital_type, AVG(value)
FROM vitals
WHERE ...
GROUP BY vital_type
```
- 🤔 장단점:
  - 장점: DB에서 집계하여 성능 우수
  - 단점: Repository가 집계 로직 포함, 유연성 저하
  - 문제: 향후 가중 평균, 중앙값 등 복잡한 로직 추가 어려움

**옵션 2: Service에서 Application 레벨 계산 (✅ 선택)**
```go
// 타입별 데이터 수집
vitalData := make(map[string][]float64)
for _, v := range vitals {
    vitalData[v.VitalType] = append(vitalData[v.VitalType], v.Value)
}

// 평균 계산
for vitalType, values := range vitalData {
    sum := 0.0
    for _, val := range values {
        sum += val
    }
    vitalAverages[vitalType] = sum / float64(len(values))
}
```
- ✅ 장점:
  - 계산 로직이 Service에 집중 (비즈니스 로직)
  - Repository는 데이터 조회만 담당 (단일 책임)
  - 향후 복잡한 계산 로직 추가 용이
  - 테스트 가능성 높음

#### 최종 설계 결정

**Service 레벨에서 평균 계산**:
1. Repository로부터 raw data 조회
2. map[string][]float64로 타입별 데이터 그룹핑
3. 각 타입별 평균 계산
4. 평균값을 기반으로 위험 조건 평가

**확장성 고려**:
- 향후 중앙값, 최대/최소값, 표준편차 등 추가 가능
- 가중 평균 (최근 데이터에 더 높은 가중치) 적용 가능

### 4. 위험 등급 결정: 규칙 엔진 vs 하드코딩

#### 문제 상황
위험 조건 평가 및 risk_level 결정 로직을 어떻게 구현할 것인가?

#### 고민 과정

**옵션 1: 규칙 엔진 (Rule Engine)**
```go
type Rule interface {
    Evaluate(averages map[string]float64) bool
    Description() string
}

type HRRule struct{}
func (r *HRRule) Evaluate(avg map[string]float64) bool {
    return avg["HR"] > 120
}
```
- 🤔 장단점:
  - 장점: 규칙 동적 추가/수정 가능
  - 단점: 현재 요구사항(3개 규칙)에 비해 과도하게 복잡

**옵션 2: 직접 구현 (✅ 선택)**
```go
var triggeredRules []string

// HR > 120
if avg, exists := vitalAverages["HR"]; exists && avg > 120 {
    triggeredRules = append(triggeredRules, "HR > 120")
}

// SBP < 90
if avg, exists := vitalAverages["SBP"]; exists && avg < 90 {
    triggeredRules = append(triggeredRules, "SBP < 90")
}

// SpO2 < 90
if avg, exists := vitalAverages["SpO2"]; exists && avg < 90 {
    triggeredRules = append(triggeredRules, "SpO2 < 90")
}
```
- ✅ 장점:
  - 코드 명확성 (규칙이 한눈에 보임)
  - 현재 요구사항에 적합
  - 테스트 용이
  - YAGNI 원칙 준수 (필요하지 않으면 만들지 않음)

#### 최종 설계 결정

**직접 구현 방식 선택**:
- 규칙이 명확하고 단순 (3개)
- 향후 규칙이 크게 증가하면 리팩토링 고려
- 현재는 가독성과 유지보수성이 더 중요

### 5. Response 구조: 어떤 정보를 포함할 것인가?

#### 최종 Response 설계
```go
type VitalRiskResponse struct {
    PatientID          string            `json:"patient_id"`
    RiskLevel          string            `json:"risk_level"`          // LOW/MEDIUM/HIGH
    TriggeredRules     []string          `json:"triggered_rules"`     // 충족된 조건 목록
    VitalAverages      map[string]float64 `json:"vital_averages"`      // 타입별 평균값
    DataPointsAnalyzed int               `json:"data_points_analyzed"` // 분석된 데이터 개수
    TimeRange          TimeRange         `json:"time_range"`          // 분석 시간 범위
    EvaluatedAt        time.Time         `json:"evaluated_at"`        // 평가 시점
}
```

**설계 이유**:
- **투명성**: 어떤 데이터를 기반으로 계산했는지 명확히 제공
- **재현성**: time_range와 evaluated_at로 결과 재현 가능
- **설명 가능성**: triggered_rules와 vital_averages로 위험 등급 근거 제시
- **신뢰성**: data_points_analyzed로 분석 데이터 양 확인 가능

## 구현 내용

### 1. Domain Layer (Interface 정의)

#### inference/param.go
```go
type VitalRiskRequest struct {
    PatientID string `json:"patient_id" binding:"required"`
}

type VitalRiskResponse struct {
    PatientID          string            `json:"patient_id"`
    RiskLevel          string            `json:"risk_level"`
    TriggeredRules     []string          `json:"triggered_rules"`
    VitalAverages      map[string]float64 `json:"vital_averages"`
    DataPointsAnalyzed int               `json:"data_points_analyzed"`
    TimeRange          TimeRange         `json:"time_range"`
    EvaluatedAt        time.Time         `json:"evaluated_at"`
}

type TimeRange struct {
    From time.Time `json:"from"`
    To   time.Time `json:"to"`
}
```

#### inference/service.go
```go
type InferenceService interface {
    CalculateVitalRisk(ctx context.Context, request VitalRiskRequest) (*VitalRiskResponse, error)
}
```

#### inference/controller.go
```go
type InferenceController interface {
    CalculateVitalRisk(ctx *gin.Context)
}
```

### 2. Service Layer (비즈니스 로직)
**파일**: `/api-server/app/service/inference_service.go`

**주요 개선 사항 (리팩토링)**:
- 환경변수 처리: `envs.VitalRiskTimeWindowHours` 사용 (중앙화)
- Vital Type 상수: `constant.VitalTypeHR/SBP/SpO2` 사용 (재사용성)

```go
type inferenceService struct {
    vitalRepo vital.VitalRepository  // vital repository 주입
}

func (i *inferenceService) CalculateVitalRisk(
    ctx context.Context,
    request inference.VitalRiskRequest,
) (*inference.VitalRiskResponse, error) {
    // 1. 환경변수에서 시간 범위 읽기 (envs 패키지 사용)
    timeWindowHours := envs.VitalRiskTimeWindowHours

    // 2. 현재 시간 기준으로 시간 범위 설정
    now := time.Now().UTC()
    from := now.Add(-time.Duration(timeWindowHours) * time.Hour)
    to := now

    // 3. Vital 데이터 조회
    vitals, err := i.vitalRepo.FindVitalsByPatientIDAndDateRange(
        ctx, request.PatientID, from, to, "")
    if err != nil {
        return nil, pkgError.WrapWithCode(err, pkgError.Get)
    }

    // 4. 각 Vital Type별로 데이터 수집 (상수 사용)
    vitalData := make(map[string][]float64)
    for _, v := range vitals {
        if v.VitalType == constant.VitalTypeHR ||
           v.VitalType == constant.VitalTypeSBP ||
           v.VitalType == constant.VitalTypeSpO2 {
            vitalData[v.VitalType] = append(vitalData[v.VitalType], v.Value)
        }
    }

    // 5. 각 Vital Type별 평균 계산
    vitalAverages := make(map[string]float64)
    for vitalType, values := range vitalData {
        if len(values) > 0 {
            sum := 0.0
            for _, val := range values {
                sum += val
            }
            vitalAverages[vitalType] = sum / float64(len(values))
        }
    }

    // 6. 위험 조건 평가
    var triggeredRules []string

    if avg, exists := vitalAverages["HR"]; exists && avg > 120 {
        triggeredRules = append(triggeredRules, "HR > 120")
    }
    if avg, exists := vitalAverages["SBP"]; exists && avg < 90 {
        triggeredRules = append(triggeredRules, "SBP < 90")
    }
    if avg, exists := vitalAverages["SpO2"]; exists && avg < 90 {
        triggeredRules = append(triggeredRules, "SpO2 < 90")
    }

    // 7. risk_level 결정
    riskLevel := "LOW"
    triggeredCount := len(triggeredRules)
    if triggeredCount >= 3 {
        riskLevel = "HIGH"
    } else if triggeredCount >= 1 {
        riskLevel = "MEDIUM"
    }

    return &inference.VitalRiskResponse{
        PatientID:          request.PatientID,
        RiskLevel:          riskLevel,
        TriggeredRules:     triggeredRules,
        VitalAverages:      vitalAverages,
        DataPointsAnalyzed: len(vitals),
        TimeRange: inference.TimeRange{
            From: from,
            To:   to,
        },
        EvaluatedAt: now,
    }, nil
}
```

**핵심 로직**:
1. 환경변수에서 시간 범위 읽기 (오류 허용 설계)
2. 날짜 범위로 Vital 데이터 조회 (도메인 간 협력)
3. 타입별 그룹핑 및 평균 계산 (Application 레벨 계산)
4. 위험 조건 평가 (명확한 if 문)
5. 충족 개수에 따른 위험 등급 결정

### 3. Controller Layer (HTTP 요청/응답)
**파일**: `/api-server/app/controller/inference_controller.go`

```go
// CalculateVitalRisk
// @Title CalculateVitalRisk
// @Description Vital 데이터 기반 위험 스코어 계산
// @Tags V1 - Inference
// @Accept json
// @Produce json
// @Param reqBody body inference.VitalRiskRequest true "위험 스코어 계산 요청"
// @Success 200 {object} output.Output{data=inference.VitalRiskResponse}
// @Failure 400 {object} output.Output "code: 400001 - Wrong parameter"
// @Failure 500 {object} output.Output "code: 100003 - Fail to get data from db"
// @Router /v1/inference/vital-risk [Post]
func (i *inferenceController) CalculateVitalRisk(ctx *gin.Context) {
    var reqBody inference.VitalRiskRequest
    if err := ctx.ShouldBindJSON(&reqBody); err != nil {
        output.AppendErrorContext(ctx, pkgError.WrapWithCode(
            err, pkgError.WrongParam, err.Error(), "fail to parse request parameter"), nil)
        return
    }

    result, err := i.service.CalculateVitalRisk(ctx, reqBody)
    if err != nil {
        output.AppendErrorContext(ctx, pkgError.Wrap(err), nil)
        return
    }

    output.Send(ctx, result)
}
```

### 4. Router Layer (라우팅)
**파일**: `/api-server/app/router/inference_router.go`

```go
func NewInferenceRouter(engine *gin.Engine, controller inference.InferenceController) {
    v1Group := engine.Group("/api/v1")
    v1Group.Use(middleware.ValidTokenMiddleware())

    inferenceGroup := v1Group.Group("/inference")
    {
        inferenceGroup.POST("/vital-risk", controller.CalculateVitalRisk)
    }
}
```

### 5. 환경변수 관리 (리팩토링)
**파일**: `/library/envs/envs.go`

리팩토링을 통해 환경변수 관리를 중앙화:
```go
var (
    // ... 기존 환경변수들 ...

    VitalRiskTimeWindowHours = getEnvAsInt("VITAL_RISK_TIME_WINDOW_HOURS", 24)
)

func getEnvAsInt(envName string, defaultVal int) int {
    envVal := os.Getenv(envName)
    if envVal == "" {
        return defaultVal
    }
    if intVal, err := strconv.Atoi(envVal); err == nil && intVal > 0 {
        return intVal
    }
    return defaultVal
}
```

**리팩토링 이점**:
- 환경변수 관리의 중앙화: 모든 환경변수가 한 곳에서 관리됨
- 기본값 명시: 코드에서 기본값을 명확히 확인 가능
- 타입 안정성: int 타입으로 직접 제공하여 변환 로직 중복 제거
- 재사용성: 다른 서비스에서도 `envs.VitalRiskTimeWindowHours` 직접 사용 가능

### 6. Vital Type 상수 정의 (리팩토링)
**파일**: `/api-server/pkg/constant/vital.go`

Vital Type을 재사용 가능한 상수로 정의:
```go
package constant

// Vital Type 상수 정의
const (
    VitalTypeHR   = "HR"   // Heart Rate (심박수)
    VitalTypeSBP  = "SBP"  // Systolic Blood Pressure (수축기 혈압)
    VitalTypeDBP  = "DBP"  // Diastolic Blood Pressure (이완기 혈압)
    VitalTypeSpO2 = "SpO2" // Oxygen Saturation (산소포화도)
    VitalTypeRR   = "RR"   // Respiratory Rate (호흡수)
    VitalTypeBT   = "BT"   // Body Temperature (체온)
)
```

**리팩토링 이점**:
- 문자열 하드코딩 제거: 타입 실수 방지
- 재사용성: 전체 프로젝트에서 일관된 vital type 사용
- 유지보수성: 상수 변경 시 한 곳만 수정
- 확장성: 새로운 vital type 추가 용이
- IDE 지원: 자동완성 및 리팩토링 도구 활용 가능

## 테스트 케이스

### Service Layer (inference_service_test.go)
1. ✅ 성공 - HIGH 위험 (모든 조건 충족)
2. ✅ 성공 - MEDIUM 위험 (2개 조건 충족)
3. ✅ 성공 - MEDIUM 위험 (1개 조건 충족)
4. ✅ 성공 - LOW 위험 (조건 충족 없음)
5. ✅ 성공 - 데이터 없음 (LOW)
6. ✅ 실패 - Repository 에러
7. ✅ 환경변수 테스트 (48시간)

### Controller Layer (inference_controller_test.go)
1. ✅ 성공 - HIGH 위험
2. ✅ 성공 - MEDIUM 위험
3. ✅ 성공 - LOW 위험
4. ✅ 실패 - patient_id 필드 없음
5. ✅ 실패 - 잘못된 JSON
6. ✅ 실패 - Service 에러 (DB 조회 실패)

**테스트 실행 결과**:
```
✅ Inference Service: 모든 테스트 통과 (7개 케이스)
✅ Inference Controller: 모든 테스트 통과 (6개 케이스)
```

## API 요청/응답 예시

### HIGH 위험 케이스
**Request**:
```http
POST /api/v1/inference/vital-risk
Authorization: Bearer <token>
Content-Type: application/json

{
  "patient_id": "P00001234"
}
```

**Response**:
```http
HTTP/1.1 200 OK

{
  "success": true,
  "data": {
    "patient_id": "P00001234",
    "risk_level": "HIGH",
    "triggered_rules": [
      "HR > 120",
      "SBP < 90",
      "SpO2 < 90"
    ],
    "vital_averages": {
      "HR": 135.2,
      "SBP": 82.5,
      "SpO2": 87.3
    },
    "data_points_analyzed": 48,
    "time_range": {
      "from": "2025-11-30T10:20:00Z",
      "to": "2025-12-01T10:20:00Z"
    },
    "evaluated_at": "2025-12-01T10:20:00Z"
  }
}
```

### MEDIUM 위험 케이스
**Response**:
```http
{
  "success": true,
  "data": {
    "patient_id": "P00001234",
    "risk_level": "MEDIUM",
    "triggered_rules": [
      "HR > 120"
    ],
    "vital_averages": {
      "HR": 130.5,
      "SBP": 110.0,
      "SpO2": 95.0
    },
    "data_points_analyzed": 24,
    "time_range": {
      "from": "2025-11-30T10:20:00Z",
      "to": "2025-12-01T10:20:00Z"
    },
    "evaluated_at": "2025-12-01T10:20:00Z"
  }
}
```

### LOW 위험 케이스
**Response**:
```http
{
  "success": true,
  "data": {
    "patient_id": "P00001234",
    "risk_level": "LOW",
    "triggered_rules": [],
    "vital_averages": {
      "HR": 80.0,
      "SBP": 115.0,
      "SpO2": 97.0
    },
    "data_points_analyzed": 24,
    "time_range": {
      "from": "2025-11-30T10:20:00Z",
      "to": "2025-12-01T10:20:00Z"
    },
    "evaluated_at": "2025-12-01T10:20:00Z"
  }
}
```

### 에러 응답 예시

#### 필수 필드 누락
**Request**:
```http
POST /api/v1/inference/vital-risk
{}
```

**Response**:
```http
HTTP/1.1 400 Bad Request

{
  "code": 400001,
  "message": "wrong parameter",
  "detail": ["Key: 'VitalRiskRequest.PatientID' Error:Field validation for 'PatientID' failed on the 'required' tag"]
}
```

## 환경변수 설정

### 기본값 사용 (24시간)
```bash
# 환경변수 설정 안 함
# 기본적으로 최근 24시간 데이터 조회
```

### 사용자 정의 시간 범위 (48시간)
```bash
export VITAL_RISK_TIME_WINDOW_HOURS=48
```

### 환경변수 처리 로직
```go
// 오류 허용 설계
timeWindowHours := 24  // 기본값
if envValue := os.Getenv("VITAL_RISK_TIME_WINDOW_HOURS"); envValue != "" {
    if hours, err := strconv.Atoi(envValue); err == nil && hours > 0 {
        timeWindowHours = hours
    }
}
```

**안전장치**:
- 환경변수 없으면 기본값 사용
- 파싱 실패 시 기본값 사용
- 0 이하 값은 무시하고 기본값 사용

## 준수한 설계 규칙

### Domain Layer
- ✅ 새로운 inference 도메인 생성 (확장성)
- ✅ Interface 추상화로 계층 분리
- ✅ Request/Response DTO 정의

### Service Layer
- ✅ Context를 첫 번째 인자로 전달
- ✅ vital repository 주입 (도메인 간 협력)
- ✅ 환경변수 처리 (기본값 보장)
- ✅ 비즈니스 로직 수행 (평균 계산, 위험 평가)
- ✅ 에러를 pkgError.WrapWithCode로 래핑

### Controller Layer
- ✅ HTTP 요청/응답만 처리
- ✅ ctx.ShouldBindJSON 사용
- ✅ pkgError.WrapWithCode로 에러 래핑
- ✅ output.Send로 응답
- ✅ Swagger 주석 포함

### Router Layer
- ✅ Version Group (/api/v1) 하위 위치
- ✅ REST 의미에 맞는 HTTP Method (POST)
- ✅ 도메인별 Resource Group 분리 (/inference)

## 주요 학습 사항

1. **도메인 분리 전략**: 새로운 비즈니스 기능은 독립 도메인으로 분리
2. **환경변수 처리**: 기본값 보장과 오류 허용 설계
3. **Application 레벨 계산**: DB 집계 vs Service 계산 trade-off
4. **규칙 엔진 vs 직접 구현**: YAGNI 원칙 준수
5. **설명 가능한 AI**: triggered_rules와 vital_averages로 결과 근거 제공
6. **도메인 간 협력**: Inference Service → Vital Repository

## 설계 원칙 적용

### Single Responsibility Principle (단일 책임 원칙)
- inference 도메인: 분석/추론만 담당
- vital 도메인: 데이터 관리만 담당

### Open/Closed Principle (개방-폐쇄 원칙)
- 새로운 inference 기능 추가 시 기존 코드 수정 없이 확장
- 새로운 위험 조건 추가는 Service 코드만 수정

### Dependency Inversion Principle (의존성 역전 원칙)
- InferenceService는 VitalRepository **인터페이스**에 의존
- 생성자를 통한 의존성 주입

### YAGNI (You Aren't Gonna Need It)
- 규칙 엔진 대신 직접 구현 (현재 요구사항에 적합)
- 향후 필요 시 리팩토링

## 다음 단계 가능한 확장

- [ ] 가중 평균 (최근 데이터에 더 높은 가중치)
- [ ] 추세 분석 (위험도 증가/감소 추세)
- [ ] 알림 기능 (HIGH 위험 시 알림)
- [ ] 규칙 동적 설정 (DB 또는 Config 파일)
- [ ] 다른 inference 기능 (질병 예측, 이상 탐지 등)
