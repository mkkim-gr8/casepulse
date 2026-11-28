# Aegis 작업 내역과 코드 안내

Aegis는 브라우저가 접속하는 웹 서비스입니다. Go(Gin) HTTP API와 React 화면(`web/`)을 함께 제공하며, 데이터는 직접 저장하지 않고 Clio·Leviathan에 gRPC로 요청합니다. 실행·API 설명은 [README.md](README.md)를 참고하고, 이 문서는 **지금까지 한 작업**과 **처음 코드를 읽는 분을 위한 안내**를 정리합니다.

## 지금까지 한 작업

| 날짜 | 커밋 / PR | 내용 |
|---|---|---|
| 2026-11-27 | `8749ea7` / PR #4 | 사건목록·사건등록 웹 서비스를 처음 만들었습니다(당시 이름 `cmd/casepulse-web`). API `GET /api/cases`, `GET /api/cases/{id}`, `POST /api/cases`, `GET /api/courts`와 React + shadcn/ui(Tailwind CSS v4) 화면을 추가했습니다. 통합 검색, 필터, 정렬, 무한 스크롤, 의뢰인/다른 당사자 표시, 상세 패널, 신건등록 탭(사건번호 붙여넣기 분할, 이해충돌 즉시 경고, 법원·재판부 검색 선택, `Ctrl+Enter` 등록, 저장 안 된 탭 닫기 확인), 밝은/어두운/시스템 테마, 로고·파비콘을 넣었습니다. 화면 빌드 결과는 `go:embed`로 실행 파일에 포함합니다. |
| 2026-11-27 | `c8b4362` / PR #5 | 화면을 다듬었습니다. 사건목록 왼쪽 독바(필터·신건등록), 사건 보기 방식(한 번 클릭 = 미리보기, 더블클릭·⌘/Ctrl+클릭·"탭으로 열기" = 사건 탭), 사건목록 탭 안에서 열리는 미리보기 패널(Esc로 닫기), 상세 카드 분리(`CaseDetails.jsx`), 맥 ⌘ / 그 밖 Ctrl 단축키 표시, 브라우저 탭 모양 디자인을 적용했습니다. |
| 2026-11-28 | `2cad1bb` | 서비스를 나누며 이름을 **Aegis**(`cmd/aegis`)로 바꾸고, 화면 소스를 저장소 최상위 `web/`으로 옮겼습니다. 사건 데이터·검증은 Clio로, 법원 목록은 Leviathan으로 넘기고 Aegis는 gRPC 클라이언트만 사용합니다. HTTP JSON ↔ gRPC 변환은 `api.go`에 있습니다. Dockerfile(Node 빌드 → Go 빌드 → 작은 실행 이미지)과 `compose.yaml`을 추가했습니다. |
| 2026-11-28 | `f85c9da` / PR #8 | `GET /healthz`와 `-healthcheck` 옵션을 추가했습니다. 시작할 때 Clio 연결을 표준 gRPC health 확인으로 검사하고, Compose는 Clio·Leviathan이 healthy가 된 뒤 Aegis를 시작합니다. |
| 2026-11-28 | `b5a45c1` / PR #10 | Clio 저장소 장애를 503으로 돌려주도록 맞췄습니다(Clio의 `Unavailable`). |
| 2026-11-28 | (이번 작업) | 초급 개발자도 읽을 수 있도록 Go·React 소스 주석을 보강하고 이 문서를 추가했습니다. |

## 요청 흐름

```text
브라우저 ──HTTP/JSON──▶ Aegis :8080
                         ├─ /api/cases*  ──gRPC──▶ Clio :9091      (사건 조회·등록)
                         ├─ /api/courts  ── 시작할 때 Leviathan :9090에서 받아 둔 법원 목록
                         ├─ /healthz     ── "ok"
                         └─ 그 밖의 경로  ── web/dist의 React 화면 파일
```

예를 들어 신건등록 화면에서 `등록`을 누르면 다음 순서로 처리됩니다.

1. `web/src/CaseForm.jsx`가 화면에서 먼저 검사하고 `api.js`의 `createCase`로 `POST /api/cases`를 보냅니다.
2. `cmd/aegis/api.go`가 JSON을 읽어 `cliov1.CaseInput`으로 바꾸고 Clio의 `CreateCase`를 호출합니다.
3. Clio가 검증·저장한 뒤 결과나 오류 코드를 돌려줍니다.
4. Aegis가 결과를 `caseView` JSON(201)으로, 오류는 400/409/503으로 바꿔 브라우저에 보냅니다.
5. 화면은 성공하면 탭을 닫고 사건목록 맨 위에 새 사건을 넣어 미리보기를 엽니다.

## 코드 읽는 순서

### Go 서버

1. **[`cmd/aegis/main.go`](main.go)** — 시작점. Leviathan·Clio 연결을 확인하고, `newHandler`로 경로를 등록한 뒤 HTTP 서버를 켭니다.
2. **[`cmd/aegis/api.go`](api.go)** — API 핸들러. JSON 자료형(`caseView` 등)과 gRPC 변환 함수가 있습니다.
3. **[`cmd/aegis/main_test.go`](main_test.go)** — 실제 Leviathan·Clio gRPC 서버를 테스트 안에서 임시 포트로 띄우고(Clio는 메모리 저장소 + 더미 사건) API를 시험합니다. 응답 형식을 확인할 때 참고하기 좋습니다.
4. **[`web/embed.go`](../../web/embed.go)** — 빌드된 화면을 실행 파일에 넣는 코드입니다.

### React 화면 (`web/src/`)

1. **`main.jsx`** — 화면 시작점(테마·툴팁·알림 제공자로 `App`을 감쌈).
2. **`App.jsx`** — 메뉴줄과 탭 관리(사건목록 / 신건등록 / 사건 탭), 데이터 불러오기, 단축키.
3. **`CaseListPage.jsx`** → **`Filters.jsx`**, **`CaseTable.jsx`**, **`CaseDrawer.jsx`** — 사건목록 탭.
4. **`CaseForm.jsx`** — 신건등록 탭.
5. **`CaseDetailPage.jsx`**, **`CaseDetails.jsx`** — 사건 탭과 상세 카드.
6. **`api.js`**(서버 호출), **`cases.js`**(검색·정렬·당사자 묶기), **`keyboard.js`**(단축키 도우미) — 화면과 관계없는 함수 모음.

`web/src/components/ui/`는 `npx shadcn@latest add <이름>`으로 받은 라이브러리 컴포넌트이므로 직접 고치기보다 그대로 사용합니다. `components/combobox.jsx`, `mode-toggle.jsx`, `theme-provider.jsx`는 이 프로젝트에서 만든 컴포넌트입니다.

## 파일별 역할

| 파일 | 역할 |
|---|---|
| `cmd/aegis/main.go` | 시작점, Gin 라우터 구성, 정적 화면 제공 |
| `cmd/aegis/api.go` | `/api/*` 핸들러, JSON ↔ gRPC 변환, 오류 코드 변환 |
| `cmd/aegis/main_test.go` | API 테스트(테스트 안에서 Leviathan·Clio gRPC 서버를 임시로 띄움) |
| `cmd/aegis/Dockerfile` | 화면 빌드 → Go 빌드 → 실행 이미지 |
| `web/embed.go` | `web/dist`를 Go 실행 파일에 포함 |
| `web/vite.config.js` | 개발 서버(`/api`를 :8080으로 넘김)와 빌드 설정 |
| `web/src/App.jsx` | 앱 뼈대, 탭 관리 |
| `web/src/CaseListPage.jsx` | 사건목록 탭(독바, 검색, 필터, 표, 미리보기) |
| `web/src/CaseTable.jsx` | 사건 표, 정렬 머리글, 무한 스크롤 |
| `web/src/Filters.jsx` | 사건 종류·접수 연도·계속 법원 필터 |
| `web/src/CaseDrawer.jsx` | 사건목록 안 오른쪽 미리보기 패널 |
| `web/src/CaseDetailPage.jsx`, `CaseDetails.jsx` | 사건 탭, 사건·당사자·법원 카드 |
| `web/src/CaseForm.jsx` | 신건등록 폼과 화면 검사 |
| `web/src/api.js`, `cases.js`, `keyboard.js` | 서버 호출, 사건 데이터 함수, 단축키 도우미 |
| `web/src/components/` | 직접 만든 컴포넌트와 shadcn/ui 컴포넌트(`ui/`) |

## 알아 두면 좋은 개념

- **관문(gateway) 서비스**: Aegis는 데이터를 갖지 않고 요청을 알맞은 서비스로 넘깁니다. 그래서 Clio 내부(DB 등)가 바뀌어도 Aegis는 gRPC 계약만 지키면 됩니다.
- **화면 검사와 서버 검사**: 화면(`CaseForm.jsx`의 `validate`)은 빠른 안내용이고, 최종 판단은 Clio가 합니다. 두 규칙은 같게 유지합니다(예: `cases.js`의 `partySide`와 Clio의 `PartySide`).
- **React 상태**: 화면에 보이는 값은 `useState`로 보관하고 set 함수로만 바꿉니다. 배열·객체는 직접 고치지 않고 새로 만들어 넘깁니다.
- **탭을 숨기기만 하기**: 다른 탭으로 옮겨도 탭 내용을 지우지 않고 `hidden`으로 숨겨서, 입력 중인 내용과 스크롤 위치가 남습니다.
- **법원 목록 스냅샷**: Aegis도 시작할 때 받은 법원 목록을 씁니다. Leviathan 데이터를 바꾸면 Aegis를 재시작해야 등록 화면에 반영됩니다.

## 남은 과제

- 사건 수정 화면, 사건 탭의 추가 기능(진행내역 등은 Clio 기능이 생긴 뒤)
- 목록이 커질 때 서버 쪽 페이지 나누기·검색
- 로그인·권한 관리
- 메뉴줄의 "더미 데이터" 표시는 실제 데이터 운영 전에 정리해야 합니다.
