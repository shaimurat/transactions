package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/gomail.v2"
	"log"
	"net/http"
	"os"
	"time"
)

// MongoDB connection
var client *mongo.Client
var transactionCollection *mongo.Collection

// Transaction model
type Transaction struct {
	ID         string     `bson:"_id,omitempty" json:"id"`
	CartItems  []CartItem `bson:"cartItems" json:"cartItems"`
	Customer   Customer   `bson:"customer" json:"customer"`
	Status     string     `bson:"status" json:"status"` // Pending, Paid, Declined
	TotalPrice float64    `bson:"totalPrice" json:"totalPrice"`
	CreatedAt  time.Time  `bson:"createdAt" json:"createdAt"`
}

// CartItem model
type CartItem struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// Customer model
type Customer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Request body
type TransactionRequest struct {
	CartItems []CartItem `json:"cartItems"`
	Customer  Customer   `json:"customer"`
}

// Generate unique transaction ID

// Connect to MongoDB
func connectDB() {
	var err error
	clientOptions := options.Client().ApplyURI("mongodb+srv://danial:Danial_2005@pokegame.fxobs.mongodb.net/?retryWrites=true&w=majority&appName=PokeGame\"")
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	transactionCollection = client.Database("PokeGame").Collection("transactions")
	fmt.Println("Connected to MongoDB")
}

// Handle transaction request
// Генерация уникального ID для транзакции
func generateTransactionID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes) // Уникальный 16-символьный ID
}

// Обработка создания транзакции (первый шаг)
func handleTransaction(c *gin.Context) {
	var request TransactionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Рассчитываем общую сумму
	var totalPrice float64
	for _, item := range request.CartItems {
		totalPrice += item.Price
	}

	// Создаем транзакцию со статусом "in process"
	transaction := Transaction{
		ID:         generateTransactionID(), // Сюда передается уникальный ID
		CartItems:  request.CartItems,
		Customer:   request.Customer,
		Status:     "in process",
		TotalPrice: totalPrice,
		CreatedAt:  time.Now(),
	}

	// Записываем в MongoDB
	_, err := transactionCollection.InsertOne(context.TODO(), transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Отправляем обратно ID, чтобы фронтенд знал, куда перенаправлять пользователя
	c.JSON(http.StatusOK, gin.H{"success": true, "transactionId": transaction.ID})
}

// Simulated payment function
// Simulated payment function (always successful)
func processPaymentMock() bool {
	return true // Always return true to simulate successful payment
}

// Generate PDF receipt
func generateReceiptPDF(transaction Transaction) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Pokemon Store Receipt")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Transaction ID: %s", transaction.ID))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Customer: %s", transaction.Customer.Name))
	pdf.Ln(10)
	pdf.Cell(40, 10, "Items:")
	pdf.Ln(10)

	for _, item := range transaction.CartItems {
		pdf.Cell(40, 10, fmt.Sprintf("- %s: $%.2f", item.Name, item.Price))
		pdf.Ln(5)
	}

	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Total: $%.2f", transaction.TotalPrice))
	pdf.Ln(10)
	pdf.Cell(40, 10, "Thank you for your purchase!")

	receiptPath := fmt.Sprintf("receipts/%s.pdf", transaction.ID)
	_ = os.Mkdir("receipts", 0755) // Ensure the receipts directory exists
	err := pdf.OutputFileAndClose(receiptPath)
	if err != nil {
		log.Println("Failed to generate PDF:", err)
	}
}
func sendReceiptEmail(transaction Transaction) {
	receiptPath := fmt.Sprintf("receipts/%s.pdf", transaction.ID)

	m := gomail.NewMessage()
	m.SetHeader("From", "pokeGame@gmail.com")
	m.SetHeader("To", transaction.Customer.Email)
	m.SetHeader("Subject", "Your Purchase Receipt")
	m.SetBody("text/plain", "Thank you for your purchase. Attached is your receipt.")
	m.Attach(receiptPath)

	d := gomail.NewDialer("smtp.gmail.com", 587, "isiki.edenovy@gmail.com", "lswy dyxe pnjd sjkk")

	if err := d.DialAndSend(m); err != nil {
		log.Println("Failed to send email:", err)
	} else {
		log.Println("Receipt emailed successfully")
	}
}

// Получение информации о транзакции
func getTransaction(c *gin.Context) {
	transactionID := c.Param("id")

	var transaction Transaction
	err := transactionCollection.FindOne(context.TODO(), map[string]interface{}{"_id": transactionID}).Decode(&transaction)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Transaction not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "transaction": transaction})
}

// Подтверждение платежа (обновление статуса)
func confirmPayment(c *gin.Context) {
	var request struct {
		TransactionID  string `json:"transactionId"`
		CardNumber     string `json:"cardNumber"`
		ExpirationDate string `json:"expirationDate"`
		CVV            string `json:"cvv"`
		Name           string `json:"name"`
		Address        string `json:"address"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Проверяем, есть ли такая транзакция
	var transaction Transaction
	err := transactionCollection.FindOne(context.TODO(), map[string]interface{}{"_id": request.TransactionID}).Decode(&transaction)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Transaction not found"})
		return
	}

	// Эмуляция платежа
	paymentSuccess := processPaymentMock()

	// Обновление статуса
	newStatus := "ended"
	if !paymentSuccess {
		newStatus = "failed"
	}

	_, err = transactionCollection.UpdateOne(
		context.TODO(),
		map[string]interface{}{"_id": request.TransactionID},
		map[string]interface{}{"$set": map[string]interface{}{"status": newStatus}},
	)

	if err == nil && paymentSuccess {
		generateReceiptPDF(transaction)
		sendReceiptEmail(transaction)
	}

	c.JSON(http.StatusOK, gin.H{"success": paymentSuccess})
}

func main() {
	connectDB()

	r := gin.Default()

	// Enable CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Разрешает запросы со всех источников
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "PUT", "DELETE"},
		AllowHeaders:     []string{"*"}, // Разрешает все заголовки
		ExposeHeaders:    []string{"*"}, // Позволяет клиенту видеть все заголовки ответа
		AllowCredentials: true,          // Позволяет передавать куки и заголовки авторизации
		MaxAge:           12 * time.Hour,
	}))

	r.POST("/api/transactions", handleTransaction)

	// Получение информации о транзакции по ID
	r.GET("/api/transaction/:id", getTransaction)
	r.StaticFS("/static", http.Dir("./static"))

	// Serve the transaction page correctly
	r.GET("/transaction", func(c *gin.Context) {
		c.File("./static/transaction.html")
	})

	// Подтверждение оплаты (изменяет статус на "ended" или "failed")
	r.POST("/api/confirm-payment", confirmPayment)
	log.Println("Transaction service running on port 8081...")
	r.Run(":8081")
}
