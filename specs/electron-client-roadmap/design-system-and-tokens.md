# Design System and Tokens

## Purpose

Зафиксировать визуальные правила и design-token baseline для `hs-analyzer-client`.

## Visual Direction

UI должен опираться на игровой companion style:

- темная насыщенная база
- полупрозрачные панели
- class-colored accents
- контрастная типографика
- узкие card art strips в списках

Нельзя скатываться в generic white dashboard или случайный default Tailwind look.

## Component Policy

Разрешенная UI component library:

- `shadcn/ui`

Допускаются:

- собственные компоненты поверх `shadcn/ui`
- native CSS/SVG визуализации

Не допускаются:

- дополнительные UI kit
- готовые chart UI libraries
- random mixed design systems

## Token Groups

Все ключевые значения должны быть оформлены как CSS variables.

### Colors

```css
:root {
  --color-bg-app: 220 28% 10%;
  --color-bg-panel: 223 26% 14%;
  --color-bg-panel-strong: 223 24% 18%;
  --color-border-soft: 220 20% 26%;
  --color-text-primary: 210 40% 96%;
  --color-text-muted: 220 18% 72%;
  --color-accent-primary: 216 92% 60%;
  --color-accent-success: 142 76% 45%;
  --color-accent-danger: 0 84% 64%;
  --color-accent-warning: 38 92% 56%;
}
```

### Radius

```css
:root {
  --radius-sm: 6px;
  --radius-md: 10px;
  --radius-lg: 16px;
  --radius-xl: 24px;
}
```

### Spacing

```css
:root {
  --space-1: 4px;
  --space-2: 8px;
  --space-3: 12px;
  --space-4: 16px;
  --space-5: 20px;
  --space-6: 24px;
}
```

### Typography

```css
:root {
  --font-family-base: "Inter", sans-serif;
  --font-size-xs: 12px;
  --font-size-sm: 14px;
  --font-size-md: 16px;
  --font-size-lg: 20px;
  --font-size-xl: 28px;
}
```

Финальный font stack может быть уточнен, но он также должен быть tokenized.

### Motion

```css
:root {
  --motion-fast: 120ms;
  --motion-base: 180ms;
  --motion-slow: 260ms;
  --ease-standard: cubic-bezier(0.2, 0.8, 0.2, 1);
}
```

### Overlay Dimensions

```css
:root {
  --overlay-card-row-width: 300px;
  --overlay-card-row-height: 20px;
  --overlay-card-count-width: 28px;
  --overlay-panel-width: 320px;
}
```

## Tailwind Integration Rule

Tailwind config должен читать tokens как source of truth.  
Компоненты не должны массово использовать hardcoded `#hex`, `px` и arbitrary values, если для этого уже существует token.

## Charting Rule

Поскольку сторонняя chart UI library запрещена, dashboard visualizations должны строиться через:

- native SVG
- CSS conic gradients
- shadcn/ui container components

Это особенно относится к:

- winrate ring
- matchup bars
- summary indicators

## Accessibility Baseline

Даже при game-oriented visual style UI должен соблюдать базовый минимум:

1. достаточный контраст текста
2. keyboard focus visibility
3. понятные hover/focus/selected states
4. tooltip content, читаемый без фоновой картинки

## Acceptance Tests

Required tests:

1. Token existence tests
2. Theme regression tests
3. Visual regression tests для critical dashboard blocks
4. Visual regression tests для overlay rows and tooltip
