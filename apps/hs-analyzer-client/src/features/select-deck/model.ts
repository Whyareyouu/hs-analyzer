import { useEffect, useState } from "react";
import { useServices } from "@/shared/api/services-context";
import type { DashboardBundle, DeckListItem } from "@/shared/model/dashboard";

export function useDeckSelection() {
  const { dashboardClient } = useServices();
  const [decks, setDecks] = useState<DeckListItem[]>([]);
  const [selectedDeckId, setSelectedDeckId] = useState("");
  const [bundle, setBundle] = useState<DashboardBundle | null>(null);
  const [status, setStatus] = useState<"loading" | "ready" | "error">("loading");
  const [stale, setStale] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");

  useEffect(() => {
    let cancelled = false;
    setStatus("loading");
    dashboardClient
      .listDecks()
      .then(({ items, stale: nextStale }) => {
        if (cancelled) {
          return;
        }
        setDecks(items);
        setStale(nextStale);
        if (items[0]) {
          setSelectedDeckId(items[0].deckId);
          return;
        }
        setSelectedDeckId("");
        setBundle(null);
        setStatus("ready");
      })
      .catch((error: Error) => {
        if (cancelled) {
          return;
        }
        setStatus("error");
        setErrorMessage(error.message);
      });

    return () => {
      cancelled = true;
    };
  }, [dashboardClient]);

  useEffect(() => {
    if (!selectedDeckId) {
      setBundle(null);
      return;
    }
    let cancelled = false;
    setStatus("loading");
    dashboardClient
      .loadDeckBundle(selectedDeckId)
      .then((nextBundle) => {
        if (cancelled) {
          return;
        }
        setBundle(nextBundle);
        setStale(nextBundle.stale);
        setStatus("ready");
      })
      .catch((error: Error) => {
        if (cancelled) {
          return;
        }
        setStatus("error");
        setErrorMessage(error.message);
      });

    return () => {
      cancelled = true;
    };
  }, [dashboardClient, selectedDeckId]);

  return {
    decks,
    selectedDeckId,
    selectDeck: setSelectedDeckId,
    bundle,
    status,
    stale,
    errorMessage,
  };
}
