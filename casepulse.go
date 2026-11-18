// Package casepulse는 법무법인·변호사사무소의 사건 관리(ERP) 시스템에서 쓰는
// 법률 기관 정보와 그 자료형을 제공합니다.
//
// 현재 제공하는 기관 목록은 다음과 같습니다.
//   - 법원: Courts (courts.go)
//   - 공소청: ProsecutionOffices (prosecutions.go)
//   - 중대범죄수사청: InvestigationAgencies (investigations.go)
//
// 모든 목록은 법령 별표(documents/pdfs)와 각 기관 홈페이지를 바탕으로 소스 코드에 직접 적은 값입니다.
// 프로그램이 실행 중에 PDF나 웹 페이지를 읽지는 않으므로, 원본 자료가 바뀌면 소스 코드를 고쳐야 합니다.
// 각 목록은 패키지가 처음 로드될 때 한 번만 만들어지고, 공개 함수는 그 복사본을 돌려줍니다.
package casepulse

// InstitutionType은 법률 업무와 관련된 공공기관의 종류입니다.
// Institution.Type에 들어가며, 여러 종류의 기관을 한 목록에서 다룰 때 구분하는 데 씁니다.
type InstitutionType string

const (
	// InstitutionTypeCourt는 법원입니다(대법원, 고등법원, 지방법원, 지원, 시·군법원 등).
	InstitutionTypeCourt InstitutionType = "court"
	// InstitutionTypeProsecution은 공소청입니다(옛 검찰청).
	InstitutionTypeProsecution InstitutionType = "prosecution"
	// InstitutionTypeInvestigation은 중대범죄수사청과 그 지방수사청입니다.
	InstitutionTypeInvestigation InstitutionType = "investigation"
	// InstitutionTypePolice는 경찰 기관입니다. 아직 목록 데이터는 없습니다.
	InstitutionTypePolice InstitutionType = "police"
	// InstitutionTypeRegistry는 등기소입니다. 아직 목록 데이터는 없습니다.
	InstitutionTypeRegistry InstitutionType = "registry"
)

// CourtLevel은 법원 조직에서의 단계(심급과 대체로 대응)입니다.
type CourtLevel string

const (
	// CourtLevelSupreme은 대법원입니다.
	CourtLevelSupreme CourtLevel = "supreme"
	// CourtLevelHigh는 고등법원 단계입니다. 원외재판부와 특허법원도 이 단계에 둡니다.
	CourtLevelHigh CourtLevel = "high"
	// CourtLevelDistrict는 지방법원 단계입니다. 지원, 가정법원, 행정법원, 회생법원도 이 단계에 둡니다.
	CourtLevelDistrict CourtLevel = "district"
	// CourtLevelLocal은 시·군법원입니다.
	CourtLevelLocal CourtLevel = "local"
)

// CourtType은 법원이 맡는 사건 분야입니다.
type CourtType string

const (
	// CourtTypeGeneral은 민사·형사 등 일반 사건을 맡는 법원입니다.
	CourtTypeGeneral CourtType = "general"
	// CourtTypeFamily는 가정법원과 그 지원입니다.
	CourtTypeFamily CourtType = "family"
	// CourtTypeAdministrative는 행정법원입니다.
	CourtTypeAdministrative CourtType = "administrative"
	// CourtTypePatent는 특허법원입니다.
	CourtTypePatent CourtType = "patent"
	// CourtTypeRehabilitation은 회생법원입니다.
	CourtTypeRehabilitation CourtType = "rehabilitation"
)

// CourtUnitType은 같은 단계 안에서 본원·지원·시·군법원을 구분합니다.
type CourtUnitType string

const (
	// CourtUnitTypeMain은 본원입니다(지원이나 시·군법원이 아닌 법원).
	CourtUnitTypeMain CourtUnitType = "main"
	// CourtUnitTypeSupport는 지원입니다(예: 인천지방법원 부천지원).
	CourtUnitTypeSupport CourtUnitType = "support"
	// CourtUnitTypeMunicipalCounty는 시·군법원입니다(예: 인천지방법원 김포시법원).
	CourtUnitTypeMunicipalCounty CourtUnitType = "municipal_county"
)

// ProsecutionLevel은 공소청 조직에서의 단계입니다.
type ProsecutionLevel string

const (
	// ProsecutionLevelNational은 최상위 기관인 공소청(본청)입니다.
	ProsecutionLevelNational ProsecutionLevel = "national"
	// ProsecutionLevelRegional은 광역공소청입니다. 고등법원에 대응합니다.
	ProsecutionLevelRegional ProsecutionLevel = "regional"
	// ProsecutionLevelDistrict는 지방공소청과 그 지청입니다. 지방법원과 지원에 대응합니다.
	ProsecutionLevelDistrict ProsecutionLevel = "district"
)

// ProsecutionUnitType은 지방공소청 단계에서 본청과 지청을 구분합니다.
type ProsecutionUnitType string

const (
	// ProsecutionUnitTypeMain은 지청이 아닌 기관입니다(공소청, 광역공소청, 지방공소청).
	ProsecutionUnitTypeMain ProsecutionUnitType = "main"
	// ProsecutionUnitTypeBranch는 지방공소청의 지청입니다(예: 수원지방공소청 성남지청).
	ProsecutionUnitTypeBranch ProsecutionUnitType = "branch"
)

// InvestigationLevel은 중대범죄수사청 조직에서의 단계입니다.
type InvestigationLevel string

const (
	// InvestigationLevelNational은 중대범죄수사청 본청입니다.
	InvestigationLevelNational InvestigationLevel = "national"
	// InvestigationLevelDistrict는 지방중대범죄수사청(지방수사청)입니다.
	InvestigationLevelDistrict InvestigationLevel = "district"
)

// Institution은 모든 기관이 공통으로 가지는 식별 정보와 연락처입니다.
// Court, ProsecutionOffice, InvestigationAgency에 포함(embed)되어 쓰입니다.
type Institution struct {
	// ID는 패키지 안에서 기관을 찾을 때 쓰는 고유 식별자입니다(예: "court-district-incheon").
	// 기관 종류별 접두사로 시작하며, 상위 기관을 가리키는 ParentCourtID 등의 값으로도 쓰입니다.
	ID string
	// Name은 기관의 공식 명칭입니다(예: "인천지방법원 부천지원").
	Name string
	// Type은 기관의 종류입니다.
	Type InstitutionType
	// PostalCode는 5자리 우편번호입니다. 주소를 확인하지 못한 기관은 비어 있습니다.
	PostalCode string
	// Address는 정규화한 주소이며 우편번호는 넣지 않습니다(예: "서울특별시 서초구 서초대로 219").
	Address string
	// Phone은 대표 전화번호입니다. 현재 제공하는 기관 목록에서는 모두 비어 있습니다.
	Phone string
	// Email은 대표 이메일 주소입니다. 현재 제공하는 기관 목록에서는 모두 비어 있습니다.
	Email string
}

// Court는 법원 한 곳의 분류, 소속 관계, 소재지, 관할 지역과 소속 재판부입니다.
type Court struct {
	Institution
	// Level은 법원의 단계입니다.
	Level CourtLevel
	// CourtType은 법원이 맡는 사건 분야입니다.
	CourtType CourtType
	// UnitType은 본원·지원·시·군법원 구분입니다.
	UnitType CourtUnitType
	// ParentCourtID는 바로 위 법원의 ID입니다. 대법원만 비어 있습니다.
	// 지원은 본원을, 시·군법원은 본원 또는 지원을, 고등법원(원외재판부 포함)은 대법원을 가리킵니다.
	ParentCourtID string
	// Seat은 법령 별표에 적힌 소재지입니다(예: "부천시", 시·군법원은 "갈말읍"처럼 읍 단위).
	// 다만 광주는 전남광주통합특별시 출범에 맞춰 "전남광주통합특별시"로 정규화합니다.
	Seat string
	// Jurisdictions는 관할 지역 목록입니다.
	// 시·도 이름은 그 시·도의 첫 항목에만 붙이고, 뒤에 이어지는 항목은 같은 시·도에 속한 것으로 읽습니다.
	// 예: {"서울특별시 종로구", "중구"}는 서울특별시 종로구와 서울특별시 중구입니다.
	// 전국을 관할하는 법원은 {"전국"}이며, 관할을 따로 적지 않은 법원은 비어 있습니다.
	Jurisdictions []string
	// Panels는 이 법원에 소속된 재판부 목록이며 재판부구성표의 순서를 따릅니다.
	Panels []CourtPanel
}

// CourtPanel은 법원에 소속된 재판부 하나입니다(예: "제1민사부", "가사제1단독").
//
// 원외재판부는 재판부가 아니라 별도의 고등법원 Court(예: "광주고등법원(전주)")로 등록하므로,
// 모든 재판부는 소속 법원의 소재지에서 재판합니다.
type CourtPanel struct {
	// ID는 "<법원 ID>-panel-<3자리 일련번호>" 형식의 고유 식별자입니다(예: "court-high-seoul-panel-001").
	// 일련번호는 재판부구성표의 순서대로 매기므로, 구성표가 바뀌면 같은 재판부의 ID도 바뀔 수 있습니다.
	ID string
	// Name은 재판부구성표에 적힌 재판부 이름입니다. 대괄호 안의 사건 분야 안내와 판사 이름은 넣지 않습니다.
	Name string
	// Seat은 재판부의 소재지로, 소속 법원의 Seat과 같습니다.
	Seat string
}

// ProsecutionOffice는 공소청 한 곳의 분류, 소속 관계, 소재지입니다.
type ProsecutionOffice struct {
	Institution
	// Level은 공소청·광역공소청·지방공소청 단계입니다.
	Level ProsecutionLevel
	// UnitType은 본청(지청이 아닌 기관)과 지청을 구분합니다.
	UnitType ProsecutionUnitType
	// ParentOfficeID는 바로 위 기관의 ID입니다. 최상위 공소청만 비어 있습니다.
	// 광역공소청은 공소청을, 지방공소청은 광역공소청을, 지청은 지방공소청을 가리킵니다.
	ParentOfficeID string
	// Seat은 규정 별표에 적힌 위치입니다(예: "부천시", "영월읍").
	Seat string
}

// InvestigationAgency는 중대범죄수사청 본청 또는 지방수사청 한 곳의 분류, 소속 관계, 소재지, 관할 지역입니다.
type InvestigationAgency struct {
	Institution
	// ShortName은 현판·표지판 등에 쓸 수 있는 약칭입니다(예: "서울지방수사청"). 본청은 비어 있습니다.
	ShortName string
	// Level은 본청과 지방수사청을 구분합니다.
	Level InvestigationLevel
	// ParentAgencyID는 바로 위 기관의 ID입니다. 본청은 비어 있고, 지방수사청은 모두 본청을 가리킵니다.
	ParentAgencyID string
	// Seat은 직제 별표 1에 적힌 위치입니다(예: 대전지방수사청은 "세종특별자치시").
	Seat string
	// Jurisdictions는 관할 지역 목록이며, 읽는 방법은 Court.Jurisdictions와 같습니다.
	Jurisdictions []string
}
