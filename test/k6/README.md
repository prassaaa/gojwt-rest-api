# K6 Performance Testing

Dokumentasi lengkap untuk performance testing menggunakan K6.

## Apa itu K6?

K6 adalah modern load testing tool yang ditulis dalam Go, dengan scripting menggunakan JavaScript. K6 dirancang untuk testing performa aplikasi modern dan API.

## Prerequisites

- K6 sudah terinstall di sistem Anda
- API server berjalan di `http://localhost:8080` (atau set `API_URL` environment variable)

```bash
# Verifikasi instalasi K6
k6 version
```

## Test Scripts

### 1. Smoke Test (`smoke-test.js`)

**Tujuan**: Verifikasi minimal bahwa sistem berfungsi dengan baik

**Karakteristik**:
- 1 virtual user
- Duration: 1 menit
- Tests: Semua endpoint dasar (health, register, login, profile)

**Kapan digunakan**:
- Setelah deploy baru
- Sebelum menjalankan load test yang lebih berat
- Untuk CI/CD pipeline

**Cara menjalankan**:
```bash
k6 run test/k6/smoke-test.js
```

**Contoh output yang baik**:
```
✓ health check OK
✓ register OK
✓ has token
✓ login OK
✓ get profile OK
✓ profile has email
✓ update profile OK
✓ change password OK

checks.........................: 100.00% ✓ 480  ✗ 0
http_req_duration..............: avg=150ms min=50ms med=120ms max=500ms p(95)=300ms
http_reqs......................: 480     8/s
```

### 2. Load Test (`load-test.js`)

**Tujuan**: Mengukur performa sistem dengan beban normal hingga tinggi

**Karakteristik**:
- Ramp up: 10 → 50 users over 4 minutes
- Steady state: 50 users for 2 minutes
- Tests: Full user flow (register, profile, update, change password)

**Thresholds**:
- 95% requests < 500ms
- Error rate < 10%

**Cara menjalankan**:
```bash
k6 run test/k6/load-test.js
```

**Metrics yang diperhatikan**:
- `http_req_duration`: Response time (target: p95 < 500ms)
- `http_req_failed`: Failed request rate (target: < 10%)
- `errors`: Custom error metric

### 3. Stress Test (`stress-test.js`)

**Tujuan**: Menemukan batas sistem dengan meningkatkan load secara bertahap

**Karakteristik**:
- Progressive load: 100 → 200 → 300 users
- Total duration: ~20 menit
- Tests: Authentication dan profile management flow

**Thresholds**:
- 95% requests < 1000ms (lebih toleran)
- Error rate < 20% (lebih toleran)

**Cara menjalankan**:
```bash
k6 run test/k6/stress-test.js
```

**Yang dicari**:
- Pada user count berapa sistem mulai degradasi?
- Apakah sistem recover setelah load turun?
- Apakah ada memory leaks?
- Bagaimana database performance di bawah tekanan?

### 4. Spike Test (`spike-test.js`)

**Tujuan**: Test bagaimana sistem handle sudden traffic spike

**Karakteristik**:
- Normal: 10 users
- **SPIKE**: Sudden jump ke 500 users dalam 10 detik!
- Maintain spike: 3 menit
- Recovery: Turun ke 10 users

**Thresholds**:
- 95% requests < 2000ms (sangat toleran)
- Error rate < 30% (sangat toleran)

**Cara menjalankan**:
```bash
k6 run test/k6/spike-test.js
```

**Use cases**:
- Viral content / sudden popularity
- Marketing campaign launch
- Black Friday / flash sale scenarios

### 5. Soak Test (`soak-test.js`)

**Tujuan**: Test reliability dan stability dalam periode waktu lama

**Karakteristik**:
- 50 concurrent users
- Duration: **30 menit**
- Tests: Full user flow dengan realistic delays

**Thresholds**:
- 95% requests < 800ms
- Error rate < 5%

**Cara menjalankan**:
```bash
k6 run test/k6/soak-test.js
```

**Yang dicari**:
- Memory leaks
- Database connection leaks
- Performance degradation over time
- Resource exhaustion

## Environment Variables

Semua test scripts mendukung custom API URL:

```bash
# Test terhadap server production
k6 run -e API_URL=https://api.production.com test/k6/load-test.js

# Test terhadap staging
k6 run -e API_URL=https://api.staging.com test/k6/smoke-test.js
```

## Output Options

### Console Output (default)
```bash
k6 run test/k6/load-test.js
```

### JSON Output
```bash
k6 run --out json=results.json test/k6/load-test.js
```

### InfluxDB (untuk monitoring real-time)
```bash
k6 run --out influxdb=http://localhost:8086/k6 test/k6/load-test.js
```

### HTML Report (dengan xk6-reporter)
```bash
k6 run --out json=results.json test/k6/load-test.js
# Convert to HTML with k6-reporter atau k6-html-reporter
```

## Interpretasi Results

### Metrics Penting

1. **http_req_duration**
   - `avg`: Average response time
   - `min/max`: Fastest/slowest request
   - `p(95)`: 95% requests lebih cepat dari nilai ini
   - `p(99)`: 99% requests lebih cepat dari nilai ini

2. **http_req_failed**
   - Rate dari failed requests
   - Target: < 0.01 (1%) untuk production

3. **http_reqs**
   - Total requests dan requests/second
   - Throughput sistem

4. **vus**
   - Virtual users aktif
   - Peak concurrent users

5. **iteration_duration**
   - Waktu untuk menyelesaikan satu iteration complete test

### Contoh Good Results

```
scenarios: (100.00%) 1 scenario, 50 max VUs, 4m30s max duration

✓ register successful
✓ profile accessible
✓ update profile OK

checks.........................: 100.00% ✓ 15000  ✗ 0
data_received..................: 45 MB   187 kB/s
data_sent......................: 12 MB   50 kB/s
http_req_blocked...............: avg=1ms     min=0s   med=0s   max=100ms  p(95)=5ms
http_req_connecting............: avg=500µs   min=0s   med=0s   max=50ms   p(95)=2ms
http_req_duration..............: avg=150ms   min=10ms med=120ms max=800ms p(95)=350ms
  { expected_response:true }...: avg=150ms   min=10ms med=120ms max=800ms p(95)=350ms
http_req_failed................: 0.00%   ✓ 0      ✗ 15000
http_req_receiving.............: avg=500µs   min=0s   med=300µs max=50ms  p(95)=2ms
http_req_sending...............: avg=300µs   min=0s   med=200µs max=30ms  p(95)=1ms
http_req_tls_handshaking.......: avg=0s      min=0s   med=0s    max=0s    p(95)=0s
http_req_waiting...............: avg=149ms   min=10ms med=119ms max=799ms p(95)=349ms
http_reqs......................: 15000   62.5/s
iteration_duration.............: avg=1.5s    min=1s   med=1.4s  max=3s    p(95)=2s
iterations.....................: 5000    20.83/s
vus............................: 50      min=10   max=50
vus_max........................: 50      min=50   max=50
```

### Red Flags 🚩

1. **High error rate** (> 5%)
   - Check server logs
   - Database connections
   - Rate limiting configuration

2. **High p(95) response time** (> 1s)
   - Database query optimization
   - N+1 query problems
   - Missing indexes

3. **Increasing response time over time**
   - Memory leaks
   - Connection pool exhaustion
   - Cache not working

4. **High http_req_blocked time**
   - DNS issues
   - Connection pool saturation
   - Network latency

## Best Practices

### 1. Test Progression
Jalankan tests dalam urutan ini:
```bash
# Step 1: Smoke test dulu
k6 run test/k6/smoke-test.js

# Step 2: Load test
k6 run test/k6/load-test.js

# Step 3: Stress test
k6 run test/k6/stress-test.js

# Step 4: Spike test
k6 run test/k6/spike-test.js

# Step 5: Soak test (optional, untuk stability testing)
k6 run test/k6/soak-test.js
```

### 2. Monitor Server
Saat menjalankan K6 tests, monitor juga:
- CPU usage
- Memory usage
- Database connections
- Network I/O
- Disk I/O

```bash
# Terminal 1: Run K6 test
k6 run test/k6/load-test.js

# Terminal 2: Monitor dengan htop
htop

# Terminal 3: Monitor database
mysql -u root -p -e "SHOW PROCESSLIST;"
```

### 3. Realistic Test Data
- Gunakan variasi data yang realistis
- Simulate different user behaviors
- Include think time (sleep) yang realistic

### 4. Clean Up
Setelah test selesai, bersihkan test data:
```sql
-- Clean up test users
DELETE FROM users WHERE email LIKE 'loadtest-%';
DELETE FROM users WHERE email LIKE 'stress-%';
DELETE FROM users WHERE email LIKE 'spike-%';
DELETE FROM users WHERE email LIKE 'soak-%';
```

## Troubleshooting

### Error: "connection refused"
```bash
# Pastikan API server berjalan
curl http://localhost:8080/health

# Atau set custom URL
k6 run -e API_URL=http://localhost:8080 test/k6/smoke-test.js
```

### Error: "too many open files"
```bash
# Linux: Increase file descriptor limit
ulimit -n 10000
```

### High error rate during tests
1. Check rate limiting configuration di `.env`
2. Increase database connection pool
3. Check server logs untuk errors

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Performance Tests

on: [push, pull_request]

jobs:
  performance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install K6
        run: |
          sudo gpg -k
          sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
          echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
          sudo apt-get update
          sudo apt-get install k6

      - name: Start API Server
        run: |
          # Start your API in background
          make run &
          sleep 5

      - name: Run Smoke Test
        run: k6 run test/k6/smoke-test.js

      - name: Run Load Test
        run: k6 run test/k6/load-test.js
```

## Resources

- [K6 Documentation](https://k6.io/docs/)
- [K6 Examples](https://github.com/grafana/k6/tree/master/examples)
- [Performance Testing Checklist](https://k6.io/docs/testing-guides/load-testing-checklist/)

## Summary

| Test Type | Duration | VUs | Purpose |
|-----------|----------|-----|---------|
| Smoke     | 1m       | 1   | Basic functionality |
| Load      | 4.5m     | 10-50 | Normal to high load |
| Stress    | 20m      | 100-300 | Find breaking point |
| Spike     | 6.5m     | 10-500 | Sudden traffic spike |
| Soak      | 34m      | 50  | Long-term stability |

**Recommendation**: Jalankan smoke test setiap deploy, load test setiap minggu, dan stress/soak test sebelum major releases.
