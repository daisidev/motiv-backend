#!/usr/bin/env node

// Test script to verify newsletter subscription functionality
// Run this after starting the backend: node test-newsletter-subscription.js

const API_BASE_URL = process.env.API_URL || 'http://localhost:8080/api/v1';

async function testNewsletterSubscription() {
  console.log('🧪 Testing Newsletter Subscription Functionality\n');

  // Test data
  const testUser = {
    name: 'Newsletter Test User',
    username: 'newsletter_test_' + Date.now(),
    email: 'newsletter.test@example.com',
    password: 'password123',
    confirmPassword: 'password123',
    role: 'host',
    newsletterSubscribed: true // This should be handled properly
  };

  let authToken = '';
  let userId = '';

  try {
    // Step 1: Test signup with newsletter subscription
    console.log('📝 Step 1: Testing signup with newsletter subscription...');
    const signupResponse = await fetch(`${API_BASE_URL}/auth/signup`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(testUser),
    });

    const signupData = await signupResponse.json();
    
    if (!signupResponse.ok) {
      console.log('❌ Signup failed:', signupData);
      return;
    }

    authToken = signupData.token;
    userId = signupData.user.id;
    
    console.log('✅ Signup successful');
    console.log('   User ID:', userId);
    console.log('   Newsletter Subscribed:', signupData.user.newsletter_subscribed);
    
    if (signupData.user.newsletter_subscribed) {
      console.log('✅ Newsletter subscription was set correctly during signup');
    } else {
      console.log('❌ Newsletter subscription was NOT set during signup');
    }

    // Step 2: Test profile update to enable newsletter
    console.log('\n📝 Step 2: Testing profile update to enable newsletter...');
    const updateResponse = await fetch(`${API_BASE_URL}/users/me`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${authToken}`,
      },
      body: JSON.stringify({
        NewsletterSubscribed: true
      }),
    });

    const updateData = await updateResponse.json();
    
    if (!updateResponse.ok) {
      console.log('❌ Profile update failed:', updateData);
    } else {
      console.log('✅ Profile update successful');
      console.log('   Newsletter Subscribed:', updateData.newsletter_subscribed);
      
      if (updateData.newsletter_subscribed) {
        console.log('✅ Newsletter subscription was updated correctly');
      } else {
        console.log('❌ Newsletter subscription was NOT updated');
      }
    }

    // Step 3: Test getting user profile to verify newsletter status
    console.log('\n📝 Step 3: Testing user profile retrieval...');
    const profileResponse = await fetch(`${API_BASE_URL}/users/me`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${authToken}`,
      },
    });

    const profileData = await profileResponse.json();
    
    if (!profileResponse.ok) {
      console.log('❌ Profile retrieval failed:', profileData);
    } else {
      console.log('✅ Profile retrieval successful');
      console.log('   User ID:', profileData.id);
      console.log('   Name:', profileData.name);
      console.log('   Email:', profileData.email);
      console.log('   Newsletter Subscribed:', profileData.newsletter_subscribed);
      
      if (profileData.newsletter_subscribed) {
        console.log('✅ Newsletter subscription persisted correctly');
      } else {
        console.log('❌ Newsletter subscription was NOT persisted');
      }
    }

    // Step 4: Test disabling newsletter subscription
    console.log('\n📝 Step 4: Testing newsletter unsubscription...');
    const unsubscribeResponse = await fetch(`${API_BASE_URL}/users/me`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${authToken}`,
      },
      body: JSON.stringify({
        NewsletterSubscribed: false
      }),
    });

    const unsubscribeData = await unsubscribeResponse.json();
    
    if (!unsubscribeResponse.ok) {
      console.log('❌ Newsletter unsubscription failed:', unsubscribeData);
    } else {
      console.log('✅ Newsletter unsubscription successful');
      console.log('   Newsletter Subscribed:', unsubscribeData.newsletter_subscribed);
      
      if (!unsubscribeData.newsletter_subscribed) {
        console.log('✅ Newsletter unsubscription was processed correctly');
      } else {
        console.log('❌ Newsletter unsubscription was NOT processed');
      }
    }

    console.log('\n🎉 Newsletter subscription testing completed!');

  } catch (error) {
    console.error('❌ Test failed with error:', error.message);
  }
}

// Run the test
testNewsletterSubscription();