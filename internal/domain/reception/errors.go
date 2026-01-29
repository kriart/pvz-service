package reception

import "errors"

var (
	ErrReceptionAlreadyOpen = errors.New("прошлая приёмка еще не закрыта")
	ErrNoOpenReception      = errors.New("нет открытой приёмки")
	ErrReceptionClosed      = errors.New("приёмка уже закрыта")
	ErrNoProducts           = errors.New("в приёмке нет товаров")
)
