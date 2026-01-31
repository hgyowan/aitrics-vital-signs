-- aitrics_db.patients definition

CREATE TABLE `patients` (
                            `id` char(36) NOT NULL COMMENT 'PK',
                            `patient_id` varchar(20) NOT NULL COMMENT '외부 환자 ID',
                            `name` varchar(50) NOT NULL COMMENT '환자 이름',
                            `gender` enum('M','F') NOT NULL COMMENT '성별',
                            `birth_date` date NOT NULL COMMENT '생년월일',
                            `version` bigint NOT NULL DEFAULT '1' COMMENT '버전',
                            `created_at` datetime(3) NOT NULL COMMENT '데이터 생성일',
                            `updated_at` datetime(3) DEFAULT NULL COMMENT '데이터 수정일',
                            `deleted_at` datetime(3) DEFAULT NULL COMMENT '데이터 삭제일',
                            PRIMARY KEY (`id`),
                            UNIQUE KEY `idx_patients_patient_id` (`patient_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- aitrics_db.vitals definition

CREATE TABLE `vitals` (
                          `patient_id` varchar(20) NOT NULL COMMENT '외부 환자 ID',
                          `recorded_at` datetime(3) NOT NULL COMMENT '레코드 기록일',
                          `vital_type` enum('HR','RR','SBP','DBP','SpO2','BT') NOT NULL COMMENT '바이탈 유형',
                          `value` double NOT NULL COMMENT '바이탈 값',
                          `version` bigint NOT NULL DEFAULT '1' COMMENT '버전',
                          `created_at` datetime(3) NOT NULL COMMENT '데이터 생성일',
                          `updated_at` datetime(3) DEFAULT NULL COMMENT '데이터 수정일',
                          `deleted_at` datetime(3) DEFAULT NULL COMMENT '데이터 삭제일',
                          PRIMARY KEY (`patient_id`,`recorded_at`,`vital_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;