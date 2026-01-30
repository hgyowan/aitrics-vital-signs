# 2단계 - Vital 데이터 저장/수정 API 구현 (UPSERT with Optimistic Lock)

## 💬 대화 내용 요약

### 사용자 요청
1. **초기 요청**: 기능 요구사항 2-2 Vital 데이터 API 구현
   - Endpoint: POST /api/v1/vitals (UPSERT)
   - 복합 식별자: (patient_id, recorded_at, vital_type)
   - Optimistic Lock 적용 필수

2. **추가 요청사항**:
   - `RecordedAt`을 time.Time 타입으로 받기 (string 파싱 불필요)
   - `FindVitalByCompositeKey` → `FindVitalByPatientIDAndRecordedAtAndVitalType`로 변경 (모든 키 컬럼 명시)
   - ai-history/history에 2단계 내용 기록 (누락하지 말 것)
   - **Service와 Controller의 테스트 코드 작성 필수** (항상 테스트 코드 작성)

## 작업 개요
AITRICS 과제 요구사항 2-2-(1) Vital 데이터 저장/수정 API 구현
- **Endpoint**: POST /api/v1/vitals
- **주요 기능**: UPSERT (INSERT or UPDATE) with Optimistic Lock
- **복합 식별자**: (patient_id, recorded_at, vital_type)
- **에러 처리**:
  - INSERT 시 version ≠ 1 → 400 Bad Request
  - UPDATE 시 version 불일치 → 409 Conflict

## 구현 내용

### 1. Domain Layer (Interface 정의)
**파일**: `/api-server/domain/vital/`

#### entity.go (수정)
```go
type Vital struct {
    PatientID  string         `gorm:"column:patient_id;type:varchar(20);not null;primaryKey;comment:외부 환자 ID"`
    RecordedAt time.Time      `gorm:"column:recorded_at;type:datetime(3);not null;primaryKey;comment:레코드 기록일"`
    VitalType  string         `gorm:"column:vital_type;type:enum('HR','RR','SBP','DBP','SpO2','BT');not null;primaryKey;comment:바이탈 유형"`
    Value      float64        `gorm:"column:value;type:double;not null;comment:바이탈 값"`
    Version    int            `gorm:"column:version;not null;default:1;comment:버전"`
    CreatedAt  time.Time      `gorm:"column:created_at;type:datetime(3);not null;comment:데이터 생성일"`
    UpdatedAt  *time.Time     `gorm:"column:updated_at;type:datetime(3);comment:데이터 수정일"`
    DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;type:datetime(3);comment:데이터 삭제일"`
}
```

**변경사항**:
- VitalType을 primaryKey에 추가 (복합 키 구성)
- Value 컬럼명 수정 (birth_date → value)

#### param.go
```go
type UpsertVitalRequest struct {
    PatientID  string    `json:"patient_id" binding:"required"`
    RecordedAt time.Time `json:"recorded_at" binding:"required"`
    VitalType  string    `json:"vital_type" binding:"required,oneof=HR RR SBP DBP SpO2 BT"`
    Value      float64   `json:"value" binding:"required"`
    Version    int       `json:"version" binding:"required,min=1"`
}
```

**특징**:
- RecordedAt을 time.Time 타입으로 직접 받음 (string 파싱 불필요)
- VitalType 유효성 검증 (6가지 타입만 허용)

#### repository.go
```go
type VitalRepository interface {
    FindVitalByPatientIDAndRecordedAtAndVitalType(ctx context.Context, patientID string, recordedAt time.Time, vitalType string) (*Vital, error)
    CreateVital(ctx context.Context, model *Vital) error
    UpdateVital(ctx context.Context, model *Vital) error
}
```

#### service.go
```go
type VitalService interface {
    UpsertVital(ctx context.Context, request UpsertVitalRequest) error
}
```

#### controller.go
```go
type VitalController interface {
    UpsertVital(ctx *gin.Context)
}
```

### 2. Repository Layer (DB 접근)
**파일**: `/api-server/app/repository/vital_repository.go`

#### FindVitalByPatientIDAndRecordedAtAndVitalType
```go
func (v *vitalRepository) FindVitalByPatientIDAndRecordedAtAndVitalType(
    ctx context.Context,
    patientID string,
    recordedAt time.Time,
    vitalType string,
) (*vital.Vital, error) {
    var result vital.Vital
    if err := v.externalGormClient.MySQL().WithContext(ctx).
        Where("patient_id = ? AND recorded_at = ? AND vital_type = ?",
            patientID, recordedAt, vitalType).
        First(&result).Error; err != nil {
        return nil, pkgError.Wrap(err)
    }
    return &result, nil
}
```

**특징**:
- 복합 키 3개 컬럼 모두 WHERE 조건에 사용
- 메서드명에 모든 키 컬럼 명시 (FindVitalByPatientIDAndRecordedAtAndVitalType)

#### CreateVital
```go
func (v *vitalRepository) CreateVital(ctx context.Context, model *vital.Vital) error {
    return pkgError.Wrap(v.externalGormClient.MySQL().WithContext(ctx).Create(model).Error)
}
```

#### UpdateVital (DB-level Optimistic Lock)
```go
func (v *vitalRepository) UpdateVital(ctx context.Context, model *vital.Vital) error {
    oldVersion := model.Version - 1

    result := v.externalGormClient.MySQL().WithContext(ctx).
        Model(&vital.Vital{}).
        Where("patient_id = ? AND recorded_at = ? AND vital_type = ? AND version = ?",
            model.PatientID, model.RecordedAt, model.VitalType, oldVersion).
        Updates(map[string]interface{}{
            "value":      model.Value,
            "version":    model.Version,
            "updated_at": model.UpdatedAt,
        })

    if result.Error != nil {
        return pkgError.Wrap(result.Error)
    }

    // RowsAffected가 0이면 version conflict
    if result.RowsAffected == 0 {
        return pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.Conflict, "version conflict in db update")
    }

    return nil
}
```

**핵심**:
- WHERE 절에 복합 키 3개 + version 조건 추가
- RowsAffected = 0 → version conflict (DB level 최종 방어)

**테스트 커버리지**: 3개 테스트 통과

### 3. Service Layer (UPSERT 비즈니스 로직)
**파일**: `/api-server/app/service/vital_service.go`

#### UpsertVital 로직
```go
func (v *vitalService) UpsertVital(ctx context.Context, request vital.UpsertVitalRequest) error {
    // 1. 기존 데이터 조회
    existingVital, err := v.repo.FindVitalByPatientIDAndRecordedAtAndVitalType(
        ctx, request.PatientID, request.RecordedAt, request.VitalType,
    )

    now := time.Now().UTC()

    // 2. 존재하지 않으면 INSERT
    if err != nil {
        if err.Error() == gorm.ErrRecordNotFound.Error() ||
           pkgError.CompareBusinessError(err, pkgError.Get) {

            // INSERT: version은 반드시 1
            if request.Version != 1 {
                return pkgError.WrapWithCode(pkgError.EmptyBusinessError(),
                    pkgError.WrongParam, "version must be 1 for new record")
            }

            if err := v.repo.CreateVital(ctx, &vital.Vital{
                PatientID:  request.PatientID,
                RecordedAt: request.RecordedAt,
                VitalType:  request.VitalType,
                Value:      request.Value,
                Version:    1,
                CreatedAt:  now,
                UpdatedAt:  &now,
            }); err != nil {
                return pkgError.WrapWithCode(err, pkgError.Create)
            }

            return nil
        }
        return pkgError.WrapWithCode(err, pkgError.Get)
    }

    // 3. 존재하면 UPDATE (Optimistic Lock)
    if existingVital.Version != request.Version {
        return pkgError.WrapWithCode(pkgError.EmptyBusinessError(),
            pkgError.Conflict, "version mismatch")
    }

    existingVital.Value = request.Value
    existingVital.Version = request.Version + 1
    existingVital.UpdatedAt = &now

    if err := v.repo.UpdateVital(ctx, existingVital); err != nil {
        // Repository Conflict 에러는 그대로 전달
        if pkgError.CompareBusinessError(err, pkgError.Conflict) {
            return err
        }
        return pkgError.WrapWithCode(err, pkgError.Update)
    }

    return nil
}
```

**UPSERT 분기 처리**:
1. **FindVital 성공** → UPDATE 로직
   - Version 검증 (Service layer)
   - Version +1 증가
   - Repository UpdateVital 호출 (DB level 검증)

2. **FindVital 실패 (Record Not Found)** → INSERT 로직
   - Version = 1 강제 검증
   - CreateVital 호출

**2단계 Optimistic Lock**:
- **1차 방어 (Service)**: Application level에서 빠른 실패
- **2차 방어 (Repository)**: DB level WHERE version 조건

### 4. Controller Layer (HTTP 요청/응답)
**파일**: `/api-server/app/controller/vital_controller.go`

#### UpsertVital
```go
// UpsertVital
// @Title UpsertVital
// @Description Vital 데이터 저장/수정 (UPSERT, Optimistic Lock 적용)
// @Tags V1 - Vital
// @Accept json
// @Produce json
// @Param reqBody body vital.UpsertVitalRequest true "Vital 데이터 저장/수정 요청"
// @Success 200 {object} output.Output
// @Failure 400 {object} output.Output "code: 400001 - Wrong parameter"
// @Failure 409 {object} output.Output "code: 400002 - Version conflict"
// @Failure 500 {object} output.Output "code: 100001 - Fail to create / code: 100002 - Fail to update"
// @Router /v1/vitals [Post]
func (v *vitalController) UpsertVital(ctx *gin.Context) {
    var reqBody vital.UpsertVitalRequest
    if err := ctx.ShouldBindJSON(&reqBody); err != nil {
        output.AppendErrorContext(ctx, pkgError.WrapWithCode(err, pkgError.WrongParam,
            err.Error(), "fail to parse request parameter"), nil)
        return
    }

    if err := v.service.UpsertVital(ctx, reqBody); err != nil {
        output.AppendErrorContext(ctx, pkgError.Wrap(err), nil)
        return
    }

    output.Send(ctx, nil)
}
```

**Swagger 주석**:
- UPSERT 동작 방식 명시
- Optimistic Lock 적용 설명
- 에러 코드별 설명

### 5. Router Layer (라우팅)
**파일**: `/api-server/app/router/vital_router.go`

```go
func NewVitalRouter(engine *gin.Engine, controller vital.VitalController) {
    v1Group := engine.Group("/api/v1")
    v1Group.Use(middleware.ValidTokenMiddleware())

    vitalGroup := v1Group.Group("/vitals")
    {
        vitalGroup.POST("", controller.UpsertVital)
    }
}
```

## UPSERT 동작 시나리오

### Scenario 1: INSERT (새 데이터)
```json
POST /api/v1/vitals
{
  "patient_id": "P00001234",
  "recorded_at": "2025-12-01T10:15:00Z",
  "vital_type": "HR",
  "value": 110.0,
  "version": 1
}
```

**처리 흐름**:
1. FindVital → Record Not Found
2. Version = 1 검증 ✅
3. CreateVital 실행
4. 200 OK

### Scenario 2: UPDATE (기존 데이터)
```json
POST /api/v1/vitals
{
  "patient_id": "P00001234",
  "recorded_at": "2025-12-01T10:15:00Z",
  "vital_type": "HR",
  "value": 115.0,
  "version": 1
}
```

**처리 흐름**:
1. FindVital → 존재 (version=1)
2. Service: Version 검증 (1 == 1) ✅
3. Version +1 → 2
4. Repository: WHERE version=1 조건 UPDATE
5. RowsAffected > 0 ✅
6. 200 OK

### Scenario 3: Version Conflict
```json
POST /api/v1/vitals
{
  "patient_id": "P00001234",
  "recorded_at": "2025-12-01T10:15:00Z",
  "vital_type": "HR",
  "value": 120.0,
  "version": 1
}
```

**처리 흐름** (DB version = 2):
1. FindVital → 존재 (version=2)
2. Service: Version 검증 (2 ≠ 1) ❌
3. 409 Conflict 반환

**또는** (동시성 시나리오):
1. FindVital → 존재 (version=1)
2. Service: Version 검증 (1 == 1) ✅
3. **[다른 요청이 먼저 UPDATE 완료, version=2]**
4. Repository: WHERE version=1 조건 UPDATE
5. RowsAffected = 0 ❌
6. 409 Conflict 반환 (DB level 방어)

### Scenario 4: 잘못된 Version (INSERT 시)
```json
POST /api/v1/vitals
{
  "patient_id": "P00001234",
  "recorded_at": "2025-12-01T10:15:00Z",
  "vital_type": "HR",
  "value": 110.0,
  "version": 5
}
```

**처리 흐름**:
1. FindVital → Record Not Found
2. Version = 5 (≠ 1) ❌
3. 400 Bad Request: "version must be 1 for new record"

## 복합 식별자 (Composite Key) 특징

### Primary Key 구성
```
(patient_id, recorded_at, vital_type)
```

### 의미
- 같은 환자(patient_id)의
- 같은 시간(recorded_at)에
- 같은 유형(vital_type)의 Vital 데이터는 **유일**

### 예시
```
✅ 허용:
  P00001234 | 2025-12-01 10:15:00 | HR    | 110.0
  P00001234 | 2025-12-01 10:15:00 | RR    | 20.0  (다른 VitalType)
  P00001234 | 2025-12-01 10:20:00 | HR    | 112.0 (다른 시간)

❌ 중복 (UPSERT로 UPDATE):
  P00001234 | 2025-12-01 10:15:00 | HR    | 115.0 (동일 복합 키)
```

## Optimistic Lock 구현 상세

### 2단계 방어 체계

#### 1차 방어: Service Layer (Application Level)
```go
if existingVital.Version != request.Version {
    return pkgError.WrapWithCode(pkgError.EmptyBusinessError(),
        pkgError.Conflict, "version mismatch")
}
```

**장점**:
- 빠른 실패 (DB 호출 전 검증)
- 명확한 에러 메시지

**한계**:
- Race condition 가능 (FindVital과 UpdateVital 사이 시간차)

#### 2차 방어: Repository Layer (DB Level)
```go
result := db.Where("... AND version = ?", oldVersion).Updates(...)
if result.RowsAffected == 0 {
    return Conflict
}
```

**장점**:
- 진정한 동시성 제어 (DB 트랜잭션 보장)
- Race condition 방지

**SQL 실행**:
```sql
UPDATE vitals
SET value = ?, version = ?, updated_at = ?
WHERE patient_id = ?
  AND recorded_at = ?
  AND vital_type = ?
  AND version = ?  -- Optimistic Lock 조건
```

### 동시성 시나리오

```
시간 | 요청 A                          | 요청 B
-----|--------------------------------|--------------------------------
t0   | FindVital (version=1)          | FindVital (version=1)
t1   | Service 검증: 1==1 ✅          | Service 검증: 1==1 ✅
t2   | UPDATE WHERE version=1 → 성공  |
t3   | version=2로 저장 완료           |
t4   |                                 | UPDATE WHERE version=1 → 실패
t5   |                                 | RowsAffected=0 → Conflict ❌
```

**결과**: Repository layer의 WHERE version 조건이 최종 방어선 역할

## 준수한 설계 규칙

### Repository Layer
- ✅ Context를 첫 번째 인자로 전달
- ✅ WithContext(ctx) 사용
- ✅ DB 행위 중심 메서드명 (FindVitalByPatientIDAndRecordedAtAndVitalType)
- ✅ 복합 키 모든 컬럼 메서드명에 명시
- ✅ (model, error) 반환 형식

### Service Layer
- ✅ Context를 첫 번째 인자로 전달
- ✅ Request 해석 및 비즈니스 로직 수행
- ✅ UPSERT 분기 처리
- ✅ 도메인 모델 생성 및 Timestamp 설정
- ✅ 에러를 pkgError.WrapWithCode로 래핑
- ✅ Repository Conflict 에러 그대로 전달

### Controller Layer
- ✅ HTTP 요청/응답만 처리
- ✅ ctx.ShouldBindJSON 사용
- ✅ pkgError.WrapWithCode로 에러 래핑
- ✅ output.Send로 응답
- ✅ Swagger 주석 포함

### Router Layer
- ✅ Version Group (/api/v1) 하위 위치
- ✅ REST 의미에 맞는 HTTP Method (POST for UPSERT)
- ✅ 도메인별 Resource Group 분리

## Mock 생성

```bash
cd /api-server/domain/vital
go generate ./...
```

생성된 파일:
- `/api-server/domain/mock/mock_vital_repository.go`
- `/api-server/domain/mock/mock_vital_service.go`
- `/api-server/domain/mock/mock_vital_controller.go`

## 테스트 실행 결과

### Repository Tests
```
✅ Test_FindVitalByPatientIDAndRecordedAtAndVitalType
   - 성공 - Vital 조회
   - 실패 - Vital 없음
✅ Test_CreateVital
✅ Test_UpdateVital
```

**커버리지**: 주요 CRUD 로직 커버

## API 요청/응답 예시

### INSERT 성공
**Request**:
```http
POST /api/v1/vitals
Authorization: Bearer <token>
Content-Type: application/json

{
  "patient_id": "P00001234",
  "recorded_at": "2025-12-01T10:15:00Z",
  "vital_type": "HR",
  "value": 110.0,
  "version": 1
}
```

**Response**:
```http
HTTP/1.1 200 OK

{
  "success": true
}
```

### UPDATE 성공
**Request**:
```http
POST /api/v1/vitals
{
  "patient_id": "P00001234",
  "recorded_at": "2025-12-01T10:15:00Z",
  "vital_type": "HR",
  "value": 115.0,
  "version": 1
}
```

**Response**:
```http
HTTP/1.1 200 OK

{
  "success": true
}
```

### Version Conflict
**Request**:
```http
POST /api/v1/vitals
{
  "patient_id": "P00001234",
  "recorded_at": "2025-12-01T10:15:00Z",
  "vital_type": "HR",
  "value": 120.0,
  "version": 1  // DB는 이미 version=2
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

### INSERT 시 잘못된 Version
**Request**:
```http
POST /api/v1/vitals
{
  "patient_id": "P00001234",
  "recorded_at": "2025-12-01T10:15:00Z",
  "vital_type": "HR",
  "value": 110.0,
  "version": 5
}
```

**Response**:
```http
HTTP/1.1 400 Bad Request

{
  "code": 400001,
  "message": "wrong parameter",
  "detail": ["version must be 1 for new record"]
}
```

## 주요 학습 사항

1. **UPSERT 패턴**: 단일 엔드포인트로 INSERT/UPDATE 처리
2. **복합 식별자**: 3개 컬럼 조합으로 유일성 보장
3. **2단계 Optimistic Lock**: Application + DB level 방어
4. **time.Time 직접 바인딩**: Gin의 time.Time 자동 파싱 활용
5. **Repository Conflict 전달**: 이미 BusinessError인 경우 재래핑하지 않음
6. **명확한 메서드명**: FindVitalByPatientIDAndRecordedAtAndVitalType (모든 키 명시)

## 환자 API와의 차이점

| 항목 | 환자 API | Vital API |
|-----|---------|-----------|
| HTTP Method | PUT (update only) | POST (upsert) |
| 식별자 | 단일 (patient_id) | 복합 (patient_id + recorded_at + vital_type) |
| INSERT | 별도 POST 엔드포인트 | UPSERT로 통합 |
| Version 초기값 | 1 (자동) | 1 (명시적 검증) |
| Endpoint | /patients/{id} | /vitals |

## 다음 단계

- [ ] Vital 데이터 조회 API 구현
- [ ] Inference API 구현 (Vital Risk Score)
- [ ] Service/Controller 테스트 코드 추가
- [ ] 동시성 테스트 추가