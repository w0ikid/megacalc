document.addEventListener('DOMContentLoaded', function() {
    const calculateButton = document.getElementById('calculate');
    const expressionInput = document.getElementById('expression');
    const expressionsList = document.getElementById('expressions-list');
    const loadingElement = document.getElementById('loading');
    
    // Hide loading spinner initially
    loadingElement.style.display = 'none';
    
    // Load existing expressions when page loads
    loadExpressions();
    
    // Set up expression polling
    setInterval(loadExpressions, 2000);
    
    // Add event listener for calculate button
    calculateButton.addEventListener('click', submitExpression);
    
    // Add event listener for enter key
    expressionInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            submitExpression();
        }
    });
    
    // Function to submit expression
    function submitExpression() {
        const expression = expressionInput.value.trim();
        if (!expression) {
            alert('Please enter an expression');
            return;
        }
        
        // Show loading
        loadingElement.style.display = 'flex';
        
        fetch('/api/v1/calculate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ expression })
        })
        .then(response => {
            if (!response.ok) {
                return response.json().then(data => {
                    throw new Error(data.error || 'Failed to submit expression');
                });
            }
            return response.json();
        })
        .then(data => {
            expressionInput.value = '';
            loadExpressions();
        })
        .catch(error => {
            alert(error.message);
        })
        .finally(() => {
            // Hide loading
            loadingElement.style.display = 'none';
        });
    }
    
    // Function to load expressions
    function loadExpressions() {
        fetch('/api/v1/expressions')
        .then(response => response.json())
        .then(data => {
            renderExpressions(data.expressions);
        })
        .catch(error => {
            console.error('Error loading expressions:', error);
        });
    }
    
    // Function to render expressions
    function renderExpressions(expressions) {
        if (!expressions || expressions.length === 0) {
            expressionsList.innerHTML = '<p class="no-expressions">No expressions calculated yet</p>';
            return;
        }
        
        // Sort expressions by status (pending first)
        expressions.sort((a, b) => {
            // Sort pending first, then processing, then completed
            const statusOrder = {
                'pending': 0,
                'processing': 1,
                'completed': 2,
                'failed': 3
            };
            return statusOrder[a.status] - statusOrder[b.status];
        });
        
        expressionsList.innerHTML = '';
        
        expressions.forEach(expr => {
            const expressionElement = document.createElement('div');
            expressionElement.className = `expression-item status-${expr.status}`;
            
            // Add ID as data attribute for reference
            expressionElement.dataset.id = expr.id;
            
            // Create status indicator
            const statusIndicator = document.createElement('div');
            statusIndicator.className = 'status-indicator';
            
            // Create expression details
            const detailsElement = document.createElement('div');
            detailsElement.className = 'expression-details';
            
            // Get expression from server by ID
            fetch(`/api/v1/expressions/${expr.id}`)
                .then(response => response.json())
                .then(data => {
                    const expression = data.expression;
                    
                    let statusText = '';
                    switch (expression.status) {
                        case 'pending':
                            statusText = 'Waiting...';
                            break;
                        case 'processing':
                            statusText = 'Processing...';
                            break;
                        case 'completed':
                            statusText = 'Result: ' + expression.result;
                            break;
                        case 'failed':
                            statusText = 'Failed';
                            break;
                    }
                    
                    detailsElement.innerHTML = `
                        <div class="expression-id">ID: ${expression.id}</div>
                        <div class="expression-status">${statusText}</div>
                    `;
                })
                .catch(error => {
                    detailsElement.innerHTML = `
                        <div class="expression-id">ID: ${expr.id}</div>
                        <div class="expression-status">Error loading details</div>
                    `;
                });
            
            expressionElement.appendChild(statusIndicator);
            expressionElement.appendChild(detailsElement);
            expressionsList.appendChild(expressionElement);
        });
    }
});