# История изменений

Все значимые изменения в проекте документируются в этом файле.

Формат основан на [Keep a Changelog](https://keepachangelog.com/ru/1.0.0/),
и этот проект следует [Semantic Versioning](https://semver.org/lang/ru/).

## [Unreleased] - В разработке

## [1.0.0] - 2024-02-20

### Добавлено
- Базовая структура проекта
- FSM (Finite State Machine) для управления состояниями
- Rate Limiter с поддержкой лимитов Telegram API
- Глобальные обработчики команд
- Логирование через zerolog
- Методы для работы с сообщениями:
  - SendMessage, SendMessageRepet
  - EditMessage, EditMessageRepet
  - DeleteMessage, DeleteMessageRepet
  - SendSticker
  - SendPinMessageEvent
  - SendUnPinAllMessageEvent

### Изменено
- Корректный путь модуля в go.mod
- Обновлена документация с примерами использования

[Unreleased]: https://github.com/DEPTH-STRIDA/telegram-bot-api-sfm/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/DEPTH-STRIDA/telegram-bot-api-sfm/releases/tag/v1.0.0 