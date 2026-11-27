import { useEffect } from "react";
import { Maximize2, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Kbd } from "@/components/ui/kbd";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { CaseInfoCard, CourtCard, PartiesCard } from "./CaseDetails.jsx";
import { MOD_KEY } from "./keyboard.js";

// 사건목록 탭 안 오른쪽에서 열리는 사건 미리보기 패널입니다.
//
// 화면 전체를 덮는 대화상자(Sheet)가 아니라 사건목록 탭 영역 안에만 뜹니다.
// 따라서 메뉴줄·탭 막대를 가리지 않고, 목록도 어둡게 덮지 않아 다른 행을 누르면 내용만 바로 바뀝니다.
// 부모(CaseListPage)는 position: relative여야 합니다.
//   - active: 사건목록 탭이 보이는 중인지. 보일 때만 Esc로 닫힙니다.
//   - onOpenTab: "탭으로 열기" 버튼. 패널을 닫고 같은 사건을 사건 탭(전체 화면)으로 엽니다.
export default function CaseDrawer({ item, active, onClose, onOpenTab }) {
  useEffect(() => {
    if (!item || !active) return;
    const onKey = (e) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [item, active, onClose]);

  if (!item) return null;
  return (
    <aside
      aria-label={`${item.number} 미리보기`}
      className="absolute inset-y-0 right-0 z-20 flex w-full max-w-[480px] flex-col border-l bg-background shadow-xl duration-200 animate-in fade-in-0 slide-in-from-right-10"
    >
      <header className="flex items-start gap-2 border-b p-4">
        <div className="flex min-w-0 flex-1 flex-col gap-1">
          <h2 className="font-mono text-lg font-semibold">{item.number}</h2>
          <p className="truncate text-sm text-muted-foreground">{item.title}</p>
        </div>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="icon-sm"
              aria-label="탭으로 열기"
              onClick={() => {
                onOpenTab(item);
                onClose();
              }}
            >
              <Maximize2 />
            </Button>
          </TooltipTrigger>
          <TooltipContent side="bottom">
            탭으로 열기 <span className="text-muted-foreground">(목록에서 더블클릭 또는</span>
            <Kbd>{MOD_KEY}</Kbd>
            <span className="text-muted-foreground">+클릭)</span>
          </TooltipContent>
        </Tooltip>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button variant="ghost" size="icon-sm" aria-label="미리보기 닫기" onClick={onClose}>
              <X />
            </Button>
          </TooltipTrigger>
          <TooltipContent side="bottom">
            닫기 <Kbd>Esc</Kbd>
          </TooltipContent>
        </Tooltip>
      </header>
      <div className="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto p-4">
        <CaseInfoCard item={item} />
        <PartiesCard item={item} />
        <CourtCard item={item} />
      </div>
    </aside>
  );
}
