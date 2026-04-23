package carddb

// CardDatabase определяет интерфейс доступа к базе карт.
type CardDatabase interface {
	// Load загружает карточную базу из источника данных.
	Load(path string) error

	// GetCard возвращает карту по CardID.
	GetCard(id string) (*CardInfo, bool)

	// GetCardByName возвращает карту по локализованному или базовому имени.
	GetCardByName(name string) (*CardInfo, bool)
}
