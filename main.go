package main

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// region Определение структур и методов для реализации учета (Группы, Студенты, Дисциплины, Занятия)

type Group struct {
	Name       string
	Students   []string
	StudentIDs map[string]uint32
}

func NewGroup(name string) *Group {
	return &Group{
		Name:       name,
		Students:   make([]string, 0),
		StudentIDs: make(map[string]uint32),
	}
}

func LoadGroup(name string) (*Group, error) {
	filename := name + ".txt"
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	group := NewGroup(name)
	if scanner.Scan() {
		group.Name = scanner.Text()
	}
	if scanner.Scan() {
		countStr := scanner.Text()
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return nil, err
		}
		for i := 0; i < count; i++ {
			if scanner.Scan() {
				studentName := scanner.Text()
				if scanner.Scan() {
					idStr := scanner.Text()
					id, err := strconv.ParseUint(idStr, 10, 32)
					if err != nil {
						return nil, err
					}
					group.Students = append(group.Students, studentName)
					group.StudentIDs[studentName] = uint32(id)
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return group, nil
}

func (g *Group) Save() error {
	filename := g.Name + ".txt"
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	fmt.Fprintln(writer, g.Name)
	fmt.Fprintln(writer, len(g.Students))

	for _, student := range g.Students {
		fmt.Fprintln(writer, student)
		fmt.Fprintln(writer, g.StudentIDs[student])
	}

	return writer.Flush()
}

func (g *Group) AddStudent(name string) {
	for _, student := range g.Students {
		if student == name {
			return
		}
	}
	h := fnv.New32a()
	h.Write([]byte(name))
	id := h.Sum32()
	g.Students = append(g.Students, name)
	g.StudentIDs[name] = id
}

func (g *Group) RemoveStudent(name string) bool {
	for i, student := range g.Students {
		if student == name {
			g.Students = append(g.Students[:i], g.Students[i+1:]...)
			delete(g.StudentIDs, name)
			return true
		}
	}
	return false
}

func (g *Group) GetStudentByID(id uint32) (string, bool) {
	for name, studentID := range g.StudentIDs {
		if studentID == id {
			return name, true
		}
	}
	return "", false
}

func (g *Group) Print() {
	fmt.Printf("Группа: %s\n", g.Name)
	fmt.Println("Студенты:")
	for _, student := range g.Students {
		fmt.Printf("  - %s\n", student)
	}
}

type PresenceRecord struct {
	StudentName string
	Present     bool
}

type Lesson struct {
	Date       string
	Attendance []PresenceRecord
}

func NewLesson(date string, group *Group) *Lesson {
	lesson := &Lesson{
		Date: date,
	}
	for _, student := range group.Students {
		lesson.Attendance = append(lesson.Attendance, PresenceRecord{
			StudentName: student,
			Present:     false,
		})
	}
	return lesson
}

func (l *Lesson) MarkPresent(studentID uint32) bool {
	for i, record := range l.Attendance {
		h := fnv.New32a()
		h.Write([]byte(record.StudentName))
		id := h.Sum32()
		if id == studentID && !record.Present {
			l.Attendance[i].Present = true
			return true
		}
	}
	return false
}

func (l *Lesson) GetStudentStatus(studentName string) bool {
	for _, record := range l.Attendance {
		if record.StudentName == studentName {
			return record.Present
		}
	}
	return false
}

func (l *Lesson) Print() {
	fmt.Printf("Занятие от %s:\n", l.Date)
	for _, record := range l.Attendance {
		status := "Отсутствует"
		if record.Present {
			status = "Присутствует"
		}
		fmt.Printf("  %s: %s\n", record.StudentName, status)
	}
}

type Subject struct {
	Name       string
	Group      *Group
	Lessons    []*Lesson
	LessonFile string
}

func NewSubject(name, groupName string) (*Subject, error) {
	subject := &Subject{
		Name: name,
	}
	group, err := LoadGroup(groupName)
	if err != nil {
		group = NewGroup(groupName)
		fmt.Println("Данные о группе не были найдены. Группа была создана.")
	}
	subject.Group = group
	subject.LessonFile = name + groupName + "L.txt"
	subject.loadLessons()
	return subject, nil
}

func (s *Subject) loadLessons() {
	file, err := os.Open(s.LessonFile)
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		countStr := scanner.Text()
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return
		}
		for i := 0; i < count; i++ {
			if scanner.Scan() {
				date := scanner.Text()
				if scanner.Scan() {
					studentCountStr := scanner.Text()
					studentCount, err := strconv.Atoi(studentCountStr)
					if err != nil {
						continue
					}
					lesson := &Lesson{Date: date}
					for j := 0; j < studentCount; j++ {
						if scanner.Scan() {
							studentName := scanner.Text()
							if scanner.Scan() {
								presentStr := scanner.Text()
								present := presentStr == "1"

								lesson.Attendance = append(lesson.Attendance, PresenceRecord{
									StudentName: studentName,
									Present:     present,
								})
							}
						}
					}
					s.Lessons = append(s.Lessons, lesson)
				}
			}
		}
	}
}

func (s *Subject) SaveLessons() error {
	file, err := os.Create(s.LessonFile)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	fmt.Fprintln(writer, len(s.Lessons))
	for _, lesson := range s.Lessons {
		fmt.Fprintln(writer, lesson.Date)
		fmt.Fprintln(writer, len(lesson.Attendance))
		for _, record := range lesson.Attendance {
			fmt.Fprintln(writer, record.StudentName)
			if record.Present {
				fmt.Fprintln(writer, "1")
			} else {
				fmt.Fprintln(writer, "0")
			}
		}
	}
	return writer.Flush()
}

func (s *Subject) AddLesson(date string) bool {
	for _, lesson := range s.Lessons {
		if lesson.Date == date {
			return false
		}
	}
	lesson := NewLesson(date, s.Group)
	s.Lessons = append(s.Lessons, lesson)
	s.SaveLessons()
	return true
}

func (s *Subject) DeleteLesson(date string) bool {
	for i, lesson := range s.Lessons {
		if lesson.Date == date {
			s.Lessons = append(s.Lessons[:i], s.Lessons[i+1:]...)
			s.SaveLessons()
			return true
		}
	}
	return false
}

func (s *Subject) ChangeLessonDate(oldDate, newDate string) bool {
	for _, lesson := range s.Lessons {
		if lesson.Date == newDate {
			return false
		}
	}
	for _, lesson := range s.Lessons {
		if lesson.Date == oldDate {
			lesson.Date = newDate
			s.SaveLessons()
			return true
		}
	}
	return false
}

func (s *Subject) FindLesson(date string) (*Lesson, int) {
	for i, lesson := range s.Lessons {
		if lesson.Date == date {
			return lesson, i
		}
	}
	return nil, -1
}

func (s *Subject) PrintLessons() {
	fmt.Printf("Даты занятий группы %s:\n", s.Group.Name)
	if len(s.Lessons) == 0 {
		fmt.Println("  Нет занятий")
	} else {
		for _, lesson := range s.Lessons {
			fmt.Printf("  - %s\n", lesson.Date)
		}
	}
}

func (s *Subject) PrintStudents() {
	fmt.Printf("Студенты группы %s:\n", s.Group.Name)
	for _, student := range s.Group.Students {
		fmt.Printf("  - %s\n", student)
	}
}

func (s *Subject) GetStudentAttendanceStatus(studentName string) (int, int) {
	presentCount := 0
	totalLessons := len(s.Lessons)
	for _, lesson := range s.Lessons {
		if lesson.GetStudentStatus(studentName) {
			presentCount++
		}
	}
	return presentCount, totalLessons
}

func (s *Subject) PrintGroupAttendance() {
	fmt.Printf("Посещаемость группы %s:\n", s.Group.Name)
	fmt.Println("===================================")
	for _, student := range s.Group.Students {
		present, total := s.GetStudentAttendanceStatus(student)
		percentage := 0.0
		if total > 0 {
			percentage = float64(present) / float64(total) * 100
		}
		fmt.Printf("%-30s: %d/%d (%.1f%%)\n", student, present, total, percentage)
	}
}

// endregion

// region Реализация основной функциональности (сканирование)

func (s *Subject) BeginScanCamera(date string) bool {
	lesson, _ := s.FindLesson(date)
	if lesson == nil {
		fmt.Println("Ошибка: занятие не найдено")
		return false
	}
	fmt.Printf("\n=== УЧЕТ ПОСЕЩАЕМОСТИ (КАМЕРА) ===\n")
	fmt.Printf("Занятие: %s\n", date)
	fmt.Printf("Группа: %s\n\n", s.Group.Name)
	lesson.Print()
	fmt.Println("\nЗапуск сканирования с камеры...")
	fmt.Println(" ESC - завершить сканирование")
	fmt.Println(" SPACE - ручной ввод ID")
	fmt.Println()
	cmd := exec.Command("python", "camera.py")
	// Каналы для общения с Python процессом
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Ошибка создания pipe: %v\n", err)
		return false
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("Ошибка создания stderr pipe: %v\n", err)
		return false
	}
	if err := cmd.Start(); err != nil {
		fmt.Printf("Ошибка запуска Python: %v\n", err)
		return false
	}
	done := make(chan bool)
	scannedCount := 0
	totalStudents := len(s.Group.Students)
	// Чтения вывода Python
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "SCAN:") {
				// Получили отсканированный ID
				idStr := strings.TrimPrefix(line, "SCAN:")
				studentID, err := strconv.ParseUint(idStr, 10, 32)
				if err != nil {
					fmt.Printf("Ошибка парсинга ID: %v\n", err)
					continue
				}
				// Ищем студента
				studentName, found := s.Group.GetStudentByID(uint32(studentID))
				if !found {
					fmt.Printf("Студент с ID %d не найден в группе!\n", studentID)
					continue
				}
				// Отмечаем присутствие
				if lesson.MarkPresent(uint32(studentID)) {
					scannedCount++
					fmt.Printf("%s отмечен как присутствующий (%d/%d)\n", studentName, scannedCount, totalStudents)
					if scannedCount%3 == 0 || scannedCount == totalStudents {
						fmt.Println("\nТекущий статус:")
						lesson.Print()
					}
				} else {
					fmt.Printf("✗ %s уже отмечен как присутствующий\n", studentName)
				}
			} else if line == "MANUAL_INPUT" {
				fmt.Print("\nВведите ID студента вручную: ")
				manualScanner := bufio.NewScanner(os.Stdin)
				if manualScanner.Scan() {
					idStr := manualScanner.Text()
					studentID, err := strconv.ParseUint(idStr, 10, 32)
					if err != nil {
						fmt.Println("Ошибка: введите число")
						continue
					}
					studentName, found := s.Group.GetStudentByID(uint32(studentID))
					if !found {
						fmt.Printf("Студент с ID %d не найден в группе!\n", studentID)
						continue
					}
					if lesson.MarkPresent(uint32(studentID)) {
						scannedCount++
						fmt.Printf("%s отмечен как присутствующий (%d/%d)\n",
							studentName, scannedCount, totalStudents)
						if scannedCount%3 == 0 || scannedCount == totalStudents {
							fmt.Println("\nТекущий статус:")
							lesson.Print()
						}
					} else {
						fmt.Printf("%s уже отмечен как присутствующий\n", studentName)
					}
				}
			} else if strings.HasPrefix(line, "INFO:") || strings.HasPrefix(line, "ERROR:") {
				fmt.Println(line)
			} else if line == "SCAN_COMPLETE" {
				fmt.Println("Сканирование завершено")
				done <- true
				return
			}
		}
	}()
	// Чтения ошибок
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Printf("Python: %s\n", line)
		}
	}()
	fmt.Println("Сканирование запущено. Нажмите ESC в окне камеры для завершения...")
	saveTicker := time.NewTicker(10 * time.Second)
	defer saveTicker.Stop()
	select {
	case <-done:
	case <-saveTicker.C:
		s.SaveLessons()
		fmt.Println("Прогресс автоматически сохранен")
		select {
		case <-done:
		case <-time.After(30 * time.Minute):
			fmt.Println("Таймаут сканирования")
		}
	}
	time.Sleep(500 * time.Millisecond)
	if cmd.Process != nil {
		cmd.Process.Signal(os.Interrupt)
		cmd.Wait()
	}
	s.SaveLessons()
	finalCount := 0
	for _, record := range lesson.Attendance {
		if record.Present {
			finalCount++
		}
	}
	fmt.Printf("\nУчет завершен. Отмечено %d из %d студентов.\n", finalCount, totalStudents)
	return true
}

func (s *Subject) BeginScanManual(date string) bool {
	lesson, _ := s.FindLesson(date)
	if lesson == nil {
		return false
	}
	fmt.Printf("\n=== УЧЕТ ПОСЕЩАЕМОСТИ (РУЧНОЙ ВВОД) ===\n")
	fmt.Printf("Занятие: %s\n", date)
	fmt.Printf("Группа: %s\n\n", s.Group.Name)
	lesson.Print()
	fmt.Println("\nВведите ID студентов (по одному, 'q' для завершения):")
	scanner := bufio.NewScanner(os.Stdin)
	scannedCount := 0
	totalStudents := len(s.Group.Students)
	for scannedCount < totalStudents {
		fmt.Printf("\nОжидание ввода ID... (отмечено: %d/%d)\n", scannedCount, totalStudents)
		fmt.Print("ID студента: ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "q" || input == "quit" {
			break
		}
		id, err := strconv.ParseUint(input, 10, 32)
		if err != nil {
			fmt.Println("Ошибка: введите число или 'q' для выхода")
			continue
		}
		studentName, found := s.Group.GetStudentByID(uint32(id))
		if !found {
			fmt.Printf("Студент с ID %d не найден в группе!\n", id)
			continue
		}
		if lesson.MarkPresent(uint32(id)) {
			fmt.Printf("✓ %s отмечен как присутствующий\n", studentName)
			scannedCount++
			fmt.Println("\nТекущий статус:")
			lesson.Print()
		} else {
			fmt.Printf("✗ %s уже отмечен как присутствующий\n", studentName)
		}
	}
	s.SaveLessons()
	fmt.Printf("\nУчет завершен. Отмечено %d из %d студентов.\n", scannedCount, totalStudents)
	return true
}

// endregion

// region UI

func (s *Subject) ChangeGroupMenu() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("\n=== ИЗМЕНЕНИЕ ГРУППЫ ===")
		s.PrintStudents()
		fmt.Println("\nВыберите опцию:")
		fmt.Println("1. Добавить студента")
		fmt.Println("2. Удалить студента")
		fmt.Println("3. Изменить имя группы")
		fmt.Println("4. Назад")
		fmt.Print("> ")
		var choice int
		fmt.Scan(&choice)
		scanner.Scan()
		switch choice {
		case 1:
			fmt.Print("Введите имя студента: ")
			scanner.Scan()
			name := scanner.Text()
			s.Group.AddStudent(name)
			s.Group.Save()
			fmt.Println("Студент добавлен")
		case 2:
			fmt.Print("Введите имя студента: ")
			scanner.Scan()
			name := scanner.Text()
			if s.Group.RemoveStudent(name) {
				s.Group.Save()
				fmt.Println("Студент удален")
			} else {
				fmt.Println("Студент не найден")
			}
		case 3:
			fmt.Print("Новое имя группы: ")
			scanner.Scan()
			newName := scanner.Text()
			oldFile := s.Group.Name + ".txt"
			newFile := newName + ".txt"
			if err := os.Rename(oldFile, newFile); err != nil {
				fmt.Printf("Ошибка переименования файла: %v\n", err)
			} else {
				s.Group.Name = newName
				s.LessonFile = s.Name + newName + "L.txt"
				s.Group.Save()
				s.SaveLessons()
				fmt.Println("Имя группы изменено")
			}
		case 4:
			s.Group.Save()
			s.SaveLessons()
			return
		default:
			fmt.Println("Неверный выбор")
		}
	}
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func Menu() {
	scanner := bufio.NewScanner(os.Stdin)
	var currentSubject *Subject
	for {
		clearScreen()
		fmt.Println("\n=== СИСТЕМА УЧЕТА ПОСЕЩАЕМОСТИ ===")
		fmt.Println("1. Выбрать дисциплину и группу")
		if currentSubject != nil {
			fmt.Printf("   Текущая: %s, Группа: %s\n", currentSubject.Name, currentSubject.Group.Name)
			fmt.Println("2. Сменить группу")
			fmt.Println("3. Управление занятиями")
			fmt.Println("4. Провести учет посещаемости")
			fmt.Println("5. Просмотр статистики")
			fmt.Println("6. Управление группой")
		} else {
			fmt.Println("2. (Сначала выберите дисциплину)")
		}
		fmt.Println("0. Выйти")
		fmt.Print("\nВыберите опцию: ")
		var choice string
		fmt.Scan(&choice)
		scanner.Scan()
		switch choice {
		case "1":
			clearScreen()
			fmt.Print("Введите название дисциплины: ")
			scanner.Scan()
			subjectName := scanner.Text()
			fmt.Print("Введите номер группы: ")
			scanner.Scan()
			groupName := scanner.Text()
			var err error
			currentSubject, err = NewSubject(subjectName, groupName)
			if err != nil {
				fmt.Printf("Ошибка: %v\n", err)
				fmt.Print("Нажмите Enter для продолжения...")
				scanner.Scan()
			} else {
				fmt.Printf("Дисциплина '%s' для группы '%s' загружена\n",
					subjectName, currentSubject.Group.Name)
				fmt.Print("Нажмите Enter для продолжения...")
				scanner.Scan()
			}
		case "2":
			if currentSubject == nil {
				fmt.Println("Сначала выберите дисциплину")
				fmt.Print("Нажмите Enter для продолжения...")
				scanner.Scan()
				break
			}
			clearScreen()
			fmt.Print("Введите номер группы: ")
			scanner.Scan()
			groupName := scanner.Text()
			group, err := LoadGroup(groupName)
			if err != nil {
				fmt.Printf("Ошибка загрузки группы: %v\n", err)
				fmt.Println("Создать новую группу? (y/n)")
				scanner.Scan()
				if strings.ToLower(scanner.Text()) == "y" {
					group = NewGroup(groupName)
					currentSubject.Group = group
					currentSubject.LessonFile = currentSubject.Name + groupName + "L.txt"
					fmt.Println("Новая группа создана")
				}
			} else {
				currentSubject.Group = group
				currentSubject.LessonFile = currentSubject.Name + groupName + "L.txt"
				currentSubject.loadLessons()
				fmt.Println("Группа загружена")
			}
			fmt.Print("Нажмите Enter для продолжения...")
			scanner.Scan()
		case "3":
			if currentSubject == nil {
				fmt.Println("Сначала выберите дисциплину")
				fmt.Print("Нажмите Enter для продолжения...")
				scanner.Scan()
				break
			}
			lessonManagementMenu(currentSubject, scanner)
		case "4":
			if currentSubject == nil {
				fmt.Println("Сначала выберите дисциплину")
				fmt.Print("Нажмите Enter для продолжения...")
				scanner.Scan()
				break
			}
			attendanceMenu(currentSubject, scanner)
		case "5":
			if currentSubject == nil {
				fmt.Println("Сначала выберите дисциплину")
				fmt.Print("Нажмите Enter для продолжения...")
				scanner.Scan()
				break
			}
			clearScreen()
			statisticsMenu(currentSubject, scanner)
		case "6":
			if currentSubject == nil {
				fmt.Println("Сначала выберите дисциплину")
				fmt.Print("Нажмите Enter для продолжения...")
				scanner.Scan()
				break
			}
			currentSubject.ChangeGroupMenu()
		case "0":
			if currentSubject != nil {
				currentSubject.Group.Save()
				currentSubject.SaveLessons()
			}
			fmt.Println("Выход из программы...")
			return
		default:
			fmt.Println("Неверный выбор")
			fmt.Print("Нажмите Enter для продолжения...")
			scanner.Scan()
		}
	}
}

func attendanceMenu(subject *Subject, scanner *bufio.Scanner) {
	clearScreen()
	subject.PrintLessons()

	fmt.Print("\nВведите дату занятия (например: 15 01 2024): ")
	scanner.Scan()
	date := scanner.Text()

	lesson, _ := subject.FindLesson(date)
	if lesson == nil {
		fmt.Println("Занятие не найдено. Создать новое? (y/n)")
		scanner.Scan()
		if strings.ToLower(scanner.Text()) == "y" {
			subject.AddLesson(date)
			fmt.Println("Занятие создано")
		} else {
			return
		}
	}

	fmt.Println("\nВыберите метод учета посещаемости:")
	fmt.Println("1. Сканирование с веб-камеры")
	fmt.Println("2. Ручной ввод ID")
	fmt.Print("> ")

	var method int
	fmt.Scan(&method)
	scanner.Scan()

	switch method {
	case 1:
		subject.BeginScanCamera(date)
	case 2:
		subject.BeginScanManual(date)
	default:
		fmt.Println("Неверный выбор, используется ручной ввод")
		subject.BeginScanManual(date)
	}

	fmt.Print("Нажмите Enter для продолжения...")
	scanner.Scan()
}

func lessonManagementMenu(subject *Subject, scanner *bufio.Scanner) {
	for {
		fmt.Println("\n=== УПРАВЛЕНИЕ ЗАНЯТИЯМИ ===")
		subject.PrintLessons()
		fmt.Println("\nВыберите опцию:")
		fmt.Println("1. Добавить занятие")
		fmt.Println("2. Удалить занятие")
		fmt.Println("3. Изменить дату занятия")
		fmt.Println("4. Просмотреть посещаемость занятия")
		fmt.Println("5. Назад")
		fmt.Print("> ")
		var choice int
		fmt.Scan(&choice)
		scanner.Scan()
		switch choice {
		case 1:
			fmt.Print("Введите дату занятия (Формат: 30 12 2000): ")
			scanner.Scan()
			date := scanner.Text()
			if subject.AddLesson(date) {
				fmt.Println("Занятие добавлено")
			} else {
				fmt.Println("Занятие с такой датой уже существует")
			}
		case 2:
			fmt.Print("Введите дату занятия для удаления: ")
			scanner.Scan()
			date := scanner.Text()
			if subject.DeleteLesson(date) {
				fmt.Println("Занятие удалено")
			} else {
				fmt.Println("Занятие не найдено")
			}
		case 3:
			fmt.Print("Введите текущую дату занятия: ")
			scanner.Scan()
			oldDate := scanner.Text()
			fmt.Print("Введите новую дату: ")
			scanner.Scan()
			newDate := scanner.Text()
			if subject.ChangeLessonDate(oldDate, newDate) {
				fmt.Println("Дата занятия изменена")
			} else {
				fmt.Println("Ошибка: занятие не найдено или новая дата уже существует")
			}
		case 4:
			fmt.Print("Введите дату занятия: ")
			scanner.Scan()
			date := scanner.Text()
			lesson, _ := subject.FindLesson(date)
			if lesson != nil {
				lesson.Print()
			} else {
				fmt.Println("Занятие не найдено")
			}
		case 5:
			return
		default:
			fmt.Println("Неверный выбор")
		}
		fmt.Print("Нажмите Enter для продолжения...")
		scanner.Scan()
	}
}

func statisticsMenu(subject *Subject, scanner *bufio.Scanner) {
	for {
		fmt.Println("\n=== СТАТИСТИКА ПОСЕЩАЕМОСТИ ===")
		fmt.Println("1. Общая статистика группы")
		fmt.Println("2. Статистика по конкретному студенту")
		fmt.Println("3. Просмотр занятия")
		fmt.Println("4. Назад")
		fmt.Print("> ")

		var choice int
		fmt.Scan(&choice)
		scanner.Scan()

		switch choice {
		case 1:
			subject.PrintGroupAttendance()

		case 2:
			subject.PrintStudents()
			fmt.Print("\nВведите имя студента: ")
			scanner.Scan()
			studentName := scanner.Text()

			present, total := subject.GetStudentAttendanceStatus(studentName)
			if total > 0 {
				percentage := float64(present) / float64(total) * 100
				fmt.Printf("\nСтатистика для %s:\n", studentName)
				fmt.Printf("Посещено: %d из %d занятий\n", present, total)
				fmt.Printf("Процент посещаемости: %.1f%%\n", percentage)
			} else {
				fmt.Println("Нет данных о занятиях или студент не найден")
			}

		case 3:
			fmt.Print("Введите дату занятия: ")
			scanner.Scan()
			date := scanner.Text()

			lesson, _ := subject.FindLesson(date)
			if lesson != nil {
				lesson.Print()
			} else {
				fmt.Println("Занятие не найдено")
			}

		case 4:
			return

		default:
			fmt.Println("Неверный выбор")
		}

		fmt.Print("\nНажмите Enter для продолжения...")
		scanner.Scan()
	}
}

// endregion

func main() {
	fmt.Println("\nНажмите Enter для запуска системы")
	bufio.NewScanner(os.Stdin).Scan()
	Menu()
}
