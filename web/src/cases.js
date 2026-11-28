// 사건 데이터를 다루는 화면 독립적인 함수와 상수입니다.

// 사건 종류별 표시 순서와 배지 색(Tailwind 클래스). 어두운 화면에서도 읽히도록 dark: 색을 함께 둡니다.
export const CATEGORIES = [
  { value: "civil", label: "민사", badge: "bg-blue-50 text-blue-700 dark:bg-blue-500/15 dark:text-blue-300" },
  { value: "civil_application", label: "민사신청", badge: "bg-slate-100 text-slate-700 dark:bg-slate-500/20 dark:text-slate-300" },
  { value: "family", label: "가사", badge: "bg-emerald-50 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300" },
  { value: "administrative", label: "행정", badge: "bg-amber-50 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300" },
  { value: "criminal", label: "형사", badge: "bg-red-50 text-red-700 dark:bg-red-500/15 dark:text-red-300" },
];
// 종류 코드로 바로 찾을 수 있게 { civil: {...}, family: {...} } 모양의 객체로 바꿔 둡니다.
export const CATEGORY_BY_VALUE = Object.fromEntries(CATEGORIES.map((c) => [c.value, c]));

// 필터의 초기값입니다. categories는 Set(중복 없는 값 모음)이라 선택한 종류를 넣고 빼기 쉽습니다.
export const EMPTY_FILTERS = { query: "", categories: new Set(), year: "", courtId: "" };

// 필터 패널에서 적용 중인 조건 수입니다(검색어 제외). 접힌 필터 아이콘에 숫자로 보여 줍니다.
export function activeFilterCount(filters) {
  return filters.categories.size + (filters.year ? 1 : 0) + (filters.courtId ? 1 : 0);
}

// 정렬할 수 있는 열과 정렬 기준값.
// 각 함수는 사건 하나를 받아 비교할 값의 배열을 돌려줍니다. 앞의 값이 같으면 다음 값으로 비교합니다
// (예: 사건번호는 연도 → 사건부호 → 일련번호 순).
export const SORTERS = {
  number: (c) => [c.year, c.code, c.serial],
  category: (c) => [CATEGORIES.findIndex((x) => x.value === c.category)],
  title: (c) => [c.title],
  client: (c) => [c.parties.find((p) => p.client)?.name ?? ""],
  court: (c) => [c.court.name, c.court.panel],
};

// compareKeys는 두 기준값 배열을 앞에서부터 비교합니다.
// a가 앞이면 음수, 뒤면 양수, 같으면 0을 돌려줍니다(Array.sort가 요구하는 규칙).
// 문자열은 localeCompare(..., "ko")로 한국어 가나다순 비교를 합니다.
function compareKeys(a, b) {
  for (let i = 0; i < a.length; i++) {
    if (a[i] === b[i]) continue;
    if (typeof a[i] === "number") return a[i] - b[i];
    return String(a[i]).localeCompare(String(b[i]), "ko");
  }
  return 0;
}

// 필터와 정렬을 적용한 새 배열을 돌려줍니다.
// 원본 배열(cases)은 바꾸지 않습니다. React에서는 상태 값을 직접 고치지 않는 것이 규칙입니다.
export function filterAndSort(cases, filters, sort) {
  // 검색어에서 모든 공백을 지우고 소문자로 바꿉니다. "2026 가합"으로 검색해도 "2026가합"과 맞도록 하기 위해서입니다.
  const q = filters.query.trim().replace(/\s+/g, "").toLowerCase();
  const rows = cases.filter((c) => {
    if (filters.categories.size && !filters.categories.has(c.category)) return false;
    if (filters.year && String(c.year) !== filters.year) return false;
    if (filters.courtId && c.court.id !== filters.courtId) return false;
    if (!q) return true;
    // 검색 대상(사건번호·사건명·법원·재판부·당사자)을 "|"로 이어 붙인 하나의 문자열(건초더미, haystack)에서
    // 검색어(바늘)를 찾습니다. 여기서도 공백을 지워 검색어와 같은 조건으로 비교합니다.
    const hay = [c.number, c.title, c.court.name, c.court.panel, ...c.parties.flatMap((p) => [p.role, p.name])]
      .join("|")
      .replace(/\s+/g, "")
      .toLowerCase();
    return hay.includes(q);
  });
  const key = SORTERS[sort.key];
  const dir = sort.dir === "asc" ? 1 : -1; // 내림차순이면 비교 결과의 부호를 뒤집습니다.
  // filter가 이미 새 배열을 만들었으므로 여기서 sort로 제자리 정렬해도 원본은 바뀌지 않습니다.
  return rows.sort((a, b) => dir * compareKeys(key(a), key(b)));
}

// "피고2"처럼 지위 끝의 번호를 뗀 값입니다. 같은 쪽 당사자를 묶을 때 씁니다.
// /\d+$/는 "문자열 끝에 붙은 숫자 하나 이상"을 뜻하는 정규식입니다.
export function roleGroup(role) {
  return role.replace(/\d+$/, "");
}

// 지위를 같은 쪽 당사자끼리 비교할 수 있는 값으로 바꿉니다(Clio 서버의 PartySide 함수와 같은 규칙).
// 괄호 안의 표기, 띄어쓰기, 끝의 번호를 지웁니다. 예: "원고1" → "원고", "피고(반소원고)" → "피고".
export function partySide(role) {
  return role.replace(/[(（][^)）]*[)）]/g, "").replace(/\s+/g, "").replace(/\d+$/, "");
}

// 의뢰인들의 지위(partySide)를 처음 나온 순서대로 중복 없이 돌려줍니다.
// 두 개 이상이면 지위가 다른 당사자를 함께 대리하는 것이므로 이해충돌입니다.
// new Set(...)으로 중복을 없애고, [...set]으로 다시 배열로 바꿉니다. filter(Boolean)은 빈 문자열을 뺍니다.
export function clientSides(parties) {
  return [...new Set(parties.filter((p) => p.client).map((p) => partySide(p.role)).filter(Boolean))];
}

// 당사자를 지위별로 묶어 입력 순서대로 [{role, names}]로 돌려줍니다.
// 바로 앞 묶음과 지위가 같으면 그 묶음에 이름을 더하고, 다르면 새 묶음을 만듭니다.
// 예: [원고1 김, 원고2 이, 피고 박] → [{role: "원고", names: [김, 이]}, {role: "피고", names: [박]}]
export function groupParties(parties) {
  const groups = [];
  for (const p of parties) {
    const role = roleGroup(p.role);
    const last = groups[groups.length - 1];
    if (last && last.role === role) last.names.push(p.name);
    else groups.push({ role, names: [p.name] });
  }
  return groups;
}

// 이름이 여럿이면 "김민준 외 2"처럼 줄여서 보여 줍니다.
export function namesSummary(names) {
  return names.length > 1 ? `${names[0]} 외 ${names.length - 1}` : names[0];
}

// 당사자를 의뢰인과 그 밖의 당사자로 나눕니다.
// clients는 입력 순서대로의 의뢰인 목록, others는 의뢰인이 아닌 당사자를 지위별로 묶은 목록입니다.
// 같은 지위라도 의뢰인이 아닌 사람은 others에 들어갑니다(예: 원고1만 의뢰인이면 원고2는 others).
export function splitParties(parties) {
  return {
    clients: parties.filter((p) => p.client),
    others: groupParties(parties.filter((p) => !p.client)),
  };
}
