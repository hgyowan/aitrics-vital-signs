## **🧩 AITRICS Backend Senior Engineer 과제 전형**

본 과제는 AITRICS의 실제 Backend 개발 환경(온프레미스, 의료 데이터, ML inference 등)을 단순화한 형태의 과제로서,**시니어 엔지니어의 설계 역량 / 코드 품질 / 동시성 제어 능력 / 테스트 / 문서화 능력 / AI 도구 활용 능력**을 평가하기 위해 설계되었습니다.

## **1. 도메인 설명 – Vital 데이터란?**

병원에서는 환자의 상태 변화를 모니터링하기 위해 다양한 **Vital Signs**(생체징후) 데이터를 실시간 수집합니다.

본 과제에서는 아래 6개의 Vital 유형을 사용합니다.

| **Vital Type** | **설명** | **단위** |
| --- | --- | --- |
| HR | 심박수 (Heart Rate) | bpm |
| RR | 호흡수 (Respiratory Rate) | breaths/min |
| SBP | 수축기 혈압 | mmHg |
| DBP | 이완기 혈압 | mmHg |
| SpO2 | 산소포화도 | % |
| BT | 체온 | ℃ |

Vital 데이터는 시간에 따라 수집되는 **시계열(Time-series) 데이터**이며, 의료 AI 모델의 입력으로 사용됩니다.

## **2. 기능 요구사항**

아래 API 및 기능을 구현해야 합니다.

### **2-1. 환자 관리 API**

### **(1) 환자 등록 API**

**Method**: POST

**Endpoint**: /api/v1/patients

**Request 예시**

{
"patient_id": "P00001234",
"name": "홍길동",
"gender": "M",
"birth_date": "1975-03-01"
}

### **(2) 환자 정보 수정 API (Optimistic Lock 必 적용)**

⚠ 반드시 요청과 DB의 version 비교를 통한 낙관적 락 적용

**Method**: PUT

**Endpoint**: /api/v1/patients/{patient_id}

**제약사항**

Request Body에 version 필수

DB version과 다르면 → 409 Conflict 반환

**Request 예시**

{
"name": "홍길동",
"gender": "M",
"birth_date": "1975-03-01",
"version": 3
}

### **2-2. Vital 데이터 API**

### **(1) Vital 데이터 저장/수정 API (UPSERT, Optimistic Lock 必 적용)**

⚠ patient_id, recorded_at, vital_type를 복합 식별자로 사용하여 UPSERT 구현⚠ 업데이트 시 반드시 version 비교를 통한 낙관적 락 적용

**Method**: POST

**Endpoint**: /api/v1/vitals

**동작 방식**

동일한 (patient_id, recorded_at, vital_type) 조합이 존재하지 않으면 → **INSERT**

동일한 조합이 이미 존재하면 → **UPDATE** (이때 version 검증 필수)

**Request 예시**

{
"patient_id": "P00001234",
"recorded_at": "2025-12-01T10:15:00Z",
"vital_type": "HR",
"value": 110.0,
"version": 1
}

**제약사항**

등록된 환자만 입력 가능

vital_type ∈ ["HR", "RR", "SBP", "DBP", "SpO2", "BT"]

version 필수

UPDATE 시 DB의 version과 다르면 → 409 Conflict 반환

INSERT 시 version은 1부터 시작

### **(2) Vital 데이터 조회 API**

**Method**: GET

**Endpoint**: /api/v1/patients/{patient_id}/vitals

**Query Parameters**

from (필수)

to (필수)

vital_type (선택)

**Response 예시**

{
"patient_id": "P00001234",
"vital_type": "HR",
"items": [
{
"recorded_at": "2025-12-01T10:15:00Z",
"value": 110.0
}
]
}

### **2-3. Inference API (단순 Rule 기반 위험 스코어)**

### **(1) Inference 요청 API**

**Method**: POST

**Endpoint**: /api/v1/inference/vital-risk

**기능 설명**

DB에 저장된 환자의 Vital 데이터를 조회하여 위험 스코어를 계산합니다.

환경변수로 설정된 시간 범위(예: 24시간) 내의 최근 데이터를 사용합니다.

조회된 모든 시점의 Vital 데이터를 분석하여 종합 스코어를 계산합니다.

**Request 예시**

{
"patient_id": "P00001234"
}

**Request 필드**

patient_id (필수): 환자 ID

**환경변수 설정**

VITAL_RISK_TIME_WINDOW_HOURS: 스코어 계산 시 조회할 최근 데이터의 시간 범위 (기본값: 24)

### **평가 규칙**

**위험 조건**

| **조건** | **의미** |
| --- | --- |
| HR > 120 | 위험 증가 |
| SBP < 90 | 위험 증가 |
| SpO2 < 90 | 위험 증가 |

**스코어링 로직**

DB에서 환경변수로 설정된 시간 범위(VITAL_RISK_TIME_WINDOW_HOURS) 내의 모든 Vital 데이터를 조회합니다.

각 Vital Type(HR, SBP, SpO2)별로 평균값을 계산합니다.

계산된 평균값에 대해 위 3가지 위험 조건을 평가합니다.

충족된 조건 개수에 따라 risk_level을 결정합니다.

충족 조건 개수에 따른 위험 등급:

| **충족 개수** | **risk_level** |
| --- | --- |
| 0 | LOW |
| 1–2 | MEDIUM |
| ≥3 | HIGH |

**Response 예시**

{
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

**필드 설명**

risk_level: LOW / MEDIUM / HIGH 중 하나

triggered_rules: 충족된 위험 조건들

vital_averages: 분석 기간 동안의 각 Vital Type별 평균값

data_points_analyzed: 분석에 사용된 총 Vital 데이터 레코드 수

time_range: 실제로 분석한 시간 범위

## **3. 인증 요구사항**

모든 API는 **Bearer Token 기반 인증** 적용

인증 실패 시 → 401 Unauthorized

**Header 예시**

Authorization: Bearer <token>

토큰은 환경 변수 또는 설정 파일로 관리

## **4. 비기능 요구사항**

| **항목** | **설명** |
| --- | --- |
| 아키텍처 | Layered 또는 DDD-lite 구조 권장 |
| 테스트 | 높은 유닛 테스트 커버리지 |
| Dockerfile | 로컬 실행 가능하도록 제공 |
| Swagger/OpenAPI | API 문서 포함 |
| Config | 환경파일(.env 등)로 분리하여 온프레미스 대응 |

## **5. Optimistic Lock 필수 적용 지점 (2곳)**

==다음 두 API는== **==반드시 version 기반 낙관적 락 적용이 필수입니다==.**

### **(A) 환자 수정 API**

PUT /api/v1/patients/{patient_id}

version mismatch → 409 Conflict

### **(B) Vital 저장/수정 API (UPDATE 시)**

POST /api/v1/vitals

UPSERT 방식으로 동작하며, 기존 데이터 UPDATE 시 version mismatch → 409 Conflict

복합 식별자: (patient_id, recorded_at, vital_type)