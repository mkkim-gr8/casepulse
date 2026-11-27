package clio

// Court는 Clio가 사건에 연결할 때 필요한 법원 정보만 담은 복사본입니다.
// 전체 기관 정보는 Leviathan이 소유하고, 이 구조체는 시작할 때 받은 목록을 검증에만 씁니다.
type Court struct {
	ID         string
	Name       string
	Seat       string
	PostalCode string
	Address    string
	Level      string
	CourtType  string
	UnitType   string
	Panels     []CourtPanel
}

// CourtPanel은 법원에 속한 재판부의 ID와 이름입니다.
type CourtPanel struct {
	ID   string
	Name string
}
