package clio

// Court는 Clio가 사건에 연결할 때 필요한 법원 정보만 담은 복사본입니다.
// 전체 기관 정보는 Leviathan이 소유하고, 이 구조체는 시작할 때 받은 목록을 검증에만 씁니다.
//
// Clio는 Leviathan의 Go 패키지를 직접 import하지 않으므로, gRPC로 받은 법원 정보를
// 이 구조체로 옮겨 담아 씁니다(옮겨 담는 코드는 cmd/clio/main.go의 courtsFromProto).
type Court struct {
	ID         string       // 법원 ID(예: "court-district-seoul-central")
	Name       string       // 법원 이름(예: "서울중앙지방법원")
	Seat       string       // 소재지(예: "서울특별시")
	PostalCode string       // 우편번호
	Address    string       // 도로명 주소
	Level      string       // 법원 단계(supreme, high, district, local)
	CourtType  string       // 법원 분야(general, family, administrative, patent, rehabilitation)
	UnitType   string       // 본원·지원·시·군법원 구분(main, support, municipal_county)
	Panels     []CourtPanel // 소속 재판부 목록. 사건의 재판부가 이 안에 있는지 검사합니다.
}

// CourtPanel은 법원에 속한 재판부의 ID와 이름입니다.
type CourtPanel struct {
	ID   string
	Name string
}
