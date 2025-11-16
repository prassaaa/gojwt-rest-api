import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export const options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up to 10 users over 30s
    { duration: '1m', target: 10 },   // Stay at 10 users for 1m
    { duration: '30s', target: 50 },  // Ramp up to 50 users over 30s
    { duration: '2m', target: 50 },   // Stay at 50 users for 2m
    { duration: '30s', target: 0 },   // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
    http_req_failed: ['rate<0.1'],    // Error rate should be less than 10%
    errors: ['rate<0.1'],
  },
};

const BASE_URL = __ENV.API_URL || 'http://localhost:8080';

export function setup() {
  // Register a test user for the load test
  const email = `loadtest-${Date.now()}@example.com`;
  const password = 'password123';

  const registerRes = http.post(`${BASE_URL}/api/v1/auth/register`, JSON.stringify({
    name: 'Load Test User',
    email: email,
    password: password,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  if (registerRes.status !== 200 && registerRes.status !== 201) {
    console.error(`Setup failed: ${registerRes.status} - ${registerRes.body}`);
    return { email, password, token: null };
  }

  // Login to get token
  const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({
    email: email,
    password: password,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  const token = loginRes.json('data.token');

  return { email, password, token };
}

export default function(data) {
  // Test 1: Health check
  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'health check status is 200': (r) => r.status === 200,
  }) || errorRate.add(1);

  sleep(1);

  // Test 2: Register new user
  const randomEmail = `user-${Date.now()}-${Math.random()}@example.com`;
  const registerRes = http.post(`${BASE_URL}/api/v1/auth/register`, JSON.stringify({
    name: 'Test User',
    email: randomEmail,
    password: 'password123',
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  const registerCheck = check(registerRes, {
    'register status is 200 or 201': (r) => r.status === 200 || r.status === 201,
    'register response has token': (r) => r.json('data.token') !== undefined,
  });

  if (!registerCheck) {
    errorRate.add(1);
    return;
  }

  const newToken = registerRes.json('data.token');
  sleep(1);

  // Test 3: Get own profile
  const profileRes = http.get(`${BASE_URL}/api/v1/profile`, {
    headers: {
      'Authorization': `Bearer ${newToken}`,
      'Content-Type': 'application/json',
    },
  });

  check(profileRes, {
    'get profile status is 200': (r) => r.status === 200,
    'profile has user data': (r) => r.json('data.email') === randomEmail,
  }) || errorRate.add(1);

  sleep(1);

  // Test 4: Update profile
  const updateRes = http.put(`${BASE_URL}/api/v1/profile`, JSON.stringify({
    name: 'Updated User Name',
  }), {
    headers: {
      'Authorization': `Bearer ${newToken}`,
      'Content-Type': 'application/json',
    },
  });

  check(updateRes, {
    'update profile status is 200': (r) => r.status === 200,
    'profile name updated': (r) => r.json('data.name') === 'Updated User Name',
  }) || errorRate.add(1);

  sleep(1);

  // Test 5: Change password
  const changePasswordRes = http.put(`${BASE_URL}/api/v1/profile/password`, JSON.stringify({
    old_password: 'password123',
    new_password: 'newpassword123',
  }), {
    headers: {
      'Authorization': `Bearer ${newToken}`,
      'Content-Type': 'application/json',
    },
  });

  check(changePasswordRes, {
    'change password status is 200': (r) => r.status === 200,
  }) || errorRate.add(1);

  sleep(2);
}

export function teardown(data) {
  console.log('Load test completed');
}
