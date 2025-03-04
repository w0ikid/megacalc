document.addEventListener("DOMContentLoaded", () => {
    const expressionInput = document.getElementById("expression");
    const calculateButton = document.getElementById("calculate");
    const expressionsList = document.getElementById("expressions-list");
    const loading = document.getElementById("loading");

    async function calculateExpression(expression) {
        try {
            const response = await fetch("/api/v1/calculate", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ expression }),
            });
            
            if (!response.ok) {
                throw new Error("Failed to submit expression");
            }

            const data = await response.json();
            return data.id;
        } catch (error) {
            console.error("Error submitting expression:", error);
            return null;
        }
    }

    async function fetchExpressions() {
        try {
            const response = await fetch("/api/v1/expressions");
            if (!response.ok) {
                throw new Error("Failed to fetch expressions");
            }

            const data = await response.json();
            return data.expressions;
        } catch (error) {
            console.error("Error fetching expressions:", error);
            return [];
        }
    }

    async function updateExpressionsList() {
        loading.style.display = "block";
        expressionsList.innerHTML = "";

        const expressions = await fetchExpressions();
        loading.style.display = "none";

        if (expressions.length === 0) {
            expressionsList.innerHTML = "<p>No expressions available.</p>";
            return;
        }

        expressions.forEach(exp => {
            const item = document.createElement("div");
            item.classList.add("expression-item");
            item.innerHTML = `<p><strong>ID:</strong> ${exp.id} <br> <strong>Status:</strong> ${exp.status} <br> <strong>Result:</strong> ${exp.result !== null ? exp.result : "Pending"}</p>`;
            expressionsList.appendChild(item);
        });
    }

    calculateButton.addEventListener("click", async () => {
        const expression = expressionInput.value.trim();
        if (!expression) return;

        const id = await calculateExpression(expression);
        if (id) {
            expressionInput.value = "";
            setTimeout(updateExpressionsList, 2000); // Delay to allow processing
        }
    });

    updateExpressionsList();
    setInterval(updateExpressionsList, 5000); // Refresh expressions list every 5 sec
});
