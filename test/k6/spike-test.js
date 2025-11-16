import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Spike test configuration - sudden increase in traffic
export const options = {
  stages: [
    { duration: '1m', target: 10 },    // Normal load
    { duration: '10s', target: 500 },  // SPIKE! Sudden increase
    { duration: '3m', target: 500 },   // Maintain spike
    { duration: '1m', target: 10 },    // Scale down to normal
    { duration: '1m', target: 0 },     // Ramp down to 0
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'],  // 95% should be below 2s during spike
    http_req_failed: ['rate<0.3'],      // Accept higher error rate during spike
  },
};

const BASE_URL = __ENV.API_URL || 'http://localhost:8080';

export function setup() {
  console.log('Starting spike test - simulating sudden traffic increase...');
  return {};
}

export default function() {
  const randomId = `${Date.now()}-${__VU}-${__ITER}`;
  const email = `spike-${randomId}@example.com`;

  // Quick registration flow
  const registerRes = http.post(`${BASE_URL}/api/v1/auth/register`, JSON.stringify({
    name: `Spike User ${randomId}`,
    email: email,
    password: 'password123',
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  const registerSuccess = check(registerRes, {
    'register successful': (r) => r.status === 200 || r.status === 201,
    'response time OK': (r) => r.timings.duration < 3000,
  });

  if (!registerSuccess) {
    errorRate.add(1);
    sleep(0.1);
    return;
  }

  sleep(0.05);

  // Login to get token
  const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({
    email: email,
    password: 'password123',
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  const loginSuccess = check(loginRes, {
    'login successful': (r) => r.status === 200,
    'has token': (r) => r.json('data.token') !== undefined && r.json('data.token') !== null,
  });

  if (!loginSuccess) {
    errorRate.add(1);
    sleep(0.1);
    return;
  }

  const token = loginRes.json('data.token');

  // Validate token exists before proceeding
  if (!token) {
    errorRate.add(1);
    sleep(0.1);
    return;
  }

  // Quick profile access
  const profileRes = http.get(`${BASE_URL}/api/v1/profile`, {
    headers: { 'Authorization': `Bearer ${token}` },
  });

  check(profileRes, {
    'profile accessible': (r) => r.status === 200,
  }) || errorRate.add(1);

  sleep(0.1); // Minimal sleep during spike
}

export function teardown(data) {
  console.log('Spike test completed');
}
