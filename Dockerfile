# 1단계: 빌드 스테이지
FROM library/golang:1.25.5-alpine AS builder
# 빌드에 필요한 도구 설치
RUN apk add --no-cache git
WORKDIR /app
# 1. 로컬 라이브러리 및 서버 코드 복사
COPY library/ ./library/
COPY api-server/ ./api-server/
# 2. swag 설치 및 실행
RUN go install github.com/swaggo/swag/cmd/swag@v1.8.7
# api-server 디렉토리로 이동
WORKDIR /app/api-server
# swagger 문서 생성 (api-server 루트에서 cmd/main.go를 참조)
RUN swag init -g cmd/main.go -d .
# 3. 의존성 해결
RUN go mod tidy
RUN go mod download
# 4. 바이너리 빌드 (중요: 빌드 대상을 ./cmd 로 변경)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -tags timetzdata -a \
    -ldflags '-w -s' \
    -o /app/aitrics-vital-signs ./cmd
# 2단계: 실행 스테이지 (최소 크기)
FROM scratch
# 빌드된 바이너리 복사
COPY --from=builder /app/aitrics-vital-signs /aitrics-vital-signs
EXPOSE 8080
# 실행
ENTRYPOINT ["/aitrics-vital-signs"]