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

## Что это

Это Android-хост для локальной игры.
Телефон поднимает встроенный сервер, а сама игра открывается как веб-интерфейс во встроенном окне или в обычном браузере.

Подходит для игры рядом:
- одна Wi-Fi сеть;
- либо точка доступа с телефона;
- отдельный сервер или ПК для запуска комнаты не нужен.

## Как использовать

1. Установите APK из `Releases`.
2. Откройте приложение и запустите сервер.
3. Скопируйте ссылку комнаты.
4. Откройте её на других устройствах.
5. Создайте лобби и начинайте игру.

## Что есть в Android-версии

- запуск локальной комнаты прямо на телефоне;
- встроенный Go-сервер внутри приложения;
- открытие игры во встроенном WebView и в браузере;
- локальные игровые ресурсы внутри APK;
- несколько APK под разные ABI в релизах;
- базовая автоматическая сборка релизов через GitHub Actions.

## Состояние проекта

Android-ветка постепенно догоняет основную игру.

Что важно понимать:
- это отдельная ветка-хост, а не полный отдельный форк игры по архитектуре;
- часть изменений из основного проекта уже перенесена;
- часть ещё переносится;
- возможны баги в локализации, отображении карт и отдельных элементах интерфейса.

## Релизы

В релизах публикуются несколько APK:
- `arm64`
- `armv7`
- `x86`
- `x8664`
- `universal`

Если не нужен универсальный APK, лучше ставить вариант под своё устройство.

## Для разработки

Основные команды:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/check-encoding.ps1
npm --prefix shared run build
npm --prefix client run build
npm --prefix scenarios test
cd server-go; go test ./...
cd android-app; .\gradlew.bat :app:assembleDebug
```

Главные директории:
- `android-app/` — Android shell;
- `server-go/` — локальный backend;
- `client/` — браузерный UI;
- `shared/` — общие контракты;
- `scenarios/` — сценарии и тесты;
- `locales/` — локализация.
