# Dokumentasi Testing

Dokumen ini menjelaskan strategi testing dan cara menjalankan test untuk project Go JWT REST API.

## Struktur Test

Project ini menggunakan arsitektur test yang bersih dengan test yang diorganisir dalam direktori `test/` terpisah:

```
test/
├── unit/              # Unit tests (utils, service layer)
│   ├── jwt_test.go
│   ├── password_test.go
│   └── user_service_test.go
├── integration/       # Integration tests (repository dengan database mock)
│   └── user_repository_test.go
├── e2e/              # End-to-end tests (full HTTP request/response cycle)
│   └── auth_handler_test.go
└── helpers/          # Test utilities dan mocks
    ├── mock_repository.go
    └── test_data.go
```

## Cakupan Test

### Unit Tests
- **JWT Utils** (`test/unit/jwt_test.go`): 14 test case
  - Pembuatan token dengan berbagai parameter
  - Validasi token (valid, expired, invalid secret, malformed)
  - Verifikasi claims
  - Validasi signing method
  - Benchmark untuk operasi token

- **Password Utils** (`test/unit/password_test.go`): 10 test case
  - Password hashing dengan berbagai input
  - Verifikasi password
  - Keunikan bcrypt salt
  - Edge cases (empty, sangat panjang, unicode passwords)
  - Benchmark untuk operasi hash/check

- **User Service** (`test/unit/user_service_test.go`): 16 test case
  - Registrasi user (sukses, email duplikat, validation errors)
  - Login user (sukses, kredensial salah)
  - Get user by ID
  - Get all users dengan pagination
  - Update user (name, email, validation)
  - Delete user

### Integration Tests
- **User Repository** (`test/integration/user_repository_test.go`): 11 test case
  - Create user
  - Find by ID dan email
  - Find all dengan pagination dan search
  - Update user
  - Delete user
  - Error handling (database errors, not found)

### E2E Tests
- **Auth Handler** (`test/e2e/auth_handler_test.go`): 11 test case
  - Register endpoint (sukses, validation errors, email duplikat)
  - Login endpoint (sukses, kredensial salah, validation)
  - Testing full HTTP request/response

## Menjalankan Test

### Prerequisites
```bash
# Install dependencies test
go get github.com/stretchr/testify@latest
go get github.com/DATA-DOG/go-sqlmock@latest
go mod tidy
```

### Jalankan Semua Test
```bash
make test
```

### Jalankan Test Suite Spesifik
```bash
# Unit tests saja
make test-unit

# Integration tests saja
make test-integration

# E2E tests saja
make test-e2e
```

### Jalankan Test dengan Coverage
```bash
# Coverage report di terminal
make test-coverage

# Generate HTML coverage report
make test-coverage-html
# Kemudian buka coverage.html di browser
```

### Jalankan Benchmark Tests
```bash
make test-bench
```

### Jalankan Test Spesifik
```bash
# Jalankan test tertentu berdasarkan nama
go test -v ./test/unit -run TestGenerateToken

# Jalankan test dalam file tertentu
go test -v ./test/unit/jwt_test.go
```

## Test Helpers

### Mock Repository
Terletak di `test/helpers/mock_repository.go`, menyediakan implementasi mock dari interface `UserRepository` menggunakan testify/mock.

**Cara Pakai:**
```go
mockRepo := new(helpers.MockUserRepository)
mockRepo.On("FindByEmail", "test@example.com").Return(user, nil)
```

### Test Data Builders
Terletak di `test/helpers/test_data.go`, menyediakan fungsi helper untuk membuat test data:

- `CreateTestUser(id, email)` - Buat test user biasa
- `CreateAdminUser(id, email)` - Buat admin test user
- `CreateRegisterRequest(name, email, password)` - Buat registration request
- `CreateLoginRequest(email, password)` - Buat login request
- `CreateUpdateUserRequest(name, email)` - Buat update request
- `CreatePaginationQuery(page, pageSize, search)` - Buat pagination query

## Pola Test

### Table-Driven Tests
Kebanyakan test menggunakan pendekatan table-driven untuk multiple test case:

```go
tests := []struct {
    name     string
    input    string
    expected string
    wantErr  bool
}{
    {name: "valid case", input: "test", expected: "result", wantErr: false},
    {name: "error case", input: "bad", expected: "", wantErr: true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        result, err := FunctionUnderTest(tt.input)
        // assertions...
    })
}
```

### Mock-Based Testing
Test service layer menggunakan mock untuk mengisolasi business logic:

```go
mockRepo := new(helpers.MockUserRepository)
userService := service.NewUserService(mockRepo, "secret", 24*time.Hour)

mockRepo.On("FindByEmail", "test@example.com").Return(user, nil)

result, err := userService.Login(request)

mockRepo.AssertExpectations(t)
```

### HTTP Testing
E2E test menggunakan `httptest` untuk testing handler:

```go
router := gin.New()
router.POST("/register", authHandler.Register)

req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
w := httptest.NewRecorder()

router.ServeHTTP(w, req)

assert.Equal(t, http.StatusCreated, w.Code)
```

## Hasil Benchmark

Contoh hasil benchmark pada AMD Ryzen 5 5600H:

```
BenchmarkGenerateToken-12         321196    3598 ns/op    2298 B/op    34 allocs/op
BenchmarkValidateToken-12         220212    5176 ns/op    2824 B/op    50 allocs/op
BenchmarkHashPassword-12              22    52.3 ms/op    5284 B/op    10 allocs/op
BenchmarkCheckPassword-12             22    52.4 ms/op    5220 B/op    11 allocs/op
```

## Continuous Integration

Test dapat diintegrasikan ke CI/CD pipeline:

```yaml
# Contoh GitHub Actions workflow
- name: Run tests
  run: make test

- name: Generate coverage
  run: make test-coverage

- name: Upload coverage
  uses: codecov/codecov-action@v3
  with:
    files: ./coverage.out
```

## Best Practices

1. **Isolation**: Setiap test harus independen dan tidak bergantung pada test lain
2. **Cleanup**: Gunakan cleanup function untuk reset state setelah test
3. **Descriptive Names**: Nama test harus jelas mendeskripsikan apa yang ditest
4. **AAA Pattern**: Struktur Arrange, Act, Assert dalam test
5. **Edge Cases**: Test boundary conditions dan error cases
6. **Mocking**: Gunakan mock untuk mengisolasi unit yang ditest
7. **Coverage**: Targetkan coverage tinggi tapi fokus pada test yang bermakna

## Troubleshooting

### Test Gagal Secara Lokal
1. Pastikan dependencies up to date: `go mod tidy`
2. Cek environment variables di `.env`
3. Clear test cache: `go clean -testcache`

### Test Lambat
- Jalankan test suite spesifik daripada semua test
- Gunakan flag `-short` untuk skip long-running tests
- Review hasil benchmark untuk identify operasi lambat

### Masalah Coverage
- Pastikan test ada di direktori `test/`
- Gunakan flag `-coverprofile` dengan path yang benar
- Cek bahwa file test memiliki suffix `_test.go`

## Perintah Tambahan

```bash
# Clean test cache
go clean -testcache

# Jalankan test dengan race detection
go test -race ./test/...

# Jalankan test verbose dengan colors
go test -v -cover ./test/...

# Jalankan test spesifik dengan timeout
go test -v -timeout 30s -run TestName ./test/unit/...

# List semua perintah make yang tersedia
make help
```

## Statistik Test

- **Total test cases**: 52
- **Total baris kode test**: ~1,887 baris
- **Test files**: 5 file
- **Success rate**: 100% (semua test PASS)

## Tips Testing

### Menjalankan Test Secara Efisien

1. **Development**: Gunakan `make test-unit` untuk feedback cepat
2. **Pre-commit**: Jalankan `make test` untuk full validation
3. **CI/CD**: Gunakan `make test-coverage` untuk track coverage

### Menulis Test Baru

1. Letakkan test di direktori yang sesuai (unit/integration/e2e)
2. Gunakan test helpers dari `test/helpers/`
3. Ikuti pola table-driven test
4. Tambahkan test untuk edge cases
5. Buat mock untuk external dependencies

### Debug Test yang Gagal

```bash
# Jalankan test dengan output verbose
go test -v ./test/unit/jwt_test.go

# Jalankan satu test case spesifik
go test -v ./test/unit -run TestGenerateToken/Valid_token_generation

# Lihat detail error dengan trace
go test -v -trace trace.out ./test/unit/...
```

## Integrasi dengan IDE

### VSCode
Install extension "Go" untuk:
- Run test langsung dari editor
- Debug test dengan breakpoint
- Coverage visualization

### GoLand
Built-in support untuk:
- Run/Debug test configuration
- Coverage analysis
- Test explorer

## Maintenance

### Update Test Dependencies
```bash
go get -u github.com/stretchr/testify
go get -u github.com/DATA-DOG/go-sqlmock
go mod tidy
```

### Review Coverage
```bash
# Generate dan buka HTML report
make test-coverage-html

# Fokus pada coverage untuk packages penting
go test -cover ./internal/service/...
```

### Performance Monitoring
```bash
# Jalankan benchmark secara berkala
make test-bench

# Compare benchmark results
go test -bench=. -benchmem ./test/unit/... > new.txt
benchcmp old.txt new.txt
```
