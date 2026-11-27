package main

// 이 파일은 데이터베이스가 생기기 전까지 쓸 더미 사건 목록을 만듭니다.
//
// 사건번호·당사자·사건명은 모두 지어낸 값이며 실제 사건과 관계없습니다.
// 법원과 재판부는 casepulse.Courts의 실제 목록에서 사건 종류에 맞는 것을 고르므로,
// 만든 사건은 모두 casepulse.Case.Validate를 통과해야 합니다(통과하지 못하면 패닉).
// 고정된 시드를 쓰므로 서버를 다시 켜도 같은 목록이 만들어집니다.

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/mkkim-gr8/casepulse"
)

// dummyCaseCount는 만들 더미 사건 수입니다.
const dummyCaseCount = 240

// caseTemplate은 사건 종류별로 더미 사건을 만들 때 쓰는 재료입니다.
type caseTemplate struct {
	category casepulse.CaseCategory
	// weight는 전체 더미 사건 중 이 종류가 뽑힐 상대 비율입니다.
	weight int
	codes  []string
	titles []string
	// court는 이 종류의 사건을 맡을 수 있는 법원인지 판단합니다.
	court func(casepulse.Court) bool
	// panelKeywords 중 하나를 이름에 포함한 재판부만 고릅니다. 앞의 키워드가 우선합니다.
	panelKeywords []string
	// parties는 당사자 목록을 만듭니다(의뢰인 표시는 markClients가 붙입니다).
	parties func(r *rand.Rand) []casepulse.Party
	// clientRoles는 의뢰인 쪽이 될 수 있는 지위입니다. 사건마다 이 중 하나를 골라 그 쪽을 대리합니다.
	clientRoles []string
}

var (
	surnames   = []string{"김", "이", "박", "최", "정", "강", "조", "윤", "장", "임", "한", "오", "서", "신", "권", "황", "안", "송", "류", "홍"}
	givenNames = []string{"민준", "서연", "도윤", "하은", "시우", "지유", "주원", "서윤", "예준", "지민", "하준", "수아", "건우", "채원", "현우", "지아", "우진", "다은", "선우", "유나"}
	companies  = []string{"한빛건설", "누리상사", "가온물산", "다솜테크", "바름개발", "새봄유통", "온새미로", "푸른식품", "해든에너지", "미르로지스"}
	agencies   = []string{"서울특별시장", "경기도지사", "강남구청장", "중부지방국세청장", "서울지방고용노동청장", "국민건강보험공단", "근로복지공단", "서초세무서장", "인천광역시 남동구청장", "부산광역시장"}
)

func personName(r *rand.Rand) string {
	return surnames[r.IntN(len(surnames))] + givenNames[r.IntN(len(givenNames))]
}

func companyName(r *rand.Rand) string {
	return "주식회사 " + companies[r.IntN(len(companies))]
}

// personOrCompany는 3분의 1 확률로 회사, 나머지는 사람 이름을 돌려줍니다.
func personOrCompany(r *rand.Rand) string {
	if r.IntN(3) == 0 {
		return companyName(r)
	}
	return personName(r)
}

// sides는 앞 지위 a명, 뒤 지위 b명의 당사자를 만듭니다. 같은 지위가 여러 명이면 "피고1", "피고2"처럼 번호를 붙입니다.
func sides(r *rand.Rand, roleA string, a int, nameA func(*rand.Rand) string, roleB string, b int, nameB func(*rand.Rand) string) []casepulse.Party {
	var parties []casepulse.Party
	add := func(role string, n int, name func(*rand.Rand) string) {
		for i := 1; i <= n; i++ {
			label := role
			if n > 1 {
				label = fmt.Sprintf("%s%d", role, i)
			}
			parties = append(parties, casepulse.Party{Role: label, Name: name(r)})
		}
	}
	add(roleA, a, nameA)
	add(roleB, b, nameB)
	return parties
}

// roleGroup은 "피고2"처럼 지위 끝의 번호를 뗀 값입니다.
func roleGroup(role string) string {
	return strings.TrimRight(role, "0123456789")
}

// markClients는 의뢰인 쪽 지위를 하나 골라, 그 지위의 당사자 중 일부 또는 전부를 의뢰인으로 표시합니다.
// 같은 쪽 당사자가 여럿이면 모두가 의뢰인은 아닐 수 있다는 점을 더미 데이터에도 반영합니다.
// 고른 지위의 당사자가 없으면(예: 신청인만 있는 사건) 첫 당사자를 의뢰인으로 합니다.
func markClients(r *rand.Rand, parties []casepulse.Party, clientRoles []string) {
	role := clientRoles[r.IntN(len(clientRoles))]
	var side []int
	for i, p := range parties {
		if roleGroup(p.Role) == role {
			side = append(side, i)
		}
	}
	if len(side) == 0 {
		side = []int{0}
	}
	marked := 0
	for _, i := range side {
		if r.IntN(10) < 6 {
			parties[i].Client = true
			marked++
		}
	}
	if marked == 0 {
		parties[side[r.IntN(len(side))]].Client = true
	}
}

// partyCount는 1명이 가장 많고 2~3명이 가끔 나오는 당사자 수를 돌려줍니다.
func partyCount(r *rand.Rand) int {
	switch n := r.IntN(10); {
	case n < 7:
		return 1
	case n < 9:
		return 2
	default:
		return 3
	}
}

func isGeneralDistrict(c casepulse.Court) bool {
	return c.Level == casepulse.CourtLevelDistrict && c.CourtType == casepulse.CourtTypeGeneral
}

var caseTemplates = []caseTemplate{
	{
		category: casepulse.CaseCategoryCivil, weight: 40,
		clientRoles: []string{"원고", "피고"},
		codes:       []string{"가합", "가단", "가소"},
		titles:      []string{"손해배상(기)", "대여금", "물품대금", "공사대금", "임대차보증금반환", "건물인도", "소유권이전등기", "부당이득금", "약정금", "채무부존재확인"},
		court:       isGeneralDistrict, panelKeywords: []string{"민사"},
		parties: func(r *rand.Rand) []casepulse.Party {
			return sides(r, "원고", partyCount(r), personOrCompany, "피고", partyCount(r), personOrCompany)
		},
	},
	{
		category: casepulse.CaseCategoryCivilApplication, weight: 15,
		clientRoles: []string{"채권자", "채무자"},
		codes:       []string{"카합", "카단", "카기"},
		titles:      []string{"부동산가압류", "채권가압류", "부동산처분금지가처분", "업무방해금지가처분", "강제집행정지", "직무집행정지가처분"},
		court:       isGeneralDistrict, panelKeywords: []string{"신청", "민사"},
		parties: func(r *rand.Rand) []casepulse.Party {
			if r.IntN(10) == 0 { // 상대방 없이 신청인만 있는 사건
				return sides(r, "신청인", 1, personOrCompany, "", 0, nil)
			}
			return sides(r, "채권자", 1, personOrCompany, "채무자", partyCount(r), personOrCompany)
		},
	},
	{
		category: casepulse.CaseCategoryFamily, weight: 15,
		clientRoles:   []string{"원고", "피고"},
		codes:         []string{"드합", "드단", "느합", "느단"},
		titles:        []string{"이혼 및 위자료", "재산분할", "양육비", "친권자 및 양육자 변경", "상속재산분할", "면접교섭"},
		court:         func(c casepulse.Court) bool { return c.CourtType == casepulse.CourtTypeFamily },
		panelKeywords: []string{"가사"},
		parties: func(r *rand.Rand) []casepulse.Party {
			p := sides(r, "원고", 1, personName, "피고", 1, personName)
			if r.IntN(3) == 0 {
				p = append(p, casepulse.Party{Role: "사건본인", Name: personName(r)})
			}
			return p
		},
	},
	{
		category: casepulse.CaseCategoryAdministrative, weight: 10,
		clientRoles: []string{"원고", "원고", "원고", "피고"},
		codes:       []string{"구합", "구단"},
		titles:      []string{"영업정지처분취소", "부당해고구제재심판정취소", "종합소득세부과처분취소", "요양불승인처분취소", "건축허가취소처분취소", "정보공개거부처분취소"},
		court: func(c casepulse.Court) bool {
			return c.CourtType == casepulse.CourtTypeAdministrative || (isGeneralDistrict(c) && c.UnitType == casepulse.CourtUnitTypeMain)
		},
		panelKeywords: []string{"행정", "제1부", "제2부", "제3부"},
		parties: func(r *rand.Rand) []casepulse.Party {
			return sides(r, "원고", partyCount(r), personOrCompany, "피고", 1, func(r *rand.Rand) string { return agencies[r.IntN(len(agencies))] })
		},
	},
	{
		category: casepulse.CaseCategoryCriminal, weight: 20,
		clientRoles: []string{"피고인"},
		codes:       []string{"고합", "고단", "고정"},
		titles:      []string{"사기", "업무상횡령", "배임", "상해", "도로교통법위반(음주운전)", "명예훼손", "특정경제범죄가중처벌등에관한법률위반(사기)", "폭행"},
		court:       isGeneralDistrict, panelKeywords: []string{"형사"},
		parties: func(r *rand.Rand) []casepulse.Party {
			return sides(r, "피고인", partyCount(r), personName, "검사", 1, personName)
		},
	},
}

// courtChoice는 한 법원과 그 법원에서 고를 수 있는 재판부 이름입니다.
type courtChoice struct {
	courtID string
	panels  []string
}

// choicesFor는 템플릿에 맞는 법원과 재판부 후보를 모읍니다. 맞는 재판부가 없는 법원은 뺍니다.
func choicesFor(t caseTemplate, courts []casepulse.Court) []courtChoice {
	var choices []courtChoice
	for _, c := range courts {
		if !t.court(c) {
			continue
		}
		for _, kw := range t.panelKeywords {
			var panels []string
			for _, p := range c.Panels {
				if strings.Contains(p.Name, kw) {
					panels = append(panels, p.Name)
				}
			}
			if len(panels) > 0 {
				choices = append(choices, courtChoice{c.ID, panels})
				break
			}
		}
	}
	if len(choices) == 0 {
		panic(fmt.Sprintf("dummy: %s 사건을 맡을 법원·재판부를 찾지 못했습니다", t.category))
	}
	return choices
}

// buildDummyCases는 더미 사건 목록을 만듭니다. 사건번호는 법원 안에서 겹치지 않습니다.
func buildDummyCases() []casepulse.Case {
	r := rand.New(rand.NewPCG(20261127, 1))
	courts := casepulse.Courts()

	choices := make([][]courtChoice, len(caseTemplates))
	totalWeight := 0
	for i, t := range caseTemplates {
		choices[i] = choicesFor(t, courts)
		totalWeight += t.weight
	}

	// 실제 업무처럼 일부 법원에 사건이 몰리도록, 각 종류마다 앞쪽(주요) 법원 몇 곳을 자주 고릅니다.
	pickCourt := func(cs []courtChoice) courtChoice {
		if r.IntN(3) > 0 {
			return cs[r.IntN(min(6, len(cs)))]
		}
		return cs[r.IntN(len(cs))]
	}

	seen := make(map[string]bool)
	cases := make([]casepulse.Case, 0, dummyCaseCount)
	for len(cases) < dummyCaseCount {
		n := r.IntN(totalWeight)
		ti := 0
		for n >= caseTemplates[ti].weight {
			n -= caseTemplates[ti].weight
			ti++
		}
		t := caseTemplates[ti]
		cc := pickCourt(choices[ti])

		parties := t.parties(r)
		markClients(r, parties, t.clientRoles)
		c := casepulse.Case{
			Number: casepulse.CaseNumber{
				Year:   2023 + r.IntN(4),
				Code:   t.codes[r.IntN(len(t.codes))],
				Serial: 100 + r.IntN(60000),
			},
			Title:    t.titles[r.IntN(len(t.titles))],
			Category: t.category,
			Parties:  parties,
			Court:    casepulse.CourtAssignment{CourtID: cc.courtID, PanelName: cc.panels[r.IntN(len(cc.panels))]},
		}
		key := cc.courtID + "/" + c.Number.String()
		if seen[key] {
			continue
		}
		seen[key] = true
		if err := c.Validate(); err != nil {
			panic(fmt.Sprintf("dummy: 만든 사건 %s가 검사를 통과하지 못했습니다: %v", c.Number, err))
		}
		cases = append(cases, c)
	}
	return cases
}
