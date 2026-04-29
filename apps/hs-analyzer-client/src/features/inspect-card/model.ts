import { useState } from "react";
import type { OverlayDeckCard, OpponentPlayedCard } from "@/shared/model/overlay";

type InspectableCard = OverlayDeckCard | OpponentPlayedCard | null;

export function useCardInspectState() {
  const [card, setCard] = useState<InspectableCard>(null);

  return {
    inspectedCard: card,
    inspectCard: setCard,
    clearInspectCard: () => setCard(null),
  };
}
