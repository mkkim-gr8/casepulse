// Package casepulse는 법률 사건 관리 시스템에서 사용하는 자료형을 정의합니다.
package casepulse

// InstitutionType은 법률 업무와 관련된 공공기관의 유형을 나타냅니다.
type InstitutionType string

const (
	InstitutionTypeCourt       InstitutionType = "court"
	InstitutionTypeProsecution InstitutionType = "prosecution"
	InstitutionTypePolice      InstitutionType = "police"
	InstitutionTypeRegistry    InstitutionType = "registry"
)

// CourtLevel은 법원의 조직상 단계를 나타냅니다.
type CourtLevel string

const (
	CourtLevelSupreme  CourtLevel = "supreme"
	CourtLevelHigh     CourtLevel = "high"
	CourtLevelDistrict CourtLevel = "district"
	CourtLevelLocal    CourtLevel = "local"
)

// CourtType은 법원의 관할 분야를 나타냅니다.
type CourtType string

const (
	CourtTypeGeneral        CourtType = "general"
	CourtTypeFamily         CourtType = "family"
	CourtTypeAdministrative CourtType = "administrative"
	CourtTypePatent         CourtType = "patent"
	CourtTypeRehabilitation CourtType = "rehabilitation"
)

// CourtUnitType은 본원, 지원 또는 시·군법원 여부를 나타냅니다.
type CourtUnitType string

const (
	CourtUnitTypeMain            CourtUnitType = "main"
	CourtUnitTypeSupport         CourtUnitType = "support"
	CourtUnitTypeMunicipalCounty CourtUnitType = "municipal_county"
)

// CourtPanelType은 재판부가 법원 소재지 안에 있는지 원외에 있는지 나타냅니다.
type CourtPanelType string

const (
	CourtPanelTypeAtSeat    CourtPanelType = "at_seat"
	CourtPanelTypeOutOfSeat CourtPanelType = "out_of_seat"
)

// Institution은 법률 기관의 연락처 정보를 담습니다.
type Institution struct {
	ID      string
	Name    string
	Type    InstitutionType
	Address string
	Phone   string
	Email   string
}

// Court는 법원의 분류, 소재지, 관할 지역과 소속 재판부 정보를 담습니다.
type Court struct {
	Institution
	Level         CourtLevel
	CourtType     CourtType
	UnitType      CourtUnitType
	ParentCourtID string
	Seat          string
	Jurisdictions []string
	Panels        []CourtPanel
}

// CourtPanel은 법원 또는 상위 재판부에 소속된 하위 재판부와 소재지를 나타냅니다.
type CourtPanel struct {
	ID     string
	Name   string
	Type   CourtPanelType
	Seat   string
	Panels []CourtPanel
}
