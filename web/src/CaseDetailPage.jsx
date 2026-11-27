import { FileText } from "lucide-react";
import { CaseInfoCard, CopyNumberButton, CourtCard, PartiesCard } from "./CaseDetails.jsx";
import CategoryBadge from "./CategoryBadge.jsx";

// 사건 탭의 내용: 한 사건을 화면 전체 폭으로 봅니다.
// 위에는 사건번호·사건명·종류를 크게 두고, 아래에 사건·법원 카드와 당사자 표를 나란히 둡니다.
// 사건 수정이나 다른 서비스(진행내역 등) 연결은 나중에 머리 오른쪽(actions 자리)에 붙이면 됩니다.
export default function CaseDetailPage({ item }) {
  return (
    <div className="h-full overflow-y-auto p-4">
      <div className="mx-auto flex max-w-6xl flex-col gap-4">
        <header className="flex flex-wrap items-start gap-3 rounded-xl bg-card p-5 ring-1 ring-foreground/10">
          <div className="flex size-10 items-center justify-center rounded-lg bg-muted">
            <FileText className="size-5 text-muted-foreground" />
          </div>
          <div className="flex min-w-0 flex-1 flex-col gap-1">
            <div className="flex flex-wrap items-center gap-2">
              <h1 className="font-mono text-xl font-semibold">{item.number}</h1>
              <CategoryBadge category={item.category} />
            </div>
            <p className="text-muted-foreground">
              {item.title} · {item.court.name} {item.court.panel}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <CopyNumberButton number={item.number} />
          </div>
        </header>

        <div className="grid gap-4 lg:grid-cols-[minmax(0,2fr)_minmax(0,3fr)]">
          <div className="flex flex-col gap-4">
            <CaseInfoCard item={item} size="default" />
            <CourtCard item={item} size="default" />
          </div>
          <PartiesCard item={item} size="default" />
        </div>
      </div>
    </div>
  );
}
