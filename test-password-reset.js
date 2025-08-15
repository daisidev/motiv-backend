// Simple test script to check password reset functionality
const fetch = require('node-fetch');

async function testPasswordReset() {
    try {
        const response = await fetch('http://localhost:8080/api/v1/auth/forgot-password', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                email: 'test@example.com'
            })
        });

        const result = await response.json();
        console.log('Status:', response.status);
        console.log('Response:', result);
    } catch (error) {
        console.error('Error testing password reset:', error);
    }
}

testPasswordReset();
