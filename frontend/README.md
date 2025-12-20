# Alerts Dashboard

React 18 + TypeScript + Vite + Tailwind CSS 프로젝트입니다.

## 실행 방법

### 1. Node.js 설치 확인

먼저 Node.js가 설치되어 있는지 확인하세요:

```bash
node --version
npm --version
```

Node.js가 없다면 [nodejs.org](https://nodejs.org/)에서 설치하세요 (LTS 버전 권장).

### 2. 의존성 설치

프로젝트 루트 디렉토리에서 다음 명령어를 실행하세요:

```bash
npm install
```

또는 yarn을 사용하는 경우:

```bash
yarn install
```

### 3. 개발 서버 실행

의존성 설치가 완료되면 개발 서버를 시작하세요:

```bash
npm run dev
```

또는 yarn을 사용하는 경우:

```bash
yarn dev
```

### 4. 브라우저에서 확인

터미널에 표시된 URL (보통 `http://localhost:5173`)로 브라우저에서 접속하세요.

## 환경 변수 설정

백엔드 API URL을 설정하려면 프로젝트 루트에 `.env` 파일을 생성하세요:

```bash
# .env 파일
VITE_API_BASE_URL=http://kube-rca-backend:8080
```

- **개발 환경**: 로컬 개발 시 `http://localhost:8080` 사용 가능
- **프로덕션 환경 (쿠버네티스)**: `http://kube-rca-backend:8080` 사용 (기본값)
- 환경 변수가 설정되지 않으면 기본값 `http://kube-rca-backend:8080`이 사용됩니다.

## 백엔드 API 연동

이 프로젝트는 백엔드 API와 연동됩니다:

- **API 엔드포인트**: `/api/alerts`
- **기본 URL**: `http://kube-rca-backend:8080` (쿠버네티스 클러스터 내부 통신)
- **응답 형식**: `AlertItem[]` 배열 또는 `{ alerts: AlertItem[] }` 객체

백엔드 API가 다른 형식으로 응답하는 경우 `src/utils/api.ts` 파일을 수정하세요.

## Docker를 사용한 실행

### 1. Docker 이미지 빌드

```bash
docker build -t alerts-dashboard .
```

### 2. Docker 컨테이너 실행

```bash
docker run -d -p 80:80 alerts-dashboard
```

### 3. 브라우저에서 확인

`http://localhost:8080`으로 접속하세요.

## 사용 가능한 명령어

- `npm run dev` - 개발 서버 시작
- `npm run build` - 프로덕션 빌드
- `npm run preview` - 빌드된 앱 미리보기
- `npm run lint` - 코드 린팅

## 프로젝트 구조

```
src/
├── main.tsx              # React 진입점
├── App.tsx               # 메인 앱 컴포넌트
├── types.ts              # TypeScript 타입 정의
├── index.css             # Tailwind CSS
├── components/           # React 컴포넌트
│   ├── TimeRangeSelector.tsx  # 시간 범위 선택기
│   ├── AlertsTable.tsx        # 알림 테이블
│   └── Pagination.tsx         # 페이지네이션
├── constants/            # 상수 정의
│   └── index.ts          # 앱 전역 상수
└── utils/                # 유틸리티 함수
    ├── api.ts            # 백엔드 API 클라이언트
    ├── mockData.ts       # Mock 데이터 생성 함수
    └── filterAlerts.ts  # 알림 필터링 유틸리티
```

## 기술 스택

- **React 18** - UI 라이브러리
- **TypeScript** - 타입 안정성
- **Vite** - 빌드 도구
- **Tailwind CSS** - 스타일링
- **ESLint** - 코드 품질 관리
- **Docker** - 컨테이너화
- **Nginx** - 프로덕션 웹 서버
