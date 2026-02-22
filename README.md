# Protocol: Bunker Host (Android)

Android-приложение, которое запускает локальную комнату для игры **«Протокол: Бункер»** прямо на телефоне.

Основной проект игры: [protocol-bunker](https://github.com/FHRha/protocol-bunker)

## Скачать

Все ссылки ведут на **GitHub Releases -> Latest**. На странице релиза выберите нужный файл:

| Платформа | Файлы в релизе |
|---|---|
| Android | [Все необходимые Android-релизы (armv7, arm64 и т.д.)](https://github.com/FHRha/protocol-bunker-android/releases) |
| macOS | В разработке |
| Windows | [Setup x64](https://github.com/FHRha/protocol-bunker/releases/latest) · [EXE x64 (zip)](https://github.com/FHRha/protocol-bunker/releases/latest) · [Portable x64](https://github.com/FHRha/protocol-bunker/releases/latest) |
| Linux x64 | [Public](https://github.com/FHRha/protocol-bunker/releases/latest) · [Server](https://github.com/FHRha/protocol-bunker/releases/latest) |
| Linux ARM64 | [Public](https://github.com/FHRha/protocol-bunker/releases/latest) · [Server](https://github.com/FHRha/protocol-bunker/releases/latest) |

> Если файлов на странице релиза пока нет — значит сборки еще не залиты (или GitHub решил устроить нам квест).

## Что это за приложение

- Телефон становится хостом комнаты.
- Друзья подключаются по ссылке.
- Хост тоже может играть с этого же телефона.
- Работает только локально: одна Wi-Fi сеть или точка доступа.

## Что уже есть

- Старт/стоп сервера в Android-приложении.
- Ссылка для подключения и кнопка копирования.
- Открытие игры во встроенном окне и в браузере.
- Логи сервера прямо в приложении.
- Фоновая работа через уведомление.
- Режим разработки для проверки механик.

## Быстрый запуск

1. Установите APK из раздела Releases.
2. Запустите приложение и нажмите старт.
3. Скопируйте ссылку комнаты и отправьте друзьям.
4. Играйте.

## Для разработчиков

### Требования

- Go 1.25+
- JDK 17
- Android SDK (`platform-tools`, `platforms;android-35`, `build-tools;35.0.0`)
- Android NDK (`ndk;27.2.12479018`)

### Локальная сборка

```bash
# 1) Go
cd server-go
go test ./...
go build ./...

# 2) Go-бинарники под Android ABI
# Windows:
.\scripts\build-android-binaries.ps1
# Linux/macOS:
./scripts/build-android-binaries.sh

# 3) APK
cd ../android-app
./gradlew assembleDebug assembleRelease
```

## Автосборка релизов

Workflow: `.github/workflows/release.yml`

После публикации релиза GitHub Actions автоматически:

- прогоняет `go test ./...`;
- собирает Android APK;
- прикладывает 5 файлов в релиз:
  - `arm64`
  - `armv7`
  - `x86`
  - `x8664`
  - `universal`

## Production-подпись APK

В `Settings -> Secrets and variables -> Actions` добавьте:

- `ANDROID_KEYSTORE_BASE64`
- `ANDROID_STORE_PASSWORD`
- `ANDROID_KEY_ALIAS`
- `ANDROID_KEY_PASSWORD`

Если эти secrets не заданы, релиз соберется с debug-подписью (только для тестов).
