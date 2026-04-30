import type { OverlayDeckCard } from "@/shared/model/overlay";
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

const SECTION_TITLES: Record<PlayerDeckSectionKey, string> = {
  deck: "In deck",
  hand: "In hand",
  other: "Other",
};

const RARITY_BACKGROUNDS: Record<string, string> = {
  LEGENDARY: "linear-gradient(90deg, rgba(123, 59, 9, 0.96), rgba(202, 138, 4, 0.62))",
  EPIC: "linear-gradient(90deg, rgba(76, 29, 149, 0.96), rgba(168, 85, 247, 0.56))",
  RARE: "linear-gradient(90deg, rgba(7, 89, 133, 0.96), rgba(59, 130, 246, 0.5))",
  COMMON: "linear-gradient(90deg, rgba(30, 41, 59, 0.96), rgba(71, 85, 105, 0.58))",
  FREE: "linear-gradient(90deg, rgba(15, 23, 42, 0.96), rgba(51, 65, 85, 0.6))",
  UNKNOWN: "linear-gradient(90deg, rgba(12, 18, 31, 0.96), rgba(30, 41, 59, 0.62))",
};

const TYPE_BACKGROUNDS: Record<string, string> = {
  SPELL: "linear-gradient(90deg, rgba(14, 116, 144, 0.94), rgba(56, 189, 248, 0.42))",
  WEAPON: "linear-gradient(90deg, rgba(88, 28, 135, 0.94), rgba(192, 132, 252, 0.42))",
  LOCATION: "linear-gradient(90deg, rgba(4, 120, 87, 0.94), rgba(52, 211, 153, 0.38))",
};

export function PlayerDeckPanel({
  cards,
  inspectMode,
  onInspect,
}: {
  cards: OverlayDeckCard[];
  inspectMode: boolean;
  onInspect(card: OverlayDeckCard | null): void;
}) {
  const sections = buildPlayerDeckSections(cards);

  return (
    <Panel className="w-[var(--overlay-panel-width)] p-3">
      <div className="space-y-3">
        {cards.length === 0 ? <div className="text-xs text-app-muted">Ждём выбранную колоду из Decks.log.</div> : null}
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
  onInspect(card: OverlayDeckCard | null): void;
}) {
  return (
    <button
      type="button"
      onMouseEnter={() => inspectMode && onInspect(card)}
      onMouseLeave={() => onInspect(null)}
      onFocus={() => inspectMode && onInspect(card)}
      onBlur={() => onInspect(null)}
      className={cn(
        "grid w-[var(--overlay-card-row-width)] grid-cols-[34px_1fr_var(--overlay-card-count-width)] items-center overflow-hidden rounded-[var(--overlay-card-row-radius)] border text-left shadow-panel transition-[filter,opacity,transform] duration-base focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-app-accent/60",
        sectionKey === "hand" ? "border-amber-200/50 brightness-110" : "border-app-border/60",
        sectionKey === "other" ? "border-slate-700/70 opacity-70 grayscale" : "opacity-100",
      )}
      style={{
        minHeight: "var(--overlay-card-row-height)",
        backgroundImage: buildRowBackground(card, sectionKey),
        backgroundPosition: "center",
        backgroundSize: "cover",
      }}
    >
      <span
        className={cn(
          "grid h-full place-items-center bg-black/40 text-xs font-bold",
          sectionKey === "hand" ? "text-amber-100" : "text-app-accent",
          sectionKey === "other" ? "text-slate-400" : "",
        )}
      >
        {card.manaCost}
      </span>
      <span className={cn("truncate px-2 text-xs font-semibold", sectionKey === "other" ? "text-slate-300" : "text-app-text")}>
        {card.name}
      </span>
      <span
        aria-label={buildCopiesLabel(card, copies)}
        className={cn(
          "grid h-full place-items-center bg-black/35 text-[11px] font-black",
          sectionKey === "hand" ? "text-amber-100" : "text-app-text",
          sectionKey === "other" ? "text-slate-400" : "",
        )}
      >
        {formatCopies(card, copies)}
      </span>
    </button>
  );
}

function buildPlayerDeckSections(cards: OverlayDeckCard[]): PlayerDeckSection[] {
  const keys: PlayerDeckSectionKey[] = ["deck", "hand", "other"];

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
      return Math.max(card.playedCopies, 0);
  }
}

function buildRowBackground(card: OverlayDeckCard, sectionKey: PlayerDeckSectionKey): string {
  const base = getCardBackground(card);
  const image = card.image ? `url(${card.image})` : "";

  if (sectionKey === "hand") {
    const highlight = "linear-gradient(90deg, rgba(251, 191, 36, 0.48), rgba(217, 119, 6, 0.18))";
    return image ? `${highlight}, ${getCardImageTint(card)}, ${image}` : `${highlight}, ${base}`;
  }

  if (sectionKey === "other") {
    const disabled = "linear-gradient(90deg, rgba(2, 6, 23, 0.96), rgba(15, 23, 42, 0.86))";
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
      return "linear-gradient(90deg, rgba(120, 53, 15, 0.9), rgba(251, 191, 36, 0.24))";
    case "EPIC":
      return "linear-gradient(90deg, rgba(76, 29, 149, 0.9), rgba(192, 132, 252, 0.22))";
    case "RARE":
      return "linear-gradient(90deg, rgba(14, 116, 144, 0.9), rgba(96, 165, 250, 0.2))";
    default:
      if (cardType === "SPELL") {
        return "linear-gradient(90deg, rgba(14, 116, 144, 0.86), rgba(56, 189, 248, 0.18))";
      }
      return "linear-gradient(90deg, rgba(3, 7, 18, 0.82), rgba(3, 7, 18, 0.28))";
  }
}

function formatCopies(card: OverlayDeckCard, copies: number): string {
  if (card.isLegendary) {
    return "*";
  }
  return copies > 1 ? String(copies) : "";
}

function buildCopiesLabel(card: OverlayDeckCard, copies: number): string {
  return card.isLegendary ? "legendary card" : `${copies} copies`;
}
