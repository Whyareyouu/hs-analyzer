import type { OpponentPlayedCard, OverlayDeckCard } from "@/shared/model/overlay";
import { cn } from "@/shared/lib/cn";
import { Panel } from "@/shared/ui/panel";

type PlayerDeckSectionKey = "deck" | "hand" | "other";

interface PlayerDeckSection {
  key: PlayerDeckSectionKey;
  title: string;
  cards: PlayerDeckSectionCard[];
  total: number;
}

interface PlayerDeckSectionCard {
  card: OverlayDeckCard;
  copies: number;
}

interface PlayerOtherSectionCard {
  card: OpponentPlayedCard;
  copies: number;
}

const SECTION_TITLES: Record<PlayerDeckSectionKey, string> = {
  deck: "In deck",
  hand: "In hand",
  other: "Other",
};

const RARITY_BACKGROUNDS: Record<string, string> = {
  LEGENDARY: "linear-gradient(90deg, rgba(14, 18, 31, 0.98) 0%, rgba(36, 23, 14, 0.96) 58%, rgba(201, 133, 35, 0.92) 100%)",
  EPIC: "linear-gradient(90deg, rgba(14, 18, 31, 0.98) 0%, rgba(32, 20, 49, 0.96) 58%, rgba(124, 73, 214, 0.9) 100%)",
  RARE: "linear-gradient(90deg, rgba(14, 18, 31, 0.98) 0%, rgba(17, 31, 56, 0.96) 58%, rgba(33, 115, 216, 0.92) 100%)",
  COMMON: "linear-gradient(90deg, rgba(14, 18, 31, 0.98) 0%, rgba(26, 33, 49, 0.96) 58%, rgba(101, 111, 131, 0.86) 100%)",
  FREE: "linear-gradient(90deg, rgba(14, 18, 31, 0.98) 0%, rgba(24, 31, 46, 0.96) 58%, rgba(79, 92, 115, 0.84) 100%)",
  UNKNOWN: "linear-gradient(90deg, rgba(14, 18, 31, 0.98) 0%, rgba(24, 31, 46, 0.96) 58%, rgba(66, 77, 100, 0.84) 100%)",
};

const TYPE_BACKGROUNDS: Record<string, string> = {
  SPELL: "linear-gradient(90deg, rgba(14, 18, 31, 0.98) 0%, rgba(17, 44, 62, 0.96) 58%, rgba(46, 150, 195, 0.9) 100%)",
  WEAPON: "linear-gradient(90deg, rgba(14, 18, 31, 0.98) 0%, rgba(48, 29, 18, 0.96) 58%, rgba(168, 97, 48, 0.88) 100%)",
  LOCATION: "linear-gradient(90deg, rgba(14, 18, 31, 0.98) 0%, rgba(19, 49, 39, 0.96) 58%, rgba(42, 149, 112, 0.88) 100%)",
};

export function PlayerDeckPanel({
  cards,
  otherCards,
  inspectMode,
  onInspect,
}: {
  cards: OverlayDeckCard[];
  otherCards: OpponentPlayedCard[];
  inspectMode: boolean;
  onInspect(card: OverlayDeckCard | OpponentPlayedCard | null, anchorRect?: DOMRect): void;
}) {
  const sections = buildPlayerDeckSections(cards);
  const otherSection = buildPlayerOtherSection(otherCards);

  return (
    <Panel className="flex w-[var(--overlay-panel-width)] flex-col p-3" style={{ maxHeight: "calc(100vh - (var(--overlay-top-offset) * 2))" }}>
      <div className="min-h-0 space-y-3 overflow-x-hidden overflow-y-auto pr-1">
        {cards.length === 0 && otherCards.length === 0 ? <div className="text-xs text-app-muted">Ждём выбранную колоду из Decks.log.</div> : null}
        {cards.length > 0
          ? sections.map((section) => (
              <section key={section.key} aria-label={`${section.title} ${section.total}`}>
                <div className="mb-1.5 flex items-center justify-between text-[10px] font-semibold uppercase tracking-[0.18em] text-app-muted">
                  <span>
                    {section.title} ({section.total})
                  </span>
                </div>
                <div className="space-y-1">
                  {section.cards.map(({ card, copies }) => (
                    <PlayerDeckCardRow
                      key={`${section.key}:${card.cardId}`}
                      card={card}
                      copies={copies}
                      sectionKey={section.key}
                      inspectMode={inspectMode}
                      onInspect={onInspect}
                    />
                  ))}
                </div>
              </section>
            ))
          : null}
        <section aria-label={`${otherSection.title} ${otherSection.total}`}>
          <div className="mb-1.5 flex items-center justify-between text-[10px] font-semibold uppercase tracking-[0.18em] text-app-muted">
            <span>
              {otherSection.title} ({otherSection.total})
            </span>
          </div>
          <div className="space-y-1">
            {otherSection.cards.map(({ card, copies }) => (
              <PlayerOtherCardRow key={`other:${card.cardId}`} card={card} copies={copies} inspectMode={inspectMode} onInspect={onInspect} />
            ))}
          </div>
        </section>
      </div>
    </Panel>
  );
}

function PlayerDeckCardRow({
  card,
  copies,
  sectionKey,
  inspectMode,
  onInspect,
}: {
  card: OverlayDeckCard;
  copies: number;
  sectionKey: PlayerDeckSectionKey;
  inspectMode: boolean;
  onInspect(card: OverlayDeckCard | OpponentPlayedCard | null, anchorRect?: DOMRect): void;
}) {
  return (
    <button
      type="button"
      onMouseEnter={(event) => inspectMode && onInspect(card, event.currentTarget.getBoundingClientRect())}
      onMouseLeave={() => onInspect(null)}
      onFocus={(event) => inspectMode && onInspect(card, event.currentTarget.getBoundingClientRect())}
      onBlur={() => onInspect(null)}
      className={cn(
        "grid w-[var(--overlay-card-row-width)] grid-cols-[42px_1fr_var(--overlay-card-count-width)] items-center overflow-hidden rounded-[var(--overlay-card-row-radius)] border text-left shadow-panel transition-[filter,opacity,transform] duration-base focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-app-accent/60",
        sectionKey === "hand" ? "border-amber-500/75 brightness-110 shadow-[0_0_0_1px_rgba(245,158,11,0.22)]" : "border-slate-800/80",
        sectionKey === "other" ? "border-slate-900/90 opacity-70 grayscale" : "opacity-100",
      )}
      style={{
        minHeight: "var(--overlay-card-row-height)",
        backgroundImage: buildRowBackground(card, sectionKey),
        backgroundPosition: buildRowBackgroundPosition(card),
        backgroundSize: buildRowBackgroundSize(card),
      }}
    >
      <span
        className={cn(
          "grid h-full place-items-center bg-black/55 text-xs font-bold",
          sectionKey === "hand" ? "text-amber-100" : "text-app-accent",
          sectionKey === "other" ? "text-slate-400" : "",
        )}
      >
        {card.manaCost}
      </span>
      <span className={cn("truncate px-3 text-xs font-semibold drop-shadow-[0_1px_1px_rgba(0,0,0,0.7)]", sectionKey === "other" ? "text-slate-300" : "text-app-text")}>
        {card.name}
      </span>
      <span
        aria-label={buildCopiesLabel(card, copies)}
        className={cn(
          "grid h-full place-items-center bg-black/55 text-[11px] font-black",
          sectionKey === "hand" ? "text-amber-100" : "text-app-text",
          sectionKey === "other" ? "text-slate-400" : "",
        )}
      >
        {formatCopies(card, copies)}
      </span>
    </button>
  );
}

function PlayerOtherCardRow({
  card,
  copies,
  inspectMode,
  onInspect,
}: {
  card: OpponentPlayedCard;
  copies: number;
  inspectMode: boolean;
  onInspect(card: OverlayDeckCard | OpponentPlayedCard | null, anchorRect?: DOMRect): void;
}) {
  return (
    <button
      type="button"
      onMouseEnter={(event) => inspectMode && onInspect(card, event.currentTarget.getBoundingClientRect())}
      onMouseLeave={() => onInspect(null)}
      onFocus={(event) => inspectMode && onInspect(card, event.currentTarget.getBoundingClientRect())}
      onBlur={() => onInspect(null)}
      className="grid w-[var(--overlay-card-row-width)] grid-cols-[42px_1fr_var(--overlay-card-count-width)] items-center overflow-hidden rounded-[var(--overlay-card-row-radius)] border border-slate-900/90 text-left opacity-70 shadow-panel grayscale transition-[filter,opacity,transform] duration-base focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-app-accent/60"
      style={{
        minHeight: "var(--overlay-card-row-height)",
        backgroundImage: buildOtherCardBackground(card),
        backgroundPosition: card.image ? "center 25%" : "center",
        backgroundSize: "cover",
      }}
    >
      <span className="grid h-full place-items-center bg-black/55 text-xs font-bold text-slate-400">{card.manaCost}</span>
      <span className="truncate px-3 text-xs font-semibold text-slate-300 drop-shadow-[0_1px_1px_rgba(0,0,0,0.7)]">{card.name}</span>
      <span aria-label={`${copies} copies`} className="grid h-full place-items-center bg-black/55 text-[11px] font-black text-slate-400">
        {copies}
      </span>
    </button>
  );
}

function buildPlayerDeckSections(cards: OverlayDeckCard[]): PlayerDeckSection[] {
  const keys: PlayerDeckSectionKey[] = ["deck", "hand"];

  return keys.map((key) => {
    const sectionCards = cards
      .map((card) => ({
        card,
        copies: getSectionCopies(card, key),
      }))
      .filter(({ copies }) => copies > 0);

    return {
      key,
      title: SECTION_TITLES[key],
      cards: sectionCards,
      total: sectionCards.reduce((sum, current) => sum + current.copies, 0),
    };
  });
}

function getSectionCopies(card: OverlayDeckCard, sectionKey: PlayerDeckSectionKey): number {
  switch (sectionKey) {
    case "deck":
      return Math.max(card.remainingCopies, 0);
    case "hand":
      return Math.max(card.drawnCopies - card.playedCopies, 0);
    case "other":
      return 0;
  }
}

function buildPlayerOtherSection(cards: OpponentPlayedCard[]) {
  const sectionCards: PlayerOtherSectionCard[] = cards
    .map((card) => ({
      card,
      copies: Math.max(card.playCount, 0),
    }))
    .filter(({ copies }) => copies > 0);

  return {
    key: "other" as const,
    title: SECTION_TITLES.other,
    cards: sectionCards,
    total: sectionCards.reduce((sum, current) => sum + current.copies, 0),
  };
}

function buildRowBackground(card: OverlayDeckCard, sectionKey: PlayerDeckSectionKey): string {
  const base = getCardBackground(card);
  const image = card.image ? `url(${card.image})` : "";

  if (sectionKey === "hand") {
    const highlight = "linear-gradient(90deg, rgba(58, 34, 10, 0.88), rgba(208, 134, 24, 0.44) 82%, rgba(255, 205, 96, 0.62) 100%)";
    return image ? `${highlight}, ${getCardImageTint(card)}, ${image}` : `${highlight}, ${base}`;
  }

  if (sectionKey === "other") {
    const disabled = "linear-gradient(90deg, rgba(4, 7, 15, 0.98), rgba(15, 23, 42, 0.92))";
    return image ? `${disabled}, linear-gradient(90deg, rgba(2, 6, 23, 0.82), rgba(2, 6, 23, 0.72)), ${image}` : disabled;
  }

  return image ? `${getCardImageTint(card)}, ${image}` : base;
}

function getCardBackground(card: OverlayDeckCard): string {
  const rarity = card.rarity.toUpperCase();
  const cardType = card.cardType.toUpperCase();

  return RARITY_BACKGROUNDS[rarity] ?? TYPE_BACKGROUNDS[cardType] ?? RARITY_BACKGROUNDS.UNKNOWN;
}

function getCardImageTint(card: OverlayDeckCard): string {
  const rarity = card.rarity.toUpperCase();
  const cardType = card.cardType.toUpperCase();

  switch (rarity) {
    case "LEGENDARY":
      return "linear-gradient(90deg, rgba(5, 8, 17, 0.86), rgba(24, 17, 12, 0.72) 58%, rgba(255, 183, 55, 0.32) 100%)";
    case "EPIC":
      return "linear-gradient(90deg, rgba(5, 8, 17, 0.86), rgba(20, 14, 37, 0.72) 58%, rgba(156, 104, 255, 0.3) 100%)";
    case "RARE":
      return "linear-gradient(90deg, rgba(5, 8, 17, 0.86), rgba(13, 26, 51, 0.72) 58%, rgba(68, 154, 255, 0.3) 100%)";
    default:
      if (cardType === "SPELL") {
        return "linear-gradient(90deg, rgba(5, 8, 17, 0.86), rgba(11, 33, 46, 0.72) 58%, rgba(73, 188, 233, 0.28) 100%)";
      }
      if (cardType === "WEAPON") {
        return "linear-gradient(90deg, rgba(5, 8, 17, 0.86), rgba(36, 22, 14, 0.72) 58%, rgba(195, 116, 58, 0.24) 100%)";
      }
      return "linear-gradient(90deg, rgba(5, 8, 17, 0.88), rgba(14, 19, 31, 0.76) 58%, rgba(117, 133, 161, 0.2) 100%)";
  }
}

function formatCopies(_card: OverlayDeckCard, copies: number): string {
  return String(copies);
}

function buildCopiesLabel(_card: OverlayDeckCard, copies: number): string {
  return `${copies} copies`;
}

function buildOtherCardBackground(card: OpponentPlayedCard): string {
  const disabled = "linear-gradient(90deg, rgba(4, 7, 15, 0.98), rgba(15, 23, 42, 0.92))";
  if (!card.image) {
    return disabled;
  }
  return `${disabled}, linear-gradient(90deg, rgba(2, 6, 23, 0.82), rgba(2, 6, 23, 0.72)), url(${card.image})`;
}

function buildRowBackgroundPosition(card: OverlayDeckCard): string {
  return card.image ? "center 25%" : "center";
}

function buildRowBackgroundSize(card: OverlayDeckCard): string {
  return card.image ? "cover" : "cover";
}
