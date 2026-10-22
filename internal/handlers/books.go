package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"online-lib/internal/models"
)

var isbnRe = regexp.MustCompile(`^\d{13}$`)

type BookHandler struct {
	DB *gorm.DB
}

type bookInput struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	Year        int    `json:"year"`
	ISBN        string `json:"isbn"`
	IsAvailable *bool  `json:"isAvailable"`
}

type bookPatchInput struct {
	Title       *string `json:"title"`
	Author      *string `json:"author"`
	Year        *int    `json:"year"`
	ISBN        *string `json:"isbn"`
	IsAvailable *bool   `json:"isAvailable"`
}

func (h *BookHandler) List(c *gin.Context) {
	query := h.DB.Model(&models.Book{})
	if author := c.Query("author"); author != "" {
		query = query.Where("author = ?", author)
	}
	var books []models.Book
	if err := query.Order("created_at ASC").Find(&books).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось получить список книг"})
		return
	}
	c.JSON(http.StatusOK, books)
}

func (h *BookHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var book models.Book
	if err := h.DB.First(&book, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "книга не найдена"})
		return
	}
	c.JSON(http.StatusOK, book)
}

func (h *BookHandler) Create(c *gin.Context) {
	var in bookInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидное тело запроса, ожидается JSON"})
		return
	}
	if errs := validateBookFields(strings.TrimSpace(in.Title), strings.TrimSpace(in.Author), &in.Year, strings.TrimSpace(in.ISBN)); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": strings.Join(errs, "; ")})
		return
	}
	var existing models.Book
	if err := h.DB.Where("isbn = ?", in.ISBN).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "книга с таким isbn уже существует"})
		return
	}
	book := models.Book{
		Title:       strings.TrimSpace(in.Title),
		Author:      strings.TrimSpace(in.Author),
		Year:        in.Year,
		ISBN:        in.ISBN,
		IsAvailable: in.IsAvailable == nil || *in.IsAvailable,
	}
	if err := h.DB.Create(&book).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "книга с таким isbn уже существует"})
		return
	}
	c.JSON(http.StatusCreated, book)
}

func (h *BookHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var in bookPatchInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидное тело запроса, ожидается JSON"})
		return
	}
	var errs []string
	if in.ISBN != nil {
		if isbn := strings.TrimSpace(*in.ISBN); !isbnRe.MatchString(isbn) {
			errs = append(errs, "isbn обязателен и должен состоять ровно из 13 цифр")
		}
	}
	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" || len([]rune(title)) > 200 {
			errs = append(errs, "title обязателен, длина от 1 до 200 символов")
		}
	}
	if in.Author != nil {
		author := strings.TrimSpace(*in.Author)
		if author == "" || len([]rune(author)) > 100 {
			errs = append(errs, "author обязателен, длина от 1 до 100 символов")
		}
	}
	if in.Year != nil {
		if y := *in.Year; y < 1450 || y > time.Now().Year() {
			errs = append(errs, fmt.Sprintf("year обязателен, от 1450 до %d", time.Now().Year()))
		}
	}
	if len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": strings.Join(errs, "; ")})
		return
	}
	var book models.Book
	if err := h.DB.First(&book, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "книга не найдена"})
		return
	}
	if in.ISBN != nil {
		isbn := strings.TrimSpace(*in.ISBN)
		var duplicate models.Book
		if err := h.DB.Where("isbn = ? AND id <> ?", isbn, book.ID).First(&duplicate).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "книга с таким isbn уже существует"})
			return
		}
	}
	updates := map[string]any{}
	if in.Title != nil {
		updates["title"] = strings.TrimSpace(*in.Title)
	}
	if in.Author != nil {
		updates["author"] = strings.TrimSpace(*in.Author)
	}
	if in.Year != nil {
		updates["year"] = *in.Year
	}
	if in.ISBN != nil {
		updates["isbn"] = strings.TrimSpace(*in.ISBN)
	}
	if in.IsAvailable != nil {
		updates["is_available"] = *in.IsAvailable
	}
	if len(updates) > 0 {
		if err := h.DB.Model(&book).Updates(updates).Error; err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "книга с таким isbn уже существует"})
			return
		}
	}
	c.JSON(http.StatusOK, book)
}

func (h *BookHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var book models.Book
	if err := h.DB.First(&book, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "книга не найдена"})
		return
	}
	if err := h.DB.Delete(&book, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось удалить книгу"})
		return
	}
	c.Status(http.StatusNoContent)
}

func parseID(c *gin.Context) (uint, bool) {
	raw := c.Param("id")
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || n == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный id, ожидается положительное число"})
		return 0, false
	}
	return uint(n), true
}

func validateBookFields(title, author string, year *int, isbn string) []string {
	var errs []string
	if title == "" || len([]rune(title)) > 200 {
		errs = append(errs, "title обязателен, длина от 1 до 200 символов")
	}
	if author == "" || len([]rune(author)) > 100 {
		errs = append(errs, "author обязателен, длина от 1 до 100 символов")
	}
	if year == nil || *year < 1450 || *year > time.Now().Year() {
		errs = append(errs, fmt.Sprintf("year обязателен, от 1450 до %d", time.Now().Year()))
	}
	if !isbnRe.MatchString(isbn) {
		errs = append(errs, "isbn обязателен и должен состоять ровно из 13 цифр")
	}
	return errs
}
