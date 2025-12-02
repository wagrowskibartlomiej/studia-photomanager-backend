# Photo Manager Backend

Backend API dla aplikacji zarządzania zdjęciami napisany w Go. Aplikacja umożliwia użytkownikom rejestrację, logowanie, przesyłanie zdjęć oraz zarządzanie ich widocznością. Administratorzy mogą zarządzać użytkownikami i banować ich.

## 📋 Funkcje

- 🔐 **Autentykacja użytkowników** - Rejestracja i logowanie z JWT
- 📸 **Zarządzanie zdjęciami** - Przesyłanie, usuwanie i przeglądanie zdjęć
- 🌐 **Publiczna galeria** - Udostępnianie zdjęć publicznie
- 👥 **Zarządzanie użytkownikami** - Panel administracyjny
- 🚫 **System banowania** - Administracja może banować użytkowników
- 🔒 **Elastyczna walidacja hasła** - Konfigurowalne poziomy bezpieczeństwa (no-validation, easy, medium, restrict, custom)

## 🛠️ Wymagania

- Go 1.25.4 lub nowszy
- SQLite3

## 📦 Instalacja

1. Sklonuj repozytorium:
```bash
git clone <repository-url>
cd backend
```

2. Zainstaluj zależności:
```bash
go mod download
```

3. Skonfiguruj aplikację:
   - Edytuj plik `config.json` zgodnie z Twoimi potrzebami
   - Domyślna konfiguracja jest już gotowa do użycia

## ⚙️ Konfiguracja

Plik `config.json` zawiera następujące opcje:

```json
{
  "server": {
    "port": "8080",        // Port serwera
    "host": ""             // Host (pusty = wszystkie interfejsy)
  },
  "database": {
    "file": "./photomanager.db"  // Ścieżka do pliku bazy danych
  },
  "jwt": {
    "secret_key": "...",         // Klucz do podpisywania JWT
    "timeout_minutes": 15        // Czas ważności tokenu (w minutach, domyślnie 15)
  },
  "photos": {
    "directory": "photos"        // Katalog na zdjęcia
  },
  "admin": {
    "default_login": "admin",    // Domyślny login administratora
    "default_password": "adminadmin"  // Domyślne hasło administratora
  },
  "password": {
    "mode": "no-validation"      // Tryb walidacji: no-validation, easy, medium, restrict, custom
  }
}
```

### Walidacja hasła

System obsługuje kilka trybów walidacji hasła:

- **no-validation** (domyślny) - Brak walidacji, akceptuje dowolne hasło
- **easy** - Minimum 3 znaki
- **medium** - Minimum 6 znaków, wymaga co najmniej jednej litery i jednej cyfry
- **restrict** - Minimum 8 znaków, wymaga wielkiej litery, małej litery, cyfry i znaku specjalnego
- **custom** - Własna konfiguracja walidacji

#### Przykład konfiguracji custom validatora:

```json
{
  "password": {
    "mode": "custom",
    "custom": {
      "min_length": 8,
      "max_length": 32,
      "require_upper": true,
      "require_lower": true,
      "require_digit": true,
      "require_special": true,
      "regex": "^[A-Za-z0-9!@#$%^&*]+$"
    }
  }
}
```

**Opcje custom validatora:**
- `min_length` - Minimalna długość hasła
- `max_length` - Maksymalna długość hasła
- `require_upper` - Wymaga wielkiej litery
- `require_lower` - Wymaga małej litery
- `require_digit` - Wymaga cyfry
- `require_special` - Wymaga znaku specjalnego
- `regex` - Opcjonalny wzorzec regex do walidacji

## 🚀 Uruchomienie

```bash
go run .
```

Lub zbuduj i uruchom:
```bash
go build
./backend
```

Serwer uruchomi się na porcie określonym w `config.json` (domyślnie `:8080`).

## 🧪 Testy

Uruchom wszystkie testy:
```bash
go test -v
```

Uruchom konkretny test:
```bash
go test -v -run TestValidatePassword
```

## 📡 API Endpoints

### Autentykacja

#### POST `/api/register`
Rejestracja nowego użytkownika.

**Request Body:**
```json
{
  "login": "username",
  "password": "password123"
}
```

**Response:**
```json
{
  "status": "ok"
}
```

#### POST `/api/login`
Logowanie użytkownika.

**Request Body:**
```json
{
  "login": "username",
  "password": "password123"
}
```

**Response:**
```json
{
  "status": "ok",
  "isAdmin": false
}
```

**Cookie:** Ustawia cookie `jwt` z tokenem autentykacji.

### Zdjęcia

#### POST `/api/add-photo`
Przesyłanie zdjęcia (wymaga autentykacji).

**Form Data:**
- `photo`: plik zdjęcia
- `public`: "1" dla publicznego, "0" dla prywatnego

**Response:**
```json
{
  "message": "Photo uploaded"
}
```

#### GET `/api/photos/{username}`
Pobranie listy zdjęć użytkownika.

**Response:**
```json
[
  {
    "filename": "photo.jpg",
    "public": true
  }
]
```

#### GET `/api/photos/{username}/{filename}`
Pobranie konkretnego zdjęcia.

#### DELETE `/api/delete-photo/{username}/{filename}`
Usunięcie zdjęcia (wymaga autentykacji, tylko właściciel).

#### POST `/api/toggle-public`
Zmiana widoczności zdjęcia (wymaga autentykacji).

**Request Body:**
```json
{
  "filename": "photo.jpg",
  "public": 1
}
```

#### GET `/api/public-gallery`
Pobranie listy wszystkich publicznych zdjęć.

**Response:**
```json
[
  {
    "user": "username",
    "filename": "photo.jpg"
  }
]
```

### Administracja

#### GET `/api/users`
Lista użytkowników (wymaga autentykacji administratora).

**Response:**
```json
[
  {
    "login": "username",
    "isBanned": false
  }
]
```

#### POST `/api/manage-ban`
Zarządzanie statusem bana użytkownika (wymaga autentykacji administratora).

**Request Body:**
```json
{
  "login": "username",
  "banned": 1
}
```

**Response:**
```json
{
  "login": "username",
  "banned": "1",
  "message": "Ban status updated"
}
```

## 📁 Struktura projektu

```
backend/
├── main.go              # Punkt wejścia, routing
├── config.go            # Wczytywanie konfiguracji
├── config.json          # Plik konfiguracyjny
├── types.go            # Struktury danych
├── database.go          # Operacje na bazie danych
├── auth.go              # Generowanie i parsowanie JWT
├── middleware.go        # Middleware autentykacji
├── handlers.go         # Handlery HTTP
├── *_test.go           # Pliki testowe
├── go.mod              # Zależności Go
└── README.md           # Ten plik
```

## 🔒 Bezpieczeństwo

- Hasła są hashowane używając bcrypt
- Autentykacja oparta na JWT w cookie
- Konfigurowalna walidacja hasła
- Middleware sprawdzający uprawnienia
- Ochrona przed banowaniem samego siebie przez administratora

## 📝 Uwagi

- Domyślny administrator jest tworzony automatycznie przy pierwszym uruchomieniu
- Zdjęcia są przechowywane lokalnie w katalogu określonym w konfiguracji
- Baza danych SQLite jest tworzona automatycznie
- Tokeny JWT są ważne przez czas określony w konfiguracji
