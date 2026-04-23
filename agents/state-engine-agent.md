На входе: RawEvent
На выходе: типизированные игровые события

Задача:
Создать слой классификации событий Hearthstone.

Нужно:

* Превратить RawEvent в:

  * DrawEvent
  * PlayCardEvent
  * DiscoverEvent
  * TurnStartEvent
  * и т.д.

Требования:

* Каждое событие должно иметь PlayerID
* Если данных недостаточно — помечать как UnknownEvent
* Не терять Raw данные

Сначала:

1. Опиши модель событий
2. Затем реализуй classifier.go
