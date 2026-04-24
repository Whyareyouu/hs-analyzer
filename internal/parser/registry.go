package parser

import "fmt"

// Registry хранит зарегистрированные парсеры по типу лога.
type Registry struct {
	parsers map[LogType]Parser
}

func NewRegistry() *Registry {
	return &Registry{
		parsers: make(map[LogType]Parser),
	}
}

func (r *Registry) Register(parser Parser) error {
	if r == nil {
		return fmt.Errorf("parser registry is nil")
	}
	if parser == nil {
		return fmt.Errorf("parser is nil")
	}

	logType := parser.LogType()
	if _, exists := r.parsers[logType]; exists {
		return fmt.Errorf("parser already registered for log type %s", logType)
	}

	r.parsers[logType] = parser
	return nil
}

func (r *Registry) MustRegister(parser Parser) {
	if err := r.Register(parser); err != nil {
		panic(err)
	}
}

func (r *Registry) Get(logType LogType) (Parser, bool) {
	if r == nil {
		return nil, false
	}

	parser, ok := r.parsers[logType]
	return parser, ok
}

func (r *Registry) List() []Parser {
	if r == nil {
		return nil
	}

	result := make([]Parser, 0, len(r.parsers))
	for _, parser := range r.parsers {
		result = append(result, parser)
	}
	return result
}
