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
**Что проверил**
- профиль работы программы, какое время занимают различные блоки - однозначных выводов сделать не удалось
- выделение памяти и объектов - много места занимало сжатие данных, хотя не во всех ответах это имеет смысл 

**Что было сделано**
- разделил мидлварь сжатия на 2 - одна используется для компрессии и декомпрессии, другая - только для декомпрессии запроса
- исправил запись хедеров для корректной работы новой схемы

**TODO**
- использовать более лёгкую библитеку для обработки json, например, easyjson
- при записи данных в файл (например в классе Auditor) не создавать промежуточный массив, использовать json.Encode 

```Type: inuse_space
Type: inuse_space
Time: 2025-10-21 01:20:27 MSK
Showing nodes accounting for -6607.32kB, 51.78% of 12760.06kB total
      flat  flat%   sum%        cum   cum%
-4512.93kB 35.37% 35.37% -5057.60kB 39.64%  compress/flate.NewWriter (inline)
-1542.01kB 12.08% 47.45% -1542.01kB 12.08%  bufio.NewReaderSize (inline)
    1539kB 12.06% 35.39%  1026.44kB  8.04%  runtime.allocm
    1027kB  8.05% 27.34%     1027kB  8.05%  bufio.NewWriterSize (inline)
 -544.67kB  4.27% 31.61%  -544.67kB  4.27%  compress/flate.(*compressor).initDeflate (inline)
 -521.05kB  4.08% 35.69%  -521.05kB  4.08%  github.com/lib/pq.map.init.0
 -516.01kB  4.04% 39.74%  -516.01kB  4.04%  github.com/jackc/pgx/v5/internal/iobufpool.init.0.func1
 -512.56kB  4.02% 43.76%  -512.56kB  4.02%  runtime.makeProfStackFP (inline)
  512.14kB  4.01% 39.74%  -512.05kB  4.01%  github.com/dmitastr/yp_observability_service/internal/repository/postgres_storage.(*Postgres).Get
 -512.14kB  4.01% 43.76%  -512.14kB  4.01%  github.com/jackc/pgx/v5.(*Conn).getRows
  512.11kB  4.01% 39.74%  1025.11kB  8.03%  net/http.(*conn).readRequest
 -512.08kB  4.01% 43.76%  -512.08kB  4.01%  regexp.compile
 -512.07kB  4.01% 47.77%  -512.07kB  4.01%  net/http.(*Server).newConn (inline)
 -512.05kB  4.01% 51.78%  -512.05kB  4.01%  github.com/golang-migrate/migrate/v4.(*Migrate).lock.func2
 -512.05kB  4.01% 55.79%  -512.05kB  4.01%  golang.org/x/sync/semaphore.(*Weighted).Acquire
  512.04kB  4.01% 51.78%   512.04kB  4.01%  github.com/golang-migrate/migrate/v4/source/file.init.0

```
