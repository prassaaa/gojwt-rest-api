import http from 'k6/http';
import { check, sleep } from 'k6';

// Smoke test configuration - minimal load to verify system works
export const options = {
  vus: 1,           // 1 virtual user
  duration: '1m',   // Run for 1 minute
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
  },
};

const BASE_URL = __ENV.API_URL || 'http://localhost:8080';

export default function() {
  // Test 1: Health check
  let res = http.get(`${BASE_URL}/health`);
  check(res, {
    'health check OK': (r) => r.status === 200,
  });
  sleep(1);

  // Test 2: Welcome endpoint
  res = http.get(`${BASE_URL}/`);
  check(res, {
    'welcome endpoint OK': (r) => r.status === 200,
    'has version info': (r) => r.json('version') !== undefined,
  });
  sleep(1);

  // Test 3: Register
  const email = `smoke-${Date.now()}@example.com`;
  res = http.post(`${BASE_URL}/api/v1/auth/register`, JSON.stringify({
    name: 'Smoke Test User',
    email: email,
    password: 'password123',
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  check(res, {
    'register OK': (r) => r.status === 200 || r.status === 201,
    'has user data': (r) => r.json('data.email') !== undefined,
  });
  sleep(1);

  // Test 4: Login
  res = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({
    email: email,
    password: 'password123',
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  check(res, {
    'login OK': (r) => r.status === 200,
    'has token': (r) => r.json('data.token') !== undefined,
  });

  const token = res.json('data.token');
  sleep(1);

  // Test 5: Get profile
  res = http.get(`${BASE_URL}/api/v1/profile`, {
    headers: { 'Authorization': `Bearer ${token}` },
  });

  check(res, {
    'get profile OK': (r) => r.status === 200,
    'profile has email': (r) => r.json('data.email') === email,
  });
  sleep(1);

  // Test 6: Update profile
  res = http.put(`${BASE_URL}/api/v1/profile`, JSON.stringify({
    name: 'Smoke Test Updated',
  }), {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  });

  check(res, {
    'update profile OK': (r) => r.status === 200,
  });
  sleep(1);

  // Test 7: Change password
  res = http.put(`${BASE_URL}/api/v1/profile/password`, JSON.stringify({
    old_password: 'password123',
    new_password: 'newpassword456',
  }), {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  });

  check(res, {
    'change password OK': (r) => r.status === 200,
  });
  sleep(2);
}
