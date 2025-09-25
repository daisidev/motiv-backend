// Enhanced test script to check password reset functionality
const fetch = require('node-fetch');

async function testPasswordReset() {
    console.log('=== TESTING PASSWORD RESET FUNCTIONALITY ===');
    
    // Test with a real email that exists in your system
    const testEmail = 'test@example.com'; // Replace with a real email from your database
    
    try {
        console.log(`\n🔄 Testing forgot-password endpoint...`);
        console.log(`📧 Using email: ${testEmail}`);
        console.log(`🌐 API URL: http://localhost:8080/api/v1/auth/forgot-password`);
        
        const response = await fetch('http://localhost:8080/api/v1/auth/forgot-password', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                email: testEmail
            })
        });

        console.log(`\n📊 Response Status: ${response.status}`);
        console.log(`📊 Response Status Text: ${response.statusText}`);
        
        const result = await response.json();
        console.log(`📋 Response Data:`, JSON.stringify(result, null, 2));
        
        if (response.ok) {
            console.log(`\n✅ SUCCESS: Password reset request processed successfully!`);
            console.log(`💌 Check the email service logs to see if email was sent`);
        } else {
            console.log(`\n❌ ERROR: Password reset failed`);
            console.log(`🔍 Error Details: ${result.error || 'Unknown error'}`);
        }
        
    } catch (error) {
        console.error(`\n❌ NETWORK ERROR: Failed to connect to backend`);
        console.error(`🔍 Error Details:`, error.message);
        console.error(`\n💡 Possible issues:`);
        console.error(`   - Backend server is not running`);
        console.error(`   - Wrong API URL`);
        console.error(`   - Network connectivity issues`);
    }
    
    console.log(`\n=== TEST COMPLETED ===`);
}

// Test with different scenarios
async function runAllTests() {
    console.log('🚀 Starting comprehensive password reset tests...\n');
    
    // Test 1: Valid email format
    await testPasswordReset();
    
    // Test 2: Invalid email format
    console.log('\n' + '='.repeat(50));
    console.log('Testing with invalid email format...');
    try {
        const response = await fetch('http://localhost:8080/api/v1/auth/forgot-password', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                email: 'invalid-email'
            })
        });
        
        const result = await response.json();
        console.log(`Status: ${response.status}, Response:`, result);
        
    } catch (error) {
        console.error('Network error:', error.message);
    }
    
    // Test 3: Empty email
    console.log('\n' + '='.repeat(50));
    console.log('Testing with empty email...');
    try {
        const response = await fetch('http://localhost:8080/api/v1/auth/forgot-password', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                email: ''
            })
        });
        
        const result = await response.json();
        console.log(`Status: ${response.status}, Response:`, result);
        
    } catch (error) {
        console.error('Network error:', error.message);
    }
    
    console.log('\n🏁 All tests completed!');
}

runAllTests();