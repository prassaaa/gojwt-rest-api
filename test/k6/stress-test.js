import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Stress test configuration - push the system beyond normal capacity
export const options = {
  stages: [
    { duration: '2m', target: 100 },   // Ramp up to 100 users
    { duration: '5m', target: 100 },   // Stay at 100 users
    { duration: '2m', target: 200 },   // Ramp up to 200 users
    { duration: '5m', target: 200 },   // Stay at 200 users
    { duration: '2m', target: 300 },   // Ramp up to 300 users (stress)
    { duration: '5m', target: 300 },   // Stay at 300 users
    { duration: '5m', target: 0 },     // Ramp down to 0
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'],  // 95% of requests should be below 1s
    http_req_failed: ['rate<0.2'],      // Error rate should be less than 20%
  },
};

const BASE_URL = __ENV.API_URL || 'http://localhost:8080';

export function setup() {
  console.log('Starting stress test...');
  return {};
}

export default function() {
  // Scenario: High load authentication flow
  const randomId = `${Date.now()}-${Math.random()}`;
  const email = `stress-user-${randomId}@example.com`;
  const password = 'password123';

  // Register
  const registerRes = http.post(`${BASE_URL}/api/v1/auth/register`, JSON.stringify({
    name: `Stress Test User ${randomId}`,
    email: email,
    password: password,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  const registerSuccess = check(registerRes, {
    'register status is 200': (r) => r.status === 200 || r.status === 201,
  });

  if (!registerSuccess) {
    errorRate.add(1);
    sleep(0.5);
    return;
  }

  sleep(0.2);

  // Login to get token
  const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({
    email: email,
    password: password,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  const loginSuccess = check(loginRes, {
    'login status is 200': (r) => r.status === 200,
    'login has token': (r) => r.json('data.token') !== undefined && r.json('data.token') !== null,
  });

  if (!loginSuccess) {
    errorRate.add(1);
    sleep(0.5);
    return;
  }

  const token = loginRes.json('data.token');

  // Validate token exists before proceeding
  if (!token) {
    errorRate.add(1);
    sleep(0.5);
    return;
  }

  sleep(0.1);

  // Get profile multiple times (simulate real user behavior)
  for (let i = 0; i < 3; i++) {
    const profileRes = http.get(`${BASE_URL}/api/v1/profile`, {
      headers: { 'Authorization': `Bearer ${token}` },
    });

    check(profileRes, {
      'profile status is 200': (r) => r.status === 200,
    }) || errorRate.add(1);

    sleep(0.2);
  }

  // Update profile
  const updateRes = http.put(`${BASE_URL}/api/v1/profile`, JSON.stringify({
    name: `Updated ${randomId}`,
  }), {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  });

  check(updateRes, {
    'update status is 200': (r) => r.status === 200,
  }) || errorRate.add(1);

  sleep(0.5);
}

export function teardown(data) {
  console.log('Stress test completed');
}
