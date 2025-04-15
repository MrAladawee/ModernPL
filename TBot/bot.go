package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type EmployeeResponse struct {
	Success   bool       `json:"success"`
	Employees []Employee `json:"employees"`
}

type Employee struct {
	EmployeeID int    `json:"employee_id"`
	LastName   string `json:"lastname"`
	FirstName  string `json:"firstname"`
	MiddleName string `json:"middlename"`
	Email      string `json:"email"`
}

type ScheduleResponse struct {
	Success  bool      `json:"success"`
	Subjects []Subject `json:"subjects"`
}

type Subject struct {
	ID                    string `json:"id"`
	Semester              int    `json:"semester"`
	Year                  int    `json:"year"`
	SubjectName           string `json:"subject_name"`
	SubjectID             int    `json:"subject_id"`
	StartDaySchedule      string `json:"start_day_schedule"`
	FinishDaySchedule     string `json:"finish_day_schedule"`
	DayWeekSchedule       int    `json:"day_week_schedule"`
	TypeWeekSchedule      int    `json:"type_week_schedule"`
	NoteSchedule          string `json:"note_schedule"`
	TotalTimeSchedule     string `json:"total_time_schedule"`
	BeginTimeSchedule     string `json:"begin_time_schedule"`
	EndTimeSchedule       string `json:"end_time_schedule"`
	TeacherID             int    `json:"teacher_id"`
	TeacherLastName       string `json:"teacher_lastname"`
	TeacherFirstName      string `json:"teacher_firstname"`
	TeacherMiddleName     string `json:"teacher_middlename"`
	NumAuditoriumSchedule string `json:"num_auditorium_schedule"`
	BuildingName          string `json:"building_name"`
	BuildingID            string `json:"building_id"`
	GroupList             string `json:"group_list"`
	SubjectKindName       string `json:"subject_kind_name"`
}

type UserData struct {
	FullName  string
	TeacherID int
	Email     string
}

var userData = make(map[int64]UserData)

// Функция для получения данных преподавателя по ФИО
func getEmployee(fullName string) (Employee, error) {

	url := "https://auth.kpfu.tyuop.ru/api/v1/employees?q=" + fullName

	resp, err := http.Get(url)

	if err != nil {
		return Employee{}, err
	}

	defer resp.Body.Close()

	var er EmployeeResponse

	if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
		return Employee{}, err
	}

	if !er.Success || len(er.Employees) == 0 {
		return Employee{}, fmt.Errorf("Сотрудник не найден")
	}

	// Берём первого найденного преподавателя
	return er.Employees[0], nil
}

// Функция для получения расписания
func getSchedule(teacherID int) ([]Subject, error) {

	url := fmt.Sprintf("https://auth.kpfu.tyuop.ru/api/v1/employees/%d/schedule", teacherID)
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var sr ScheduleResponse

	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, err
	}

	if !sr.Success {
		return nil, fmt.Errorf("Ошибка получения расписания")
	}

	return sr.Subjects, nil
}

// Функция для фильтрации расписания по дню недели (1 — понедельник, 2 — вторник, …, 7 — воскресенье)
func filterScheduleByDay(subjects []Subject, day int) []Subject {

	var result []Subject

	for _, subj := range subjects {
		if subj.DayWeekSchedule == day {
			result = append(result, subj)
		}
	}

	return result
}

// Функция для формирования текстового сообщения из списка предметов
func scheduleMessage(subjects []Subject) string {

	if len(subjects) == 0 {
		return "😔 В этот день пар не найдено."
	}

	var msg strings.Builder

	for _, subj := range subjects {
		msg.WriteString(fmt.Sprintf("📚 Предмет: %s\n⏰ Время: %s\n📍 Аудитория: %s, %s\n👥 Группы: %s\n\n",
			subj.SubjectName,
			subj.TotalTimeSchedule,
			subj.NumAuditoriumSchedule,
			subj.BuildingName,
			subj.GroupList))
	}

	return msg.String()
}

// Функция для генерации клавиатуры с кнопками дней недели и дополнительными кнопками
func getWeekButtons() tgbotapi.ReplyKeyboardMarkup {

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Понедельник"),
			tgbotapi.NewKeyboardButton("Вторник"),
			tgbotapi.NewKeyboardButton("Среда"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Четверг"),
			tgbotapi.NewKeyboardButton("Пятница"),
			tgbotapi.NewKeyboardButton("Суббота"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Воскресенье"),
			tgbotapi.NewKeyboardButton("Общее"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Перезапустить"),
		),
	)
	return keyboard

}

func main() {
	// Чтение токена Telegram из переменной окружения
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("Переменная окружения TELEGRAM_BOT_TOKEN не установлена")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Авторизовались как %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {

		if update.Message != nil {
			chatID := update.Message.Chat.ID

			// Обработка команды "перезапустить"
			if strings.ToLower(update.Message.Text) == "перезапустить" {
				delete(userData, chatID)
				msg := tgbotapi.NewMessage(chatID, "🔄 Перезапуск.\n📝 Введите ФИО для получения расписания:")
				bot.Send(msg)
				continue
			}

			// Если данные пользователя ещё не сохранены – считаем, что сообщение содержит ФИО
			if _, exists := userData[chatID]; !exists {

				fullName := update.Message.Text
				employee, err := getEmployee(fullName)

				if err != nil {
					msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при поиске преподавателя.\nПроверьте правильность введённых данных и попробуйте ещё раз.")
					bot.Send(msg)
					continue
				}

				// Сохраняем данные пользователя
				userData[chatID] = UserData{
					FullName:  fullName,
					TeacherID: employee.EmployeeID,
					Email:     employee.Email,
				}

				confirmation := fmt.Sprintf("✅ Пользователь найден!\n🆔 ID: %d\n✉️ Email: %s\n\nВыберите день недели для получения расписания:", employee.EmployeeID, employee.Email)
				msg := tgbotapi.NewMessage(chatID, confirmation)
				msg.ReplyMarkup = getWeekButtons()
				bot.Send(msg)

				continue
			}

			// Если данные уже сохранены – обрабатываем команды для дней недели или "Общее"
			text := strings.ToLower(update.Message.Text)
			user := userData[chatID]

			subjects, err := getSchedule(user.TeacherID)

			if err != nil {
				log.Printf("❌ Ошибка получения расписания для teacherID %d: %v", user.TeacherID, err)
				msg := tgbotapi.NewMessage(chatID, "❌ Ошибка получения расписания.\nПопробуйте позже или нажмите «Перезапустить».")
				bot.Send(msg)
				continue
			}

			var scheduleText string

			switch text {
			case "понедельник":
				scheduleText = scheduleMessage(filterScheduleByDay(subjects, 1))
			case "вторник":
				scheduleText = scheduleMessage(filterScheduleByDay(subjects, 2))
			case "среда":
				scheduleText = scheduleMessage(filterScheduleByDay(subjects, 3))
			case "четверг":
				scheduleText = scheduleMessage(filterScheduleByDay(subjects, 4))
			case "пятница":
				scheduleText = scheduleMessage(filterScheduleByDay(subjects, 5))
			case "суббота":
				scheduleText = scheduleMessage(filterScheduleByDay(subjects, 6))
			case "воскресенье":
				scheduleText = scheduleMessage(filterScheduleByDay(subjects, 7))
			case "общее":
				scheduleText = scheduleMessage(subjects)
			default:
				scheduleText = "🤔 Неизвестная команда.\nВыберите день недели из клавиатуры."
			}

			msg := tgbotapi.NewMessage(chatID, scheduleText)
			msg.ReplyMarkup = getWeekButtons()
			bot.Send(msg)
		}
	}
}
