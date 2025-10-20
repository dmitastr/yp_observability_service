# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

## Оптимизация проекта
```Type: inuse_space
Time: 2025-10-21 01:20:27 MSK
Showing nodes accounting for -2447.51kB, 19.18% of 12760.06kB total
Dropped 3 nodes (cum <= 63.80kB)
      flat  flat%   sum%        cum   cum%
    2565kB 20.10% 20.10%  2052.44kB 16.08%  runtime.allocm
   -1028kB  8.06% 12.05%    -1028kB  8.06%  bufio.NewReaderSize (inline)
    1028kB  8.06% 20.10%     1028kB  8.06%  bufio.NewWriterSize (inline)
-1024.28kB  8.03% 12.07% -1024.28kB  8.03%  github.com/jackc/pgx/v5.(*Conn).getRows
 -902.59kB  7.07%  5.00%  -898.41kB  7.04%  compress/flate.NewWriter (inline)
 -521.05kB  4.08%  0.92%  -521.05kB  4.08%  github.com/lib/pq.map.init.0
 -516.01kB  4.04%  3.13%  -516.01kB  4.04%  github.com/jackc/pgx/v5/internal/iobufpool.init.0.func1
 -512.56kB  4.02%  7.14%  -512.56kB  4.02%  runtime.makeProfStackFP (inline)
  512.22kB  4.01%  3.13%   512.22kB  4.01%  runtime.malg
 -512.08kB  4.01%  7.14%  -512.08kB  4.01%  regexp.compile
 -512.07kB  4.01% 11.16%  -512.07kB  4.01%  net/http.(*Server).newConn (inline)
 -512.05kB  4.01% 15.17%  -512.05kB  4.01%  github.com/golang-migrate/migrate/v4.(*Migrate).lock.func2
 -512.05kB  4.01% 19.18%  -512.05kB  4.01%  golang.org/x/sync/semaphore.(*Weighted).Acquire
```
