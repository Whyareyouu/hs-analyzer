# Dashboard Visual Reference

## Purpose

Этот документ формализует dashboard design reference, предоставленный для `hs-analyzer-client`.

Оригинальная картинка является входным артефактом обсуждения, но после фиксации этого документа именно markdown-описание в репозитории считается source of truth для реализации layout, composition и visual priorities.

## Reference Intent

Главный экран desktop-клиента должен выглядеть как игровой companion dashboard, а не как generic admin panel.

Визуальный baseline:

- темный цельный background
- крупные rounded panels
- мягкие внутренние границы
- явное визуальное разделение sidebar и main content
- акцентные class icons
- крупный заголовок выбранной колоды
- компактные, но хорошо читаемые карточки в grid/list layout

## Main Layout

Экран делится на две основные зоны:

1. Левая вертикальная sidebar-колонка со списком колод
2. Правая контентная зона с detail view выбранной колоды

Sidebar должна быть визуально тяжелее обычного nav rail и восприниматься как полноценная deck browser panel.

## Sidebar Reference

### Purpose

Sidebar показывает все колоды пользователя.

### Visual Structure

Верх sidebar содержит:

- заголовок `Мои колоды`
- компактный badge с общим количеством колод

Ниже идет вертикальный список deck items.

### Deck Item Composition

Каждый deck item содержит:

- class/deck icon слева
- название колоды
- winrate
- запись вида `wins-losses`

Выбранный item должен иметь:

- более яркий фон
- accent border/outline
- усиленное выделение без агрессивного glow

Не выбранные items должны оставаться читаемыми, но визуально слабее selected state.

## Main Header Reference

Вверху content area показывается header выбранной колоды:

- крупная class/deck icon
- название колоды
- winrate и record строкой ниже

Header не должен сливаться с summary cards ниже. Он выступает как отдельная hero-зона deck detail.

## Summary Row Reference

После header идет ряд из трех summary panels:

1. `Общая статистика`
   - круговой winrate indicator
   - процент winrate в центре
   - wins/losses справа внутри панели

2. `Матчи`
   - total matches
   - winrate summary

3. `Лучшие матапы`
   - список лучших matchup entries
   - class icons слева
   - colored winrate справа

Эти панели должны быть равновесными по высоте и визуально составлять единый ряд.

## Winrate Ring Reference

Winrate indicator должен быть реалистично близок к референсу:

- большой круговой элемент
- числовой winrate в центре
- подпись ниже в центре

Допустимая реализация:

- SVG circle
- conic-gradient + mask

Сторонняя chart library для этого не допускается.

## Matchup List Reference

Список лучших матчапов должен быть компактным и сканируемым:

- иконка класса
- текст `Против <класс/архетип>`
- цветной процент справа

Смысловая окраска:

- хорошие значения ближе к зеленому
- средние к желто-оранжевому
- плохие к красному

Даже если в конкретном блоке показываются только лучшие matchup-и, color scale должна быть общей и последовательной по приложению.

## Deck Cards Section Reference

Нижняя крупная панель содержит список карт выбранной колоды.

### Section Header

Заголовок:

- `Карты в колоде`

### Card Item Composition

Каждая строка/ячейка карты должна показывать:

- mana cost badge слева
- название карты
- copies справа
- optional rarity/legendary indicator

Секция не должна быть обычной HTML-таблицей. Она должна выглядеть как dense premium card matrix.

### Grid Behavior

На desktop baseline cards section ближе к multi-column grid, чем к одной длинной колонке.

Приоритет:

1. высокая плотность данных
2. быстрая читаемость
3. ровные колонки
4. сохранение воздуха между карточками

## Panel Language

Все основные панели dashboard должны использовать единый визуальный язык:

- rounded corners
- soft border
- subtle gradient/darken fill
- consistent internal padding
- consistent title typography

Недопустимо, чтобы sidebar, summary cards и deck cards section выглядели как UI из разных приложений.

## Color and Contrast Reference

Референс указывает на:

- темный сине-графитовый фон
- синий primary accent
- зеленый accent для positive metrics
- красный accent для losses
- золотистый accent для special/legendary card markers

Это не означает pixel-perfect copy, но отклонение должно оставаться внутри того же visual family.

## Acceptance Implications

Реализация dashboard не считается принятой, если:

1. sidebar потеряла выраженный selected state
2. summary row перестала быть трехпанельной hero summary
3. winrate ring заменен на безликий текст без visual weight
4. cards section выглядит как generic table/list без deck-builder характера
5. итоговый экран визуально оторвался от этого reference baseline
