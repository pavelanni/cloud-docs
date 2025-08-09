// Simple JavaScript for Cloud Docs demo
console.log('Cloud Docs static assets loaded successfully!');

document.addEventListener('DOMContentLoaded', function() {
    // Add a simple interaction
    const headers = document.querySelectorAll('h2');
    headers.forEach(header => {
        header.addEventListener('click', function() {
            this.style.color = this.style.color === 'rgb(16, 185, 129)' ? '#2563eb' : '#10b981';
        });
    });
    
    // Add timestamp to demonstrate dynamic content
    const timestamp = document.getElementById('timestamp');
    if (timestamp) {
        timestamp.textContent = new Date().toLocaleString();
    }
});