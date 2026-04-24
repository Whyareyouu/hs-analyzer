package formatter

import "fmt"

// Registry хранит зарегистрированные formatter'ы.
type Registry struct {
	formatters map[string]Formatter
}

// NewRegistry создает новый registry formatter'ов.
func NewRegistry() *Registry {
	return &Registry{
		formatters: make(map[string]Formatter),
	}
}

func (r *Registry) Register(formatter Formatter) error {
	if r == nil {
		return fmt.Errorf("formatter registry is nil")
	}
	if formatter == nil {
		return fmt.Errorf("formatter is nil")
	}

	name := formatter.Name()
	if _, exists := r.formatters[name]; exists {
		return fmt.Errorf("formatter already registered: %s", name)
	}

	r.formatters[name] = formatter
	return nil
}

func (r *Registry) MustRegister(formatter Formatter) {
	if err := r.Register(formatter); err != nil {
		panic(err)
	}
}

func (r *Registry) Get(name string) (Formatter, bool) {
	if r == nil {
		return nil, false
	}

	formatter, ok := r.formatters[name]
	return formatter, ok
}
