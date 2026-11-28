import { useEffect, useState } from "react";
import { Building2, Check, Copy, Info, UserRound, Users } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { cn } from "@/lib/utils";
import CategoryBadge from "./CategoryBadge.jsx";

// 사건 상세 정보 카드들입니다. 오른쪽 상세 패널(CaseDrawer)과 사건 탭(CaseDetailPage)이 함께 씁니다.

// 이름-값 목록(dl). 왼쪽이 항목 이름, 오른쪽이 값입니다.
// items는 [["사건번호", 값], ["사건명", 값], ...] 모양의 배열입니다.
// className="contents"는 감싼 div가 레이아웃에 영향을 주지 않게 해서 dt/dd가 바로 격자 칸에 놓이게 합니다.
function Fields({ items }) {
  return (
    <dl className="grid grid-cols-[5.5rem_1fr] gap-x-3 gap-y-2 text-sm">
      {items.map(([k, v]) => (
        <div key={k} className="contents">
          <dt className="text-muted-foreground">{k}</dt>
          <dd>{v}</dd>
        </div>
      ))}
    </dl>
  );
}

// 사건번호를 클립보드에 복사하는 버튼. 복사하면 잠깐 "복사됨"으로 바뀝니다.
export function CopyNumberButton({ number }) {
  const [copied, setCopied] = useState(false);
  // 다른 사건을 보게 되면(number가 바뀌면) "복사됨" 표시를 원래대로 돌립니다.
  useEffect(() => setCopied(false), [number]);
  return (
    <Button
      variant="ghost"
      size="xs"
      // navigator.clipboard는 브라우저의 클립보드 기능입니다. 지원하지 않는 환경이면 ?.로 조용히 건너뜁니다.
      onClick={() => navigator.clipboard?.writeText(number).then(() => setCopied(true))}
    >
      {copied ? <Check /> : <Copy />}
      {copied ? "복사됨" : "사건번호 복사"}
    </Button>
  );
}

// CaseInfoCard는 사건번호·구성(연도/부호/일련번호)·종류·사건명을 보여 주는 카드입니다.
// size = "sm"은 기본값으로, 미리보기 패널은 작게(sm), 사건 탭은 크게(default) 씁니다.
// <>...</>는 Fragment로, 여러 요소를 감싸는 태그 없이 묶을 때 씁니다.
export function CaseInfoCard({ item, size = "sm" }) {
  return (
    <Card size={size}>
      <CardHeader>
        <CardTitle className="flex items-center gap-1.5">
          <Info className="size-4" /> 사건
        </CardTitle>
        <CardAction>
          <CopyNumberButton number={item.number} />
        </CardAction>
      </CardHeader>
      <CardContent>
        <Fields
          items={[
            ["사건번호", <span className="font-mono">{item.number}</span>],
            [
              "구성",
              <>
                {item.year}년 · 사건부호 <b>{item.code}</b> · 일련번호 {item.serial}
              </>,
            ],
            ["사건 종류", <CategoryBadge category={item.category} />],
            ["사건명", item.title],
          ]}
        />
      </CardContent>
    </Card>
  );
}

// PartiesCard는 당사자 목록 표입니다. 의뢰인 행은 배경색과 굵은 글씨, "의뢰인" 배지로 강조합니다.
export function PartiesCard({ item, size = "sm" }) {
  const clientCount = item.parties.filter((p) => p.client).length;
  return (
    <Card size={size}>
      <CardHeader>
        <CardTitle className="flex items-center gap-1.5">
          <Users className="size-4" /> 당사자 ({item.parties.length})
        </CardTitle>
        <CardAction className="text-xs text-muted-foreground">의뢰인 {clientCount}명</CardAction>
      </CardHeader>
      <CardContent className="px-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[30%] pl-4">지위</TableHead>
              <TableHead>이름</TableHead>
              <TableHead className="w-24 pr-4">의뢰인</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {item.parties.map((p, i) => (
              // 당사자에는 고유 ID가 없고 순서가 바뀌지 않으므로 배열 위치(i)를 key로 씁니다.
              <TableRow key={i} className={cn(p.client && "bg-primary/5")}>
                <TableCell className="pl-4">{p.role}</TableCell>
                <TableCell className={cn(p.client && "font-semibold")}>{p.name}</TableCell>
                <TableCell className="pr-4">
                  {p.client && (
                    <Badge className="gap-1">
                      <UserRound /> 의뢰인
                    </Badge>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}

// CourtCard는 계속 법원·재판부·소재지·주소를 보여 줍니다. 주소가 없는 법원(예: 신설 법원)은 안내 문구를 보여 줍니다.
export function CourtCard({ item, size = "sm" }) {
  return (
    <Card size={size}>
      <CardHeader>
        <CardTitle className="flex items-center gap-1.5">
          <Building2 className="size-4" /> 계속 법원
        </CardTitle>
      </CardHeader>
      <CardContent>
        <Fields
          items={[
            ["법원", item.court.name],
            ["재판부", item.court.panel],
            ["소재지", item.court.seat],
            [
              "주소",
              item.court.address ? (
                `(${item.court.postalCode}) ${item.court.address}`
              ) : (
                <span className="text-muted-foreground">등록된 주소 없음</span>
              ),
            ],
          ]}
        />
      </CardContent>
    </Card>
  );
}
