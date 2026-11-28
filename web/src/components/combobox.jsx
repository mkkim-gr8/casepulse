import { useState } from "react";
import { Check, ChevronsUpDown } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { cn } from "@/lib/utils";

// 띄어쓰기를 무시하고 비교해, "서울 중앙"으로도 "서울중앙지방법원"을 찾을 수 있게 합니다.
const normalize = (s) => s.replace(/\s+/g, "").toLowerCase();

// 검색해서 고르는 선택 상자입니다(shadcn/ui의 Combobox 구성: Popover + Command).
//   - items: 고를 항목 목록, getKey/getLabel: 항목의 값과 표시 이름, getHint: 오른쪽에 흐리게 보일 보조 정보
//   - value: 현재 선택한 항목의 key, onChange: 항목이 바뀌면 새 key를 전달하는 함수
export function Combobox({
  items,
  value,
  onChange,
  getKey = (x) => x,
  getLabel = (x) => x,
  getHint,
  placeholder = "선택하세요",
  searchPlaceholder = "검색",
  emptyText = "찾는 항목이 없습니다",
  disabled,
  invalid,
  className,
}) {
  const [open, setOpen] = useState(false); // 목록 팝업이 열려 있는지
  const selected = items.find((x) => getKey(x) === value); // 현재 선택된 항목(없으면 undefined)

  // 구조: PopoverTrigger(누르는 버튼) → PopoverContent(떠오르는 상자) 안에 Command(검색창 + 목록).

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          aria-invalid={invalid || undefined}
          disabled={disabled}
          className={cn("w-full justify-between font-normal", !selected && "text-muted-foreground", className)}
        >
          <span className="truncate">{selected ? getLabel(selected) : placeholder}</span>
          <ChevronsUpDown className="opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-(--radix-popover-trigger-width) p-0" align="start">
        {/* filter는 항목마다 호출되어 1이면 보이고 0이면 숨깁니다. 띄어쓰기를 무시하고 포함 여부로 비교합니다. */}
        <Command filter={(itemValue, search) => (normalize(itemValue).includes(normalize(search)) ? 1 : 0)}>
          <CommandInput placeholder={searchPlaceholder} />
          <CommandList>
            <CommandEmpty>{emptyText}</CommandEmpty>
            <CommandGroup>
              {items.map((item) => {
                const key = getKey(item);
                return (
                  <CommandItem
                    key={key}
                    value={getLabel(item)}
                    onSelect={() => {
                      onChange(key);
                      setOpen(false);
                    }}
                  >
                    <Check className={cn(key === value ? "opacity-100" : "opacity-0")} />
                    <span className="truncate">{getLabel(item)}</span>
                    {getHint && <span className="ml-auto text-xs text-muted-foreground">{getHint(item)}</span>}
                  </CommandItem>
                );
              })}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
