# Overlay Window Strategy

## Purpose

Зафиксировать, как именно должен работать overlay runtime поверх Hearthstone, включая позиционирование, режимы интерактивности и layout constraints.

## Window Model

`OverlayWindow` создается как:

- transparent
- frameless
- always-on-top
- skipTaskbar
- visually lightweight

Окно не должно дублировать обычный app chrome и не должно выглядеть как отдельное desktop-приложение поверх игры.

## Positioning Strategy

Overlay позиционируется относительно окна Hearthstone.

Минимальные обязанности main process:

1. Найти окно Hearthstone.
2. Получить его bounds.
3. Синхронизировать bounds overlay window.
4. Учитывать DPI scaling.
5. Реагировать на move/resize/focus изменения.

## Supported Runtime Modes

Поддерживаемый baseline:

- windowed
- borderless-windowed

Если конкретная конфигурация не позволяет надежно показать overlay поверх fullscreen exclusive режима, это должно считаться явным unsupported/degraded case.

## Overlay Layout

### Left Panel

Расположение:

- левый верхний угол окна игры

Назначение:

- моя дека
- оставшееся количество копий

### Right Panel

Расположение:

- правая сторона экрана игры

Назначение:

- все карты, уже разыгранные оппонентом

### Tooltip Layer

Назначение:

- подробная карточка карты
- появляется только в inspect mode

## Layout Tokens

Все ключевые размеры выносятся в CSS variables.

Минимальный набор:

```css
:root {
  --overlay-left-offset: 16px;
  --overlay-top-offset: 16px;
  --overlay-right-offset: 16px;
  --overlay-panel-gap: 12px;
  --overlay-card-row-width: 300px;
  --overlay-card-row-height: 20px;
  --overlay-card-row-radius: 6px;
  --overlay-panel-width: 320px;
  --overlay-tooltip-width: 320px;
}
```

Числа above являются baseline defaults и могут быть скорректированы, но только через tokens.

## Row Rendering Rules

Строка карты должна:

1. занимать токенизированную ширину
2. занимать токенизированную высоту
3. использовать background art strip при наличии картинки
4. показывать count отдельно от названия
5. иметь читаемость поверх яркого арта

Для карт с count `0` допускаются:

- dimming
- desaturation
- low-opacity style

Но строка не должна silently исчезать, если пользователь ожидает видеть весь состав колоды.

## Interactivity Model

### Passive Mode

- pointer events выключены
- overlay не мешает игре
- tooltip недоступен

### Inspect Mode

- pointer events включены
- можно hover/scroll/select карточки
- tooltip/popover работает

Способ входа в inspect mode:

- global shortcut
- tray/menu toggle
- in-app toggle

Конкретный UX control может быть уточнен на этапе реализации, но сам двухрежимный runtime обязателен.

## Tooltip Rules

Tooltip по карточке показывает:

- название
- описание
- изображение, если есть

Если изображения нет:

- показываются название и описание

Tooltip не должен использовать нестабильный raw HTML из карточной базы без нормализации/санитизации.

## Failure and Recovery Behavior

1. IF Hearthstone window не найдено THEN overlay не должен показываться в случайном месте без контекста
2. WHEN игра закрывается THEN overlay SHALL скрываться или переходить в standby state
3. WHEN игра запускается повторно THEN overlay SHALL повторно позиционироваться
4. WHEN live transport временно теряется THEN overlay SHALL показывать degraded state, а не freeze старого UI как будто данные актуальны

## Acceptance Tests

Required tests:

1. Overlay window creation test
2. Overlay positioning test
3. DPI/resolution regression tests
4. Passive/inspect mode transition tests
5. Tooltip interaction tests
6. Window lost/recovered tests
