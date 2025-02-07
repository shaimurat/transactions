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
func generateTransactionID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

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
func handleTransaction(c *gin.Context) {
	var request TransactionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Calculate total price
	var totalPrice float64
	for _, item := range request.CartItems {
		totalPrice += item.Price
	}

	// Create transaction
	transaction := Transaction{
		ID:         generateTransactionID(),
		CartItems:  request.CartItems,
		Customer:   request.Customer,
		Status:     "Pending Payment",
		TotalPrice: totalPrice,
		CreatedAt:  time.Now(),
	}

	// Simulate payment (randomly succeed or fail)
	paymentSuccess := processPaymentMock()

	// Update transaction status
	if paymentSuccess {
		transaction.Status = "Paid"
		generateReceiptPDF(transaction)
		sendReceiptEmail(transaction)
	} else {
		transaction.Status = "Declined"
	}

	// Insert transaction into MongoDB
	_, err := transactionCollection.InsertOne(context.TODO(), transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Respond to the main server
	c.JSON(http.StatusOK, gin.H{"success": paymentSuccess, "transaction": transaction})
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
func main() {
	connectDB()

	r := gin.Default()

	// Enable CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Change this to your frontend domain in production
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
	}))

	r.POST("/api/transactions", handleTransaction)

	log.Println("Transaction service running on port 8081...")
	r.Run(":8081")
}
