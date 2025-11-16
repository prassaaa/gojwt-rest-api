import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Soak test configuration - test reliability over extended period
export const options = {
  stages: [
    { duration: '2m', target: 50 },    // Ramp up to 50 users
    { duration: '30m', target: 50 },   // Stay at 50 users for 30 minutes
    { duration: '2m', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<800'],   // 95% should be below 800ms
    http_req_failed: ['rate<0.05'],     // Error rate should be below 5%
    errors: ['rate<0.05'],
  },
};

const BASE_URL = __ENV.API_URL || 'http://localhost:8080';

export function setup() {
  console.log('Starting soak test - testing system reliability over 30 minutes...');
  return {};
}

export default function() {
  const randomId = `${Date.now()}-${__VU}-${__ITER}`;
  const email = `soak-${randomId}@example.com`;
  const password = 'password123';

  // Complete user flow

  // 1. Register
  let res = http.post(`${BASE_URL}/api/v1/auth/register`, JSON.stringify({
    name: `Soak User ${randomId}`,
    email: email,
    password: password,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  let success = check(res, {
    'register successful': (r) => r.status === 200 || r.status === 201,
  });

  if (!success) {
    errorRate.add(1);
    sleep(5);
    return;
  }

  sleep(1);

  // Login to get token
  res = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({
    email: email,
    password: password,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  success = check(res, {
    'login successful': (r) => r.status === 200,
    'has token': (r) => r.json('data.token') !== undefined && r.json('data.token') !== null,
  });

  if (!success) {
    errorRate.add(1);
    sleep(5);
    return;
  }

  const token = res.json('data.token');

  // Validate token exists before proceeding
  if (!token) {
    errorRate.add(1);
    sleep(5);
    return;
  }

  sleep(1);

  // 2. Get profile (multiple times to simulate browsing)
  for (let i = 0; i < 2; i++) {
    res = http.get(`${BASE_URL}/api/v1/profile`, {
      headers: { 'Authorization': `Bearer ${token}` },
    });

    check(res, {
      'get profile OK': (r) => r.status === 200,
    }) || errorRate.add(1);

    sleep(3);
  }

  // 3. Update profile
  res = http.put(`${BASE_URL}/api/v1/profile`, JSON.stringify({
    name: `Updated Soak User ${randomId}`,
  }), {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  });

  check(res, {
    'update profile OK': (r) => r.status === 200,
  }) || errorRate.add(1);

  sleep(4);

  // 4. Health check
  res = http.get(`${BASE_URL}/health`);
  check(res, {
    'health check OK': (r) => r.status === 200,
  }) || errorRate.add(1);

  sleep(5);
}

export function teardown(data) {
  console.log('Soak test completed - check for memory leaks and degradation');
}
