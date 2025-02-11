document.addEventListener("DOMContentLoaded", async () => {
    const urlParams = new URLSearchParams(window.location.search);
    const transactionId = urlParams.get("id");

    if (!transactionId) {
        document.getElementById("paymentStatus").innerText = "Transaction ID is missing.";
        return;
    }

    document.getElementById("transactionId").innerText = transactionId;

    try {
        const response = await fetch(`https://transactions-production-e9c4.up.railway.app/api/transaction-details/${transactionId}`);
        const transaction = await response.json();

        if (!transaction.success) {
            document.getElementById("paymentStatus").innerText = "Transaction not found.";
            return;
        }

        document.getElementById("totalAmount").innerText = transaction.transaction.totalPrice.toFixed(2);
    } catch (error) {
        document.getElementById("paymentStatus").innerText = "Error loading transaction.";
        console.error(error);
    }
});

// Функция отправки данных карты на сервер
async function submitPayment() {
    const transactionId = document.getElementById("transactionId").innerText;

    const paymentData = {
        transactionId: transactionId,
        cardNumber: document.getElementById("cardNumber").value.replace(/\s/g, ""),
        expirationDate: document.getElementById("expirationDate").value,
        cvv: document.getElementById("cvv").value,
        name: document.getElementById("cardName").value,
        address: document.getElementById("billingAddress").value
    };

    try {
        const response = await fetch("https://transactions-production-e9c4.up.railway.app/api/confirm-payment", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(paymentData)
        });

        const result = await response.json();

        if (result.success) {
            document.getElementById("paymentStatus").innerText = "Payment successful!";
            // Перенаправление на страницу с подтверждением
            setTimeout(() => {
                window.location.href = `../confirmation.html`;
            }, 2000);
        } else {
            document.getElementById("paymentStatus").innerText = "Payment failed. Please check your details.";
        }
    } catch (error) {
        document.getElementById("paymentStatus").innerText = "Error processing payment.";
        console.error(error);
    }
}
