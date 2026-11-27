import { useEffect, useState } from "react";
import { Building2, Check, Copy, Info, UserRound, Users } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { cn } from "@/lib/utils";
import CategoryBadge from "./CategoryBadge.jsx";

// 이름-값 목록(dl). 왼쪽이 항목 이름, 오른쪽이 값입니다.
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

// 오른쪽에서 열리는 사건 상세 패널.
export default function CaseDrawer({ item, onClose }) {
  const [copied, setCopied] = useState(false);
  useEffect(() => setCopied(false), [item]);
  const clientCount = item?.parties.filter((p) => p.client).length ?? 0;

  return (
    <Sheet open={!!item} onOpenChange={(open) => !open && onClose()}>
      <SheetContent className="w-full gap-0 sm:max-w-[480px]">
        {item && (
          <>
            <SheetHeader className="border-b">
              <SheetTitle className="font-mono text-lg">{item.number}</SheetTitle>
              <SheetDescription>{item.title}</SheetDescription>
            </SheetHeader>
            <div className="flex flex-col gap-3 overflow-y-auto p-4">
              <Card size="sm">
                <CardHeader>
                  <CardTitle className="flex items-center gap-1.5">
                    <Info className="size-4" /> 사건
                  </CardTitle>
                  <CardAction>
                    <Button
                      variant="ghost"
                      size="xs"
                      onClick={() => navigator.clipboard?.writeText(item.number).then(() => setCopied(true))}
                    >
                      {copied ? <Check /> : <Copy />}
                      {copied ? "복사됨" : "사건번호 복사"}
                    </Button>
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

              <Card size="sm">
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

              <Card size="sm">
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
            </div>
          </>
        )}
      </SheetContent>
    </Sheet>
  );
}
