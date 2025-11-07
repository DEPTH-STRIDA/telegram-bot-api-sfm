package main

import (
	"log"
	"tgfsm"
	"tgfsm/events"
)

func main() {

	token := "8078686403:AAHu0EQeR6vb35LcFO_WgngbAyZ9XUyrvwQ"

	// Создаем состояния для простого слайдера
	states, err := events.NewSimpleSliderEvent(
		events.WithSimpleSliderMessageTriggers("/start", "/slider"),
		events.WithSimpleSliderTexts(
			"Текст 1: Это первый элемент слайдера",
			"Текст 2: Это второй элемент слайдера",
			"Текст 3: Это третий элемент слайдера",
			"Текст 4: Это четвертый элемент слайдера",
			"Текст 5: Это пятый элемент слайдера",
		),
		events.WithSimpleSliderPrevButtonText("◀️ Назад"),
		events.WithSimpleSliderNextButtonText("Вперед ▶️"),
		// Дополнительные кнопки на второй строке
		events.WithSimpleSliderAdditionalButtons(
			events.SimpleSliderButton{Text: "Главное меню", Callback: "menu"},
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	bot, err := tgfsm.NewBot(token, tgfsm.WithStates(states))
	if err != nil {
		log.Fatal(err)
	}

	bot.Start(0, 10)

	select {}
}
