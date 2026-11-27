import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { CircleAlert, FilePlus2, Filter, SearchX, Search, X } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "@/components/ui/empty";
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from "@/components/ui/input-group";
import { Kbd } from "@/components/ui/kbd";
import { Spinner } from "@/components/ui/spinner";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import Filters from "./Filters.jsx";
import CaseTable from "./CaseTable.jsx";
import CaseDrawer from "./CaseDrawer.jsx";
import { EMPTY_FILTERS, activeFilterCount, filterAndSort } from "./cases.js";
import { isPlainKey, isTyping } from "./keyboard.js";

// 독바의 아이콘 버튼. 마우스를 올리면 오른쪽에 이름(과 단축키)이 나옵니다.
//   - pressed: 켜져 있는 상태(예: 필터 패널이 열림)
//   - highlight: 주목할 상태(예: 필터가 적용 중)
//   - badge: 아이콘 오른쪽 위에 붙일 숫자
function DockButton({ label, shortcut, pressed, highlight, badge, onClick, children }) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          variant={pressed ? "secondary" : "ghost"}
          size="icon"
          aria-label={label}
          aria-pressed={pressed}
          onClick={onClick}
          className={cn("relative size-9", highlight && "text-primary")}
        >
          {children}
          {badge != null && (
            <Badge className="absolute -top-1 -right-1 h-4 min-w-4 rounded-full px-1 text-[10px] tabular-nums">
              {badge}
            </Badge>
          )}
        </Button>
      </TooltipTrigger>
      <TooltipContent side="right">
        {label}
        {shortcut && <Kbd>{shortcut}</Kbd>}
      </TooltipContent>
    </Tooltip>
  );
}

// 사건목록 탭의 내용: 왼쪽 필터, 검색창, 사건 표, 상세 패널.
//   - cases, error: App이 불러온 사건 목록(등록한 사건도 App이 여기에 더합니다)
//   - active: 이 탭이 보이는 중인지. 보일 때만 "/" 단축키가 동작합니다.
//   - onNewCase: 사건 등록 탭을 엽니다.
//   - openCase: 값이 바뀌면 그 사건의 상세 패널을 엽니다(방금 등록한 사건 보여 주기).
//   - onOpenCaseTab: 사건을 사건 탭(전체 화면)으로 엽니다.
export default function CaseListPage({ cases, error, active, onNewCase, openCase, onOpenCaseTab }) {
  const [filters, setFilters] = useState(EMPTY_FILTERS);
  const [sort, setSort] = useState({ key: "number", dir: "desc" });
  const [selected, setSelected] = useState(null);
  const closePreview = useCallback(() => setSelected(null), []);
  // 필터 패널은 기본으로 접어 두고, 펼침 여부는 다음 방문 때도 유지합니다.
  const [filtersOpen, setFiltersOpen] = useState(() => localStorage.getItem("casepulse.filtersOpen") === "1");
  useEffect(() => localStorage.setItem("casepulse.filtersOpen", filtersOpen ? "1" : "0"), [filtersOpen]);
  const filterCount = activeFilterCount(filters);
  const searchRef = useRef(null);

  useEffect(() => {
    if (openCase) setSelected(openCase);
  }, [openCase]);

  // "/" 키로 검색창으로 이동합니다.
  useEffect(() => {
    if (!active) return;
    const onKey = (e) => {
      if (e.key === "/" && isPlainKey(e) && !isTyping()) {
        e.preventDefault();
        searchRef.current?.focus();
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [active]);

  const all = cases ?? [];
  const years = useMemo(() => [...new Set(all.map((c) => c.year))].sort((a, b) => b - a), [cases]);
  const courts = useMemo(() => {
    const m = new Map();
    for (const c of all) {
      const e = m.get(c.court.id) ?? { id: c.court.id, name: c.court.name, count: 0 };
      e.count++;
      m.set(c.court.id, e);
    }
    return [...m.values()].sort((a, b) => b.count - a.count || a.name.localeCompare(b.name, "ko"));
  }, [cases]);
  const filtered = useMemo(() => filterAndSort(all, filters, sort), [cases, filters, sort]);

  const onSort = (key) =>
    setSort((s) =>
      s.key === key ? { key, dir: s.dir === "asc" ? "desc" : "asc" } : { key, dir: key === "number" ? "desc" : "asc" },
    );

  let body;
  if (error) {
    body = (
      <Alert variant="destructive">
        <CircleAlert />
        <AlertTitle>사건 목록을 불러오지 못했습니다</AlertTitle>
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  } else if (!cases) {
    body = (
      <div className="flex flex-1 items-center justify-center">
        <Spinner className="size-6" />
      </div>
    );
  } else if (filtered.length === 0) {
    body = (
      <Empty className="flex-1 rounded-xl border bg-card">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <SearchX />
          </EmptyMedia>
          <EmptyTitle>조건에 맞는 사건이 없습니다</EmptyTitle>
          <EmptyDescription>검색어나 필터를 바꿔 보세요.</EmptyDescription>
        </EmptyHeader>
        <EmptyContent>
          <Button variant="outline" onClick={() => setFilters(EMPTY_FILTERS)}>
            필터 초기화
          </Button>
        </EmptyContent>
      </Empty>
    );
  } else {
    body = (
      <CaseTable
        rows={filtered}
        sort={sort}
        onSort={onSort}
        selectedId={selected?.id}
        onSelect={setSelected}
        onOpen={(c) => {
          setSelected(null);
          onOpenCaseTab(c);
        }}
      />
    );
  }

  // 미리보기 패널(CaseDrawer)이 이 탭 영역 안에서만 뜨도록 바깥 상자를 relative로 둡니다.
  return (
    <div className="relative flex h-full overflow-hidden">
      <div className="flex h-full min-w-0 flex-1">
        {/* 왼쪽 독바: 이 화면에서 자주 쓰는 도구 아이콘을 세로로 둡니다. 필터 패널을 열어도 그대로 남습니다. */}
        <nav aria-label="사건목록 도구" className="flex w-12 shrink-0 flex-col items-center gap-1 border-r bg-background py-3">
          <DockButton
            label={filtersOpen ? "필터 닫기" : filterCount ? `필터 열기 (${filterCount}개 적용 중)` : "필터 열기"}
            pressed={filtersOpen}
            highlight={filterCount > 0}
            badge={filterCount || null}
            onClick={() => setFiltersOpen((open) => !open)}
          >
            <Filter />
          </DockButton>
          <DockButton label="신건등록" shortcut="N" onClick={onNewCase}>
            <FilePlus2 />
          </DockButton>
        </nav>

        <div className={cn("grid min-w-0 flex-1 gap-3 p-3", filtersOpen ? "grid-cols-[260px_1fr]" : "grid-cols-1")}>
          {filtersOpen && (
            <aside className="min-h-0 overflow-auto">
              <Filters
                cases={all}
                filters={filters}
                onChange={setFilters}
                years={years}
                courts={courts}
                onCollapse={() => setFiltersOpen(false)}
              />
            </aside>
          )}

          <div className="flex min-h-0 min-w-0 flex-col gap-2">
            <div className="flex shrink-0 items-center gap-3">
              <InputGroup className="max-w-md bg-background">
                <InputGroupAddon>
                  <Search />
                </InputGroupAddon>
                <InputGroupInput
                  ref={searchRef}
                  placeholder="사건번호, 사건명, 당사자, 법원, 재판부 검색"
                  value={filters.query}
                  onChange={(e) => setFilters({ ...filters, query: e.target.value })}
                />
                <InputGroupAddon align="inline-end">
                  {filters.query ? (
                    <InputGroupButton
                      size="icon-xs"
                      aria-label="검색어 지우기"
                      onClick={() => setFilters({ ...filters, query: "" })}
                    >
                      <X />
                    </InputGroupButton>
                  ) : (
                    <Kbd>/</Kbd>
                  )}
                </InputGroupAddon>
              </InputGroup>
              <span className="text-sm text-muted-foreground">
                {cases ? `전체 ${all.length}건 · 검색 결과 ${filtered.length}건` : ""}
              </span>
            </div>
            {body}
          </div>
        </div>
      </div>

      <CaseDrawer item={selected} active={active} onClose={closePreview} onOpenTab={onOpenCaseTab} />
    </div>
  );
}
