package domain

type Entity interface {
    EventList() []Event
    AddAndApplyEvent(event Event) // Метод для того, чтобы сущность применила событие на своё состояние
    CleanEventList()              // Чистит события в репозитории после того, как сущность была сохранена в БД
}
