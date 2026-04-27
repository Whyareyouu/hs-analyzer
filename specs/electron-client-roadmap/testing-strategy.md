# Testing Strategy

## Acceptance Rule

Каждый пункт roadmap принимается только при наличии написанных и проходящих тестов.

Код без тестов не считается завершенным.  
UI без regression/contract coverage для критичных сценариев не считается завершенным.

## Test Levels

### Unit Tests

Использовать для:

- selectors и mappers live payload -> UI model
- deck remaining calculations
- opponent played aggregation
- token/config helpers
- feature-level hooks
- preload API wrappers

### Component Tests

Использовать для:

- card row
- card tooltip
- player deck panel
- opponent played panel
- deck sidebar item
- deck summary cards
- matchup list/ring

### Contract Tests

Использовать для:

- local overlay snapshot schema
- WebSocket message envelope
- preload bridge API
- backend deck endpoints
- renderer decoder/validator behavior

### Integration Tests

Использовать для:

- CLI local transport -> Electron client live updates
- overlay reconnect flow
- main window deck selection -> deck detail render
- backend deck overview -> dashboard widgets
- inspect mode toggle -> tooltip availability

### End-to-End Tests

Использовать для:

- запуск desktop app -> получение live snapshot -> overlay update
- новый draw event -> decrement в player deck panel
- opponent play event -> update в right panel
- открытие приложения -> выбор колоды -> отображение stats/matchups/cards

### Visual Regression Tests

Использовать для:

- overlay layout
- deck sidebar
- deck detail summary
- tooltip layout
- zero-count deck row state

### Performance Tests

Использовать для:

- latency between incoming live event and overlay repaint
- overlay memory footprint during long-running match
- reconnect time after local transport restart
- dashboard rendering with large deck/matchup datasets

## Required Scenario Matrix

1. Клиент стартует без live transport и показывает degraded state.
2. Клиент получает initial live snapshot и рендерит оба overlay списка.
3. Draw event уменьшает remaining count в левой панели.
4. Opponent play event обновляет правую панель без влияния на левую.
5. Новый матч сбрасывает opponent played cards и восстанавливает player deck state для новой игры.
6. Inspect mode включает tooltip interaction.
7. Passive mode не блокирует окно Hearthstone.
8. Dashboard показывает список колод и деталь выбранной колоды.
9. Backend недоступен, но overlay продолжает работать.
10. CLI live transport перезапускается, а Electron клиент переподключается.

## Recommended Tooling

Минимальный рекомендуемый стек тестирования:

- `vitest`
- `@testing-library/react`
- contract fixtures
- Playwright для desktop/Electron e2e

Конкретная реализация runner может быть уточнена позже, но покрытие уровней тестирования обязательно.

## CI Gate

Минимальный gate для клиентской подсистемы:

1. unit tests
2. component tests
3. contract tests
4. critical integration tests

Если в CI для Electron e2e нужны отдельные Windows jobs, это должно быть предусмотрено как часть rollout.

## Documentation Tests

Любое изменение versioned contract examples в этой roadmap должно сопровождаться обновлением fixtures и contract tests.
