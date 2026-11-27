import { useEffect, useRef, useState } from "react";
import { ArrowDown, ArrowUp, ArrowUpDown } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Spinner } from "@/components/ui/spinner";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import CategoryBadge from "./CategoryBadge.jsx";
import { namesSummary, roleGroup, splitParties } from "./cases.js";

// 처음에 그리는 행 수이자, 목록 끝에 닿을 때마다 더 그리는 행 수입니다.
const BATCH_SIZE = 50;

function Role({ children }) {
  return <span className="mr-1 text-xs text-muted-foreground">{children}</span>;
}

// 여러 명일 때만 마우스를 올리면 전체 목록을 보여 줍니다.
function WithTooltip({ content, children }) {
  if (!content) return children;
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span className="cursor-default">{children}</span>
      </TooltipTrigger>
      <TooltipContent>{content}</TooltipContent>
    </Tooltip>
  );
}

// 의뢰인 열: 첫 의뢰인의 지위와 이름, 여러 명이면 "외 N"과 전체 목록 툴팁.
function ClientCell({ clients }) {
  const sameRole = clients.every((p) => roleGroup(p.role) === roleGroup(clients[0].role));
  const role = sameRole ? roleGroup(clients[0].role) : clients[0].role;
  return (
    <WithTooltip content={clients.length > 1 && clients.map((p) => `${p.role} ${p.name}`).join(", ")}>
      <Role>{role}</Role>
      {namesSummary(clients.map((p) => p.name))}
    </WithTooltip>
  );
}

// 다른 당사자 열: 의뢰인이 아닌 당사자를 지위별로 묶어 첫 묶음을 보여 주고, 나머지 묶음은 +N으로 표시합니다.
function OthersCell({ others }) {
  if (others.length === 0) return <span className="text-muted-foreground">—</span>;
  const [first, ...rest] = others;
  return (
    <span className="inline-flex items-center gap-1.5">
      <WithTooltip content={first.names.length > 1 && first.names.join(", ")}>
        <Role>{first.role}</Role>
        {namesSummary(first.names)}
      </WithTooltip>
      {rest.length > 0 && (
        <WithTooltip content={rest.map((g) => `${g.role} ${g.names.join(", ")}`).join(" · ")}>
          <Badge variant="secondary" className="h-4 rounded-full px-1.5 text-[10px]">
            +{rest.length}
          </Badge>
        </WithTooltip>
      )}
    </span>
  );
}

function SortHead({ id, label, sort, onSort, className }) {
  const active = sort.key === id;
  const Icon = active ? (sort.dir === "asc" ? ArrowUp : ArrowDown) : ArrowUpDown;
  return (
    <TableHead
      className={className}
      aria-sort={active ? (sort.dir === "asc" ? "ascending" : "descending") : "none"}
    >
      <button
        type="button"
        onClick={() => onSort(id)}
        className="inline-flex items-center gap-1 rounded-sm font-medium outline-none hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring/50"
      >
        {label}
        <Icon className={cn("size-3.5", active ? "text-primary" : "text-muted-foreground/60")} />
      </button>
    </TableHead>
  );
}

// 사건 목록 표. 스크롤이 목록 끝에 닿으면 다음 BATCH_SIZE건을 이어서 그립니다(무한 스크롤).
//
// rows는 검색·필터·정렬을 마친 전체 목록입니다. 데이터는 이미 모두 받아 두었으므로
// 서버에 다시 요청하지 않고, 한 번에 그리는 행 수만 늘려 화면이 무거워지지 않게 합니다.
export default function CaseTable({ rows, sort, onSort, selectedId, onSelect }) {
  const [visibleCount, setVisibleCount] = useState(BATCH_SIZE);
  const scrollRef = useRef(null);
  const sentinelRef = useRef(null);
  const hasMore = visibleCount < rows.length;

  // 검색·필터·정렬이 바뀌어 목록이 달라지면 맨 위부터 다시 보여 줍니다.
  useEffect(() => {
    setVisibleCount(BATCH_SIZE);
    if (scrollRef.current) scrollRef.current.scrollTop = 0;
  }, [rows]);

  // 표 맨 아래의 표시용 요소(sentinel)가 스크롤 영역에 들어오면 다음 묶음을 그립니다.
  // 미리 200px 앞에서 불러와 끊김 없이 이어지게 합니다.
  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel || !hasMore) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) setVisibleCount((n) => n + BATCH_SIZE);
      },
      { root: scrollRef.current, rootMargin: "0px 0px 200px 0px" },
    );
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [hasMore, visibleCount]);

  const shown = Math.min(visibleCount, rows.length);
  return (
    <>
      {/* 표 머리글을 고정하기 위해 스크롤은 이 상자가 맡습니다. */}
      <div ref={scrollRef} className="case-scroll min-h-0 flex-1 overflow-auto rounded-xl bg-card ring-1 ring-foreground/10">
        <Table className="[&_td]:py-1.5 [&_th]:h-9">
          <TableHeader className="sticky top-0 z-10 bg-card shadow-[inset_0_-1px_0_var(--border)]">
            <TableRow className="hover:bg-transparent">
              <SortHead id="number" label="사건번호" sort={sort} onSort={onSort} />
              <SortHead id="category" label="종류" sort={sort} onSort={onSort} />
              <SortHead id="title" label="사건명" sort={sort} onSort={onSort} />
              <SortHead id="client" label="의뢰인" sort={sort} onSort={onSort} />
              <TableHead>다른 당사자</TableHead>
              <SortHead id="court" label="계속 법원" sort={sort} onSort={onSort} />
              <TableHead>재판부</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.slice(0, visibleCount).map((c) => {
              const { clients, others } = splitParties(c.parties);
              return (
                <TableRow
                  key={c.id}
                  onClick={() => onSelect(c)}
                  data-state={selectedId === c.id ? "selected" : undefined}
                  className="cursor-pointer"
                >
                  <TableCell className="font-mono tabular-nums">{c.number}</TableCell>
                  <TableCell>
                    <CategoryBadge category={c.category} />
                  </TableCell>
                  <TableCell className="max-w-64 truncate" title={c.title}>
                    {c.title}
                  </TableCell>
                  <TableCell>
                    <ClientCell clients={clients} />
                  </TableCell>
                  <TableCell>
                    <OthersCell others={others} />
                  </TableCell>
                  <TableCell>{c.court.name}</TableCell>
                  <TableCell>{c.court.panel}</TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
        {hasMore && (
          <div ref={sentinelRef} className="flex justify-center p-3">
            <Spinner />
          </div>
        )}
      </div>
      <p className="shrink-0 text-xs text-muted-foreground">
        {rows.length}건 중 {shown}건 표시{hasMore ? " · 아래로 스크롤하면 더 보입니다" : ""}
      </p>
    </>
  );
}
