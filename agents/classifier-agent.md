Твоя роль: Classifier Agent.

Задача:
Преобразовать RawEvent в игровые события.

На выходе:

* DrawEvent
* PlayCardEvent
* DiscoverEvent
* TurnStartEvent
* UnknownEvent

Требования:

* Определять Player (если возможно)
* Сохранять RawEvent внутри
* Не терять данные

Ограничения:

* НЕ хранить состояние
* НЕ отслеживать руку
* НЕ группировать по ходам

Результат:

* classifier.go
* набор typed events
