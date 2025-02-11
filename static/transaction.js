document.addEventListener("DOMContentLoaded", async () => {
    const urlParams = new URLSearchParams(window.location.search);
    const transactionId = urlParams.get("id");

    if (!transactionId) {
        document.getElementById("paymentStatus").innerText = "Transaction ID is missing.";
        return;
    }

    document.getElementById("transactionId").innerText = transactionId;

    // Fetch transaction details
    try {
        const response = await fetch(`https://transactions-production-e9c4.up.railway.app/api/transaction/${transactionId}`);
        const transaction = await response.json();

        if (!transaction || !transaction.success) {
            document.getElementById("paymentStatus").innerText = "Transaction not found.";
            return;
        }

        document.getElementById("totalAmount").innerText = transaction.transaction.totalPrice.toFixed(2);
    } catch (error) {
        document.getElementById("paymentStatus").innerText = "Error loading transaction.";
        console.error(error);
    }
});

// Apply input formatting
document.getElementById("cardNumber").addEventListener("input", function (e) {
    let value = e.target.value.replace(/\D/g, ""); // Remove non-digits
    value = value.replace(/(\d{4})/g, "$1 ").trim(); // Add space every 4 digits
    e.target.value = value;
});

document.getElementById("expirationDate").addEventListener("input", function (e) {
    let value = e.target.value.replace(/\D/g, ""); // Remove non-digits
    if (value.length >= 2) {
        value = value.substring(0, 2) + "/" + value.substring(2, 4);
    }
    e.target.value = value.substring(0, 5); // Limit to MM/YY
});

document.getElementById("cvv").addEventListener("input", function (e) {
    e.target.value = e.target.value.replace(/\D/g, "").substring(0, 4); // Only digits, max 4
});

// Function to submit payment
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
            localStorage.setItem("paymentSuccess", "Your payment was successful!");

            console.log(("Stored in localStorage:", localStorage.getItem("paymentSuccess")))
            window.location.href = "https://awesomeproject1-production.up.railway.app/pokemonsPage";
        } else {
            document.getElementById("paymentStatus").innerText = result.error;
        }
    } catch (error) {
        document.getElementById("paymentStatus").innerText = "Error processing payment.";
        console.error(error);
    }
}
