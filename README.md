# Protocol_Bunker-android-host

Отдельный Android-host проект для локальной игры Protocol: Bunker.

## Важно: режим сети

Проект рассчитан только на локальную компанию рядом:
- одна Wi-Fi сеть, или
- hotspot (точка доступа) телефона-хоста.

WAN/внешний IP в этом проекте не используется.

## Структура

- `android-app/` - Android приложение (Kotlin, ForegroundService).
- `server-go/` - локальный Go сервер игры.
- `shared-contract/` - зафиксированные API-контракты и примеры JSON.
- `client/`, `shared/`, `scenarios/`, `assets/` - материалы основной игры.

## Что уже работает

- Android UI управления сервером: порт, старт/стоп, статус, LAN URL, логи.
- Запуск через `ForegroundService` + постоянная нотификация.
- Открытие клиента в браузере и во встроенном WebView.
- Реальный `client/dist` обслуживается из `server-go` (SPA fallback).
- Реализованы HTTP/WS API для игрового флоу (лобби -> матч -> завершение).
- Режим стримера/overlay удалён из Android-host сборки.
- `DEV_MODE` запускает реальный Go-сервер в режиме `dev_tab` + включает dev-сценарии.
- Mock backend сохранён только как аварийный fallback (если Go-бинарь не запускается/недоступен).
- ABI-сборка Go бинаря и упаковка в assets:
  - `arm64-v8a/server-go`
  - `armeabi-v7a/server-go`
  - `x86_64/server-go`
  - `x86/server-go`
- CI-скрипты: `go test`, ABI build, `assembleDebug`, `assembleRelease`, опциональные device smoke.

## Требования для локальной сборки

- Go 1.25+
- JDK 17
- Android SDK (platform-tools, platforms;android-35, build-tools;35.0.0)
- Android NDK (рекомендуется `ndk;27.2.12479018`) для ABI-сборки `server-go`

Можно использовать SDK внутри репозитория: `android-app/.android-sdk`.

## Release-конфиг Android

Шаблон: `android-app/keystore.properties.example`

Скопируйте в `android-app/keystore.properties` и заполните:
- `APP_ID`
- `VERSION_CODE`
- `VERSION_NAME`
- `RELEASE_STORE_FILE`
- `RELEASE_STORE_PASSWORD`
- `RELEASE_KEY_ALIAS`
- `RELEASE_KEY_PASSWORD`

Если release-keystore не задан, `assembleRelease` подписывается debug-ключом (только для теста).

## Сборка

### 1) Go сервер

```bash
cd server-go
go test ./...
go build ./...
```

### 2) Android ABI-бинарники

Windows:

```powershell
cd server-go
.\scripts\build-android-binaries.ps1
```

Linux/macOS:

```bash
cd server-go
./scripts/build-android-binaries.sh
```

### 3) Android APK

```powershell
cd android-app
.\gradlew.bat assembleDebug assembleRelease
```

## CI-команды

Linux/macOS:

```bash
./scripts/ci.sh
```

Windows:

```powershell
.\scripts\ci.ps1
```

Скрипты автоматически:
- запускают `go test ./...`,
- собирают ABI-бинарники,
- запускают `assembleDebug` и `assembleRelease`,
- при включении device-режима запускают `connectedDebugAndroidTest` и smoke e2e.

Device-режим:
- Linux/macOS: `RUN_DEVICE_TESTS=1 ./scripts/ci.sh`
- Windows: `.\scripts\ci.ps1 -RunDeviceTests`

## GitHub Releases (автосборка APK)

В репозитории настроен workflow: `.github/workflows/release.yml`.

Как выпускать версию:

1. Создайте tag в формате `vX.Y.Z` (например, `v0.1.0`).
2. В GitHub откройте `Releases` -> `Draft a new release`.
3. Выберите этот tag, заполните заголовок и текст релиза (Markdown), при необходимости отметьте `Set as a pre-release`.
4. Нажмите `Publish release`.

После публикации GitHub Actions автоматически:
- запускает `go test ./...`,
- пересобирает Go ABI-бинарники для `arm64-v8a`, `armeabi-v7a`, `x86_64`, `x86`,
- собирает `assembleRelease` с ABI splits,
- прикрепляет в релиз 5 APK:
  - `arm64-v8a`
  - `armeabi-v7a`
  - `x86_64`
  - `x86`
  - `universal`

Важно:
- workflow ожидает минимум 5 APK и завершится ошибкой, если их меньше;
- если не настроены секреты keystore, release будет debug-signed (только для теста).

Для production-подписи задайте secrets в GitHub:
- `ANDROID_KEYSTORE_BASE64`
- `ANDROID_STORE_PASSWORD`
- `ANDROID_KEY_ALIAS`
- `ANDROID_KEY_PASSWORD`

### Как писать текст релиза (Release notes)

Проще всего использовать такой шаблон в окне GitHub Release:

```md
## Что нового
- ...

## Исправления
- ...

## APK файлы
- arm64-v8a: большинство современных Android
- armeabi-v7a: старые ARM-устройства
- x86_64: эмуляторы/редкие x86_64 устройства
- x86: старые x86 эмуляторы
- universal: универсальный APK (больше размер)

## Примечания
- Статус: pre-release/release
- Известные ограничения: ...
```

## Тесты

### Go

```bash
cd server-go
go test ./...
```

### Android instrumented

```powershell
cd android-app
.\gradlew.bat connectedDebugAndroidTest
```

### Smoke e2e через ADB

Windows:

```powershell
cd android-app
.\scripts\smoke-e2e.ps1
```

Linux/macOS:

```bash
cd android-app
./scripts/smoke-e2e.sh
```
