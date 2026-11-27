// 사건 데이터를 다루는 화면 독립적인 함수와 상수입니다.

// 사건 종류별 표시 순서와 배지 색(Tailwind 클래스). 어두운 화면에서도 읽히도록 dark: 색을 함께 둡니다.
export const CATEGORIES = [
  { value: "civil", label: "민사", badge: "bg-blue-50 text-blue-700 dark:bg-blue-500/15 dark:text-blue-300" },
  { value: "civil_application", label: "민사신청", badge: "bg-slate-100 text-slate-700 dark:bg-slate-500/20 dark:text-slate-300" },
  { value: "family", label: "가사", badge: "bg-emerald-50 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300" },
  { value: "administrative", label: "행정", badge: "bg-amber-50 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300" },
  { value: "criminal", label: "형사", badge: "bg-red-50 text-red-700 dark:bg-red-500/15 dark:text-red-300" },
];
export const CATEGORY_BY_VALUE = Object.fromEntries(CATEGORIES.map((c) => [c.value, c]));

export const EMPTY_FILTERS = { query: "", categories: new Set(), year: "", courtId: "" };

// 필터 패널에서 적용 중인 조건 수입니다(검색어 제외). 접힌 필터 아이콘에 숫자로 보여 줍니다.
export function activeFilterCount(filters) {
  return filters.categories.size + (filters.year ? 1 : 0) + (filters.courtId ? 1 : 0);
}

// 정렬할 수 있는 열과 정렬 기준값.
export const SORTERS = {
  number: (c) => [c.year, c.code, c.serial],
  category: (c) => [CATEGORIES.findIndex((x) => x.value === c.category)],
  title: (c) => [c.title],
  client: (c) => [c.parties.find((p) => p.client)?.name ?? ""],
  court: (c) => [c.court.name, c.court.panel],
};

function compareKeys(a, b) {
  for (let i = 0; i < a.length; i++) {
    if (a[i] === b[i]) continue;
    if (typeof a[i] === "number") return a[i] - b[i];
    return String(a[i]).localeCompare(String(b[i]), "ko");
  }
  return 0;
}

// 필터와 정렬을 적용한 새 배열을 돌려줍니다.
export function filterAndSort(cases, filters, sort) {
  const q = filters.query.trim().replace(/\s+/g, "").toLowerCase();
  const rows = cases.filter((c) => {
    if (filters.categories.size && !filters.categories.has(c.category)) return false;
    if (filters.year && String(c.year) !== filters.year) return false;
    if (filters.courtId && c.court.id !== filters.courtId) return false;
    if (!q) return true;
    const hay = [c.number, c.title, c.court.name, c.court.panel, ...c.parties.flatMap((p) => [p.role, p.name])]
      .join("|")
      .replace(/\s+/g, "")
      .toLowerCase();
    return hay.includes(q);
  });
  const key = SORTERS[sort.key];
  const dir = sort.dir === "asc" ? 1 : -1;
  return rows.sort((a, b) => dir * compareKeys(key(a), key(b)));
}

// "피고2"처럼 지위 끝의 번호를 뗀 값입니다. 같은 쪽 당사자를 묶을 때 씁니다.
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
export function clientSides(parties) {
  return [...new Set(parties.filter((p) => p.client).map((p) => partySide(p.role)).filter(Boolean))];
}

// 당사자를 지위별로 묶어 입력 순서대로 [{role, names}]로 돌려줍니다.
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
