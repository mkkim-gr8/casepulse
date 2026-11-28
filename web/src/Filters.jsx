import { useMemo } from "react";
import { ChevronLeft, Filter } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import CategoryBadge from "./CategoryBadge.jsx";
import { CATEGORIES } from "./cases.js";

// shadcn Select는 빈 문자열 값을 쓸 수 없으므로 "전체"를 이 값으로 나타냅니다.
const ALL = "all";

// 왼쪽 필터 패널: 사건 종류, 접수 연도, 계속 법원. onCollapse는 패널을 접는 버튼에 연결됩니다.
export default function Filters({ cases, filters, onChange, years, courts, onCollapse }) {
  // 사건 종류별 사건 수입니다(예: { civil: 96, criminal: 48 }). 체크박스 옆에 숫자로 보여 줍니다.
  const counts = useMemo(() => {
    const m = {};
    for (const c of cases) m[c.category] = (m[c.category] ?? 0) + 1;
    return m;
  }, [cases]);

  // 사건 종류 체크박스를 누르면 선택 목록에 넣거나 뺍니다.
  // 기존 Set을 직접 고치지 않고 복사본(next)을 만들어 고친 뒤 넘겨야 React가 바뀐 것을 알아챕니다.
  const toggleCategory = (value) => {
    const next = new Set(filters.categories);
    next.has(value) ? next.delete(value) : next.add(value);
    onChange({ ...filters, categories: next });
  };
  // 적용 중인 필터가 하나라도 있으면 "초기화" 버튼을 켭니다.
  const active = filters.categories.size > 0 || filters.year || filters.courtId;

  return (
    <Card size="sm" className="gap-3 py-3">
      <CardHeader className="px-3">
        <CardTitle className="flex items-center gap-1.5 text-sm">
          <Filter className="size-4" /> 필터
        </CardTitle>
        <CardAction className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="xs"
            disabled={!active}
            onClick={() => onChange({ ...filters, categories: new Set(), year: "", courtId: "" })}
          >
            초기화
          </Button>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="ghost" size="icon-xs" aria-label="필터 접기" onClick={onCollapse}>
                <ChevronLeft />
              </Button>
            </TooltipTrigger>
            <TooltipContent side="right">필터 접기</TooltipContent>
          </Tooltip>
        </CardAction>
      </CardHeader>
      <CardContent className="flex flex-col gap-3 px-3">
        <section className="flex flex-col gap-2">
          <h3 className="text-xs font-medium text-muted-foreground">사건 종류</h3>
          {CATEGORIES.map((c) => (
            <div key={c.value} className="flex items-center gap-2">
              <Checkbox
                id={`filter-${c.value}`}
                checked={filters.categories.has(c.value)}
                onCheckedChange={() => toggleCategory(c.value)}
              />
              <Label htmlFor={`filter-${c.value}`} className="flex flex-1 items-center justify-between font-normal">
                <CategoryBadge category={c.value} />
                <span className="text-xs text-muted-foreground tabular-nums">{counts[c.value] ?? 0}</span>
              </Label>
            </div>
          ))}
        </section>
        <Separator />
        <section className="flex flex-col gap-2">
          <h3 className="text-xs font-medium text-muted-foreground">접수 연도</h3>
          <Select
            value={filters.year || ALL}
            onValueChange={(v) => onChange({ ...filters, year: v === ALL ? "" : v })}
          >
            <SelectTrigger size="sm" className="w-full">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={ALL}>전체</SelectItem>
              {years.map((y) => (
                <SelectItem key={y} value={String(y)}>
                  {y}년
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </section>
        <Separator />
        <section className="flex flex-col gap-2">
          <h3 className="text-xs font-medium text-muted-foreground">계속 법원</h3>
          <Select
            value={filters.courtId || ALL}
            onValueChange={(v) => onChange({ ...filters, courtId: v === ALL ? "" : v })}
          >
            <SelectTrigger size="sm" className="w-full">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={ALL}>전체 ({courts.length}곳)</SelectItem>
              {courts.map((c) => (
                <SelectItem key={c.id} value={c.id}>
                  {c.name} ({c.count})
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </section>
      </CardContent>
    </Card>
  );
}
