# backend

## Basic Gin API Server

이 프로젝트는 Go 언어와 Gin 프레임워크를 사용한 **아주 기초적인 REST API 예제**입니다.

### 1. 환경 준비

- **Go 설치**: Go 1.22 이상이 설치되어 있어야 합니다.
- **모듈 의존성 설치** (처음 한 번만 실행):

```bash
cd /Users/bohyunchoi/Github/backend
go mod tidy
```

### 2. 서버 실행

```bash
cd /Users/bohyunchoi/Github/backend
go run main.go
```

기본적으로 `http://localhost:8080` 에서 서버가 실행됩니다.

### 3. API 테스트

- **루트 엔드포인트**

```bash
curl http://localhost:8080/
```

예상 응답:

```json
{
  "status": "ok",
  "message": "Gin basic API server is running"
}
```

- **테스트 엔드포인트 (`/ping`)**

```bash
curl http://localhost:8080/ping
```

예상 응답:

```json
{
  "message": "pong"
}
```

이 정도 구성이면 Gin을 이용해 기본적인 API 요청/응답 흐름을 테스트하기에 충분합니다.
