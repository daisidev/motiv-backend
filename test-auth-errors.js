// Test script to verify improved authentication error handling
const fetch = require('node-fetch');

const API_BASE = 'http://localhost:8080/api/v1';

async function testAuthErrors() {
    console.log('🧪 Testing Authentication Error Handling\n');
    console.log('=' .repeat(60));

    const tests = [
        {
            name: 'Login with non-existent email',
            endpoint: '/auth/login',
            method: 'POST',
            data: {
                email: 'nonexistent@example.com',
                password: 'password123'
            },
            expectedError: 'No account found with this email address'
        },
        {
            name: 'Login with correct email but wrong password',
            endpoint: '/auth/login',
            method: 'POST',
            data: {
                email: 'test@example.com', // Replace with existing email in your DB
                password: 'wrongpassword'
            },
            expectedError: 'Incorrect password. Please try again'
        },
        {
            name: 'Login with empty email',
            endpoint: '/auth/login',
            method: 'POST',
            data: {
                email: '',
                password: 'password123'
            },
            expectedError: 'Email is required'
        },
        {
            name: 'Login with empty password',
            endpoint: '/auth/login',
            method: 'POST',
            data: {
                email: 'test@example.com',
                password: ''
            },
            expectedError: 'Password is required'
        },
        {
            name: 'Login with invalid email format',
            endpoint: '/auth/login',
            method: 'POST',
            data: {
                email: 'invalid-email',
                password: 'password123'
            },
            expectedError: 'Invalid email format'
        },
        {
            name: 'Signup with password less than 8 characters',
            endpoint: '/auth/signup',
            method: 'POST',
            data: {
                name: 'Test User',
                username: 'testuser',
                email: 'newuser@example.com',
                password: '123456', // Less than 8 characters
                confirmPassword: '123456',
                role: 'host'
            },
            expectedError: 'Password must be at least 8 characters long'
        },
        {
            name: 'Signup with mismatched passwords',
            endpoint: '/auth/signup',
            method: 'POST',
            data: {
                name: 'Test User',
                username: 'testuser',
                email: 'newuser@example.com',
                password: 'password123',
                confirmPassword: 'differentpassword',
                role: 'host'
            },
            expectedError: 'Passwords do not match'
        },
        {
            name: 'Signup with existing email',
            endpoint: '/auth/signup',
            method: 'POST',
            data: {
                name: 'Test User',
                username: 'newusername',
                email: 'test@example.com', // Use existing email
                password: 'password123',
                confirmPassword: 'password123',
                role: 'host'
            },
            expectedError: 'Email already exists'
        },
        {
            name: 'Signup with invalid email format',
            endpoint: '/auth/signup',
            method: 'POST',
            data: {
                name: 'Test User',
                username: 'testuser',
                email: 'invalid-email-format',
                password: 'password123',
                confirmPassword: 'password123',
                role: 'host'
            },
            expectedError: 'Invalid email format'
        },
        {
            name: 'Signup with empty required fields',
            endpoint: '/auth/signup',
            method: 'POST',
            data: {
                name: '',
                username: '',
                email: '',
                password: '',
                confirmPassword: '',
                role: 'host'
            },
            expectedError: 'Name is required' // First validation error
        }
    ];

    let passedTests = 0;
    let failedTests = 0;

    for (const test of tests) {
        console.log(`\n🔬 Testing: ${test.name}`);
        console.log(`📡 ${test.method} ${API_BASE}${test.endpoint}`);
        console.log(`📝 Data:`, JSON.stringify(test.data, null, 2));
        console.log(`🎯 Expected Error: "${test.expectedError}"`);

        try {
            const response = await fetch(`${API_BASE}${test.endpoint}`, {
                method: test.method,
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(test.data)
            });

            const result = await response.json();
            
            console.log(`📊 Status: ${response.status}`);
            console.log(`📋 Response:`, JSON.stringify(result, null, 2));

            if (!response.ok && result.error) {
                if (result.error === test.expectedError) {
                    console.log(`✅ PASS: Got expected error message`);
                    passedTests++;
                } else {
                    console.log(`❌ FAIL: Expected "${test.expectedError}", got "${result.error}"`);
                    failedTests++;
                }
            } else {
                console.log(`❌ FAIL: Expected error but got success or no error message`);
                failedTests++;
            }

        } catch (error) {
            console.log(`❌ NETWORK ERROR: ${error.message}`);
            failedTests++;
        }

        console.log('-'.repeat(50));
    }

    console.log(`\n📈 TEST SUMMARY:`);
    console.log(`✅ Passed: ${passedTests}`);
    console.log(`❌ Failed: ${failedTests}`);
    console.log(`📊 Total: ${passedTests + failedTests}`);
    console.log(`🎯 Success Rate: ${((passedTests / (passedTests + failedTests)) * 100).toFixed(1)}%`);

    if (failedTests === 0) {
        console.log(`\n🎉 All tests passed! Error handling is working correctly.`);
    } else {
        console.log(`\n⚠️  Some tests failed. Please check the backend implementation.`);
    }
}

// Run the tests
testAuthErrors().catch(console.error);