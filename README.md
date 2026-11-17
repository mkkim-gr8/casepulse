# casepulse
golang based micro service architecture for legal office and its letigation case management 

casepulse 는 법무법인, 변호사사무소에서의 ERP 시스템에 필요한 자료구조를 정의한다. 법원, 공소청(검찰), 경찰, 등기소 정보를 관리한다.

본 프로젝트 casepulse는 Micro Service로 이용한다. 독립한 하나의 서비스임. 외부에서 gRPC를 통해 각 기관정보를 요청하면 그 응답을 돌려준다.

## gRPC API

기관 정보 조회 API 계약은 [`api/proto/casepulse/v1/institution.proto`](api/proto/casepulse/v1/institution.proto)에 정의한다. 기관 ID로 단건 조회하고, 기관 종류와 페이지 토큰으로 목록을 조회할 수 있다.

법원 기관은 법원 단계, 관할 분야, 본원·지원·시·군법원 구분, 상위 법원, 소재지와 관할 지역을 별도 상세정보로 표현한다. [`Courts`](courts.go)는 [`documents/pdfs`](documents/pdfs)에 둔 법원 명칭·소재지 및 관할구역 별표를 바탕으로 221개 법원 목록을 제공한다. `Seat`은 시·군 단위의 소재지를 담으며 도로명 주소는 포함하지 않는다. 관할구역 별표의 개정 시점은 자료별로 다르므로 최신 법령과 대조해 갱신해야 한다.

[`documents/pdfs/재판부구성표.pdf`](documents/pdfs/재판부구성표.pdf)에 나온 재판부는 소속 `Court`의 하위 `Panels`에 둔다. 합의부·단독 재판부와 조정·재정·회생·파산 등 표에 기재된 재판 단위를 재판부명으로 기록한다. 원외재판부는 별도의 고등법원 `Court`로 표현하고, 예를 들어 서울고등법원 인천 원외재판부는 `서울고등법원(인천)`이라는 이름의 고등법원으로 등록한다. 원외 고등법원은 본원 고등법원과 동급이므로 둘 다 대법원을 상위 기관으로 가리키며, 인천 재판부들은 `서울고등법원(인천)`의 `Panels`에 직접 둔다. 판사 개인 명단과 대괄호 안의 사건 분야 안내는 포함하지 않는다. 구성표 개정 시 이 데이터도 함께 갱신해야 한다.

Buf를 설치한 뒤 아래 명령으로 API 정의를 검사하고 Go 코드를 생성한다.

```sh
buf lint
buf generate
```

현재는 프로토콜 정의와 코드 생성 설정까지 준비되어 있다. 실제 gRPC 서버와 저장소 연결은 별도로 구현해야 한다.
